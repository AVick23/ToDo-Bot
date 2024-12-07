package savetask

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/AVick23/ToDo-Bot/database"
	"github.com/AVick23/ToDo-Bot/models"
	"github.com/AVick23/ToDo-Bot/valid"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var (
	userStateTask    = make(map[int64]string)
	inputDescription = make(map[int64]string)
	inputDate        = make(map[int64]string)
	inputTime        = make(map[int64]string)
)

func SaveTaskUser(bot *tgbotapi.BotAPI, update tgbotapi.Update, db *sql.DB) {
	chatID := update.Message.Chat.ID

	if _, ok := userStateTask[chatID]; !ok {
		userStateTask[chatID] = "description"
	}

	switch userStateTask[chatID] {
	case "description":
		handleTaskDescription(bot, update, chatID)
	case "date":
		handleTaskDate(bot, update, chatID)
	case "time":
		handleTaskTime(bot, update, chatID, db)
	}
}

func handleTaskDescription(bot *tgbotapi.BotAPI, update tgbotapi.Update, chatID int64) {
	inputDescription[chatID] = update.Message.Text
	msg := tgbotapi.NewMessage(chatID, "📅 Давай запланируем твою задачу! Введи дату в формате *дд.мм.гггг* или нажми 'Пропустить'. 😊")
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⏩ Пропустить", "skip_date"),
		),
	)
	msg.ReplyMarkup = keyboard
	bot.Send(msg)
	userStateTask[chatID] = "date"
}

func handleTaskDate(bot *tgbotapi.BotAPI, update tgbotapi.Update, chatID int64) {
	input := update.Message.Text
	parseDate, err := time.Parse("02.01.2006", input)
	if err != nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "⚠️ Упс, кажется ты ввел что-то не так. Пожалуйста, используй формат: *дд.мм.гггг*. 😉")
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("⏩ Пропустить", "skip_date"),
			),
		)
		msg.ReplyMarkup = keyboard
		bot.Send(msg)
		return
	}

	today := time.Now().Truncate(24 * time.Hour)

	if parseDate.Before(today) {
		msg := tgbotapi.NewMessage(chatID, "🚫 Дата не может быть в прошлом. Попробуй ещё раз! 😊")
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("⏩ Пропустить", "skip_date"),
			),
		)
		msg.ReplyMarkup = keyboard
		bot.Send(msg)
		return
	}

	inputDate[chatID] = input

	msg := tgbotapi.NewMessage(chatID, "⏰ Отлично! Теперь введи время в формате *чч:мм*, или нажми 'Пропустить'.")
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⏩ Пропустить", "skip_time"),
		),
	)
	msg.ReplyMarkup = keyboard
	bot.Send(msg)

	userStateTask[chatID] = "time"
}

func handleTaskTime(bot *tgbotapi.BotAPI, update tgbotapi.Update, chatID int64, db *sql.DB) {
	input := update.Message.Text
	parseTime, err := time.Parse("15:04", input)

	if err != nil || !valid.IsValidTime(parseTime) {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "⚠️ Хмм, что-то не так. Пожалуйста, введи время в формате *чч:мм*. 🙃")
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("⏩ Пропустить", "skip_time"),
			),
		)
		msg.ReplyMarkup = keyboard
		bot.Send(msg)
		return
	}

	today := time.Now().Format("02.01.2006")
	if inputTime[chatID] == today && parseTime.Before(time.Now()) {
		msg := tgbotapi.NewMessage(chatID, "🚫 Ты не можешь запланировать на прошлое время. Попробуй снова! 😉")
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("⏩ Пропустить", "skip_time"),
			),
		)
		msg.ReplyMarkup = keyboard
		bot.Send(msg)
		return
	}

	inputTime[chatID] = input
	userStateTask[chatID] = ""

	task := models.Task{
		Description: strings.TrimSpace(inputDescription[chatID]),
	}

	if inputDate[chatID] != "" {
		date := inputDate[chatID]
		task.Date = &date
	}
	if inputTime[chatID] != "" {
		time := inputTime[chatID]
		task.Time = &time
	}

	username := fmt.Sprintf("%v", chatID)

	id, err := database.SaveUser(db, username)
	if err != nil {
		fmt.Printf("😟 Не удалось сохранить пользователя в базу данных: %v", err)
	}

	err = database.SaveTasks(db, id, task)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "😔 Что-то пошло не так. Попробуй ещё раз!")
		bot.Send(msg)
		fmt.Printf("Ошибка сохранения задачи: %v", err)
		return
	}

	response := "🎉 Задача успешно сохранена!\n✍️ *" + task.Description + "*"
	if task.Date != nil {
		response += "\n🗓️ Дата: *" + *task.Date + "*"
	}
	if task.Time != nil {
		response += "\n🕰️ Время: *" + *task.Time + "*"
	}
	msg := tgbotapi.NewMessage(chatID, response)
	bot.Send(msg)

	clearTaskState(chatID)
}

func SaveTaskDate(bot *tgbotapi.BotAPI, chatID int64) {
	if userStateTask[chatID] != "date" {
		return
	}
	inputDate[chatID] = ""
	msg := tgbotapi.NewMessage(chatID, "⏰ Хорошо! Теперь введи время в формате *чч:мм*, или нажми 'Пропустить'.")
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⏩ Пропустить", "skip_time"),
		),
	)
	msg.ReplyMarkup = keyboard
	bot.Send(msg)
	userStateTask[chatID] = "time"
}

func SaveTaskTime(bot *tgbotapi.BotAPI, chatID int64, db *sql.DB) {
	if userStateTask[chatID] != "time" {
		return
	}

	inputTime[chatID] = ""

	userStateTask[chatID] = ""

	task := models.Task{
		Description: strings.TrimSpace(inputDescription[chatID]),
	}

	if inputDate[chatID] != "" {
		date := inputDate[chatID]
		task.Date = &date
	}
	if inputTime[chatID] != "" {
		time := inputTime[chatID]
		task.Time = &time
	}

	username := fmt.Sprintf("%v", chatID)
	id, err := database.SaveUser(db, username)
	if err != nil {
		fmt.Printf("😟 Не удалось сохранить пользователя в базу данных: %v", err)
		return
	}

	err = database.SaveTasks(db, id, task)
	if err != nil {
		response := "😔 Что-то пошло не так. Попробуй ещё раз!"
		bot.Send(tgbotapi.NewMessage(chatID, response))
		fmt.Printf("Ошибка сохранения задачи: %v", err)
		return
	}

	response := "🎉 Задача успешно сохранена!\n✍️ *" + task.Description + "*"
	if task.Date != nil {
		response += fmt.Sprintf("\n🗓️ Дата: *%s*", *task.Date)
	}
	if task.Time != nil {
		response += fmt.Sprintf("\n🕰️ Время: *%s*", *task.Time)
	}
	bot.Send(tgbotapi.NewMessage(chatID, response))

	clearTaskState(chatID)
}

func clearTaskState(chatID int64) {
	delete(inputDescription, chatID)
	delete(inputDate, chatID)
	delete(inputTime, chatID)
	delete(userStateTask, chatID)
}
