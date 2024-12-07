package handlercommand

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/AVick23/ToDo-Bot/database"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var (
	response  string
	buttons   [][]tgbotapi.InlineKeyboardButton
	emojiList = []string{"😊", "🚀", "✨", "🌟", "💡", "✅", "📅", "📌", "🕒", "🎯"}
)

func HandleDefaultCallback(bot *tgbotapi.BotAPI, update tgbotapi.Update, db *sql.DB, username string, callbackData string) {

	tasks, err := database.GetTasks(db, username)
	if err != nil {
		log.Printf("Ошибка получения задач: %v", err)
		msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "❌ Произошла ошибка при загрузке задач. Пожалуйста, попробуйте ещё раз.")
		bot.Send(msg)
		return
	}

	tasksplan, err := database.GetTasksPlanned(db, username)
	if err != nil {
		log.Printf("Ошибка получения запланированных задач: %v", err)
		msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "❌ Не удалось загрузить запланированные задачи. Попробуйте позже.")
		bot.Send(msg)
		return
	}

	taskday, err := database.GetTasksDay(db, username)
	if err != nil {
		log.Printf("Ошибка получения задач на сегодня: %v", err)
		msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "❌ Не удалось загрузить задачи на сегодня. Попробуйте позже.")
		bot.Send(msg)
		return
	}

	buttons = [][]tgbotapi.InlineKeyboardButton{}

	switch callbackData {
	case "day":
		response = "📅 *Ваши задачи на сегодня:*\n"
		for _, task := range taskday {
			buttonText := fmt.Sprintf("%s %s (до %s)", getRandomEmoji(), task.Task, task.Date)
			buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(buttonText, "task_"+task.Task),
			))
		}
	case "planned":
		response = "🗓️ *Ваши запланированные задачи:*\n"
		for _, task := range tasksplan {
			buttonText := fmt.Sprintf("%s %s (до %s)", getRandomEmoji(), task.Task, task.Date)
			if task.Notification.Valid {
				buttonText += fmt.Sprintf(" 🔔 %s", task.Notification.String)
			}
			buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(buttonText, "task_"+task.Task),
			))
		}
	case "tasks":
		response = "📋 *Вот ваш список задач:*\n"
		for _, task := range tasks {
			buttonText := fmt.Sprintf("%s %s", getRandomEmoji(), task)
			buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(buttonText, "task_"+task),
			))
		}
	case "create_list":
		response = "✏️ Пожалуйста, введите название для нового списка задач:"
	}

	if len(buttons) > 0 {
		replyMarkup := tgbotapi.NewInlineKeyboardMarkup(buttons...)
		msg := tgbotapi.NewEditMessageTextAndMarkup(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, response, replyMarkup)
		bot.Send(msg)
	} else {
		msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "ℹ️ К сожалению, у вас пока нет задач.")
		bot.Send(msg)
	}
}

func getRandomEmoji() string {
	rand.Seed(time.Now().UnixNano())
	return emojiList[rand.Intn(len(emojiList))]
}
