package savetask

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/AVick23/ToDo-Bot/database"
	"github.com/AVick23/ToDo-Bot/models"
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
	msg := tgbotapi.NewMessage(chatID, "Введите дату в формате дд.мм.гггг или нажмите 'Пропустить'.")
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Пропустить", "skip_date"),
		),
	)
	msg.ReplyMarkup = keyboard
	bot.Send(msg)
	userStateTask[chatID] = "date"
}

func handleTaskDate(bot *tgbotapi.BotAPI, update tgbotapi.Update, chatID int64) {
	inputDate[chatID] = update.Message.Text
	msg := tgbotapi.NewMessage(chatID, "А теперь можете ввести время в формате чч:мм")
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Пропустить", "skip_time"),
		),
	)
	msg.ReplyMarkup = keyboard
	bot.Send(msg)
	userStateTask[chatID] = "time"
}

func handleTaskTime(bot *tgbotapi.BotAPI, update tgbotapi.Update, chatID int64, db *sql.DB) {
	inputTime[chatID] = update.Message.Text
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
		fmt.Printf("Не получилось сохранить в базу данных ID пользователя %v", err)
	}

	err = database.SaveTasks(db, id, task)
	if err != nil {
		response := "Кажется произошла ошибка, попробуйте ещё раз"
		msg := tgbotapi.NewMessage(chatID, response)
		bot.Send(msg)
		fmt.Printf("Произошла ошибка %v", err)
		return
	}

	response := "Ваша задача успешно сохранена: \n✍️" + task.Description
	if task.Date != nil {
		response += "\n🗓️" + *task.Date
	}
	if task.Time != nil {
		response += "\n🕰️" + *task.Time
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
	msg := tgbotapi.NewMessage(chatID, "А теперь можете ввести время в формате чч:мм")
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Пропустить", "skip_time"),
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
		fmt.Printf("Не получилось сохранить в базу данных ID пользователя %v", err)
	}

	err = database.SaveTasks(db, id, task)
	if err != nil {
		response := "Кажется произошла ошибка, попробуйте ещё раз"
		bot.Send(tgbotapi.NewMessage(chatID, response))
		fmt.Printf("Произошла ошибка %v", err)
		return
	}

	response := fmt.Sprintf("Ваша задача успешно сохранена:\n✍️ %s", task.Description)
	if task.Date != nil {
		response += fmt.Sprintf("\n🗓️ %s", *task.Date)
	}
	if task.Time != nil {
		response += fmt.Sprintf("\n🕰️ %s", *task.Time)
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
