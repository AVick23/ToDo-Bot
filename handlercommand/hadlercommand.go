package handlercommand

import (
	"database/sql"
	"log"

	"github.com/AVick23/ToDo-Bot/database"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var (
	response string
	buttons  [][]tgbotapi.InlineKeyboardButton
)

func HandleDefaultCallback(bot *tgbotapi.BotAPI, update tgbotapi.Update, db *sql.DB, username string, callbackData string) {
	tasks, err := database.GetTasks(db, username)
	if err != nil {
		log.Printf("Ошибка получения задач: %v", err)
		msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Не удалось получить задания, попробуйте ещё раз")
		bot.Send(msg)
		return
	}

	tasksplan, err := database.GetTasksPlanned(db, username)
	if err != nil {
		log.Printf("Ошибка получения задач: %v", err)
		msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Не удалось получить задания, попробуйте ещё раз")
		bot.Send(msg)
		return
	}

	taskday, err := database.GetTasksDay(db, username)
	if err != nil {
		log.Printf("Ошибка получения задач: %v", err)
		msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Не удалось получить задания, попробуйте ещё раз")
		bot.Send(msg)
		return
	}

	buttons = [][]tgbotapi.InlineKeyboardButton{}

	switch callbackData {
	case "day":
		response = "Это ваши задачи на сегодня:\n"
		for _, task := range taskday {
			buttonText := task.Task + " (до " + task.Date + ")"
			buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(buttonText, "task_"+task.Task),
			))
		}
	case "planned":
		response = "Это ваши запланированные задачи:\n"
		for _, task := range tasksplan {
			buttonText := task.Task + " (до " + task.Date + ")"
			if task.Notification.Valid {
				buttonText += " - " + task.Notification.String
			}
			buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(buttonText, "task_"+task.Task),
			))
		}
	case "tasks":
		response = "Вот список ваших задач"
		for _, task := range tasks {
			buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(task, "task_"+task),
			))
		}
	case "create_list":
		response = "Введите название нового списка"
	}

	if len(buttons) > 0 {
		replyMarkup := tgbotapi.NewInlineKeyboardMarkup(buttons...)
		msg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, response)
		msg.ReplyMarkup = &replyMarkup
		bot.Send(msg)
	} else {
		msg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, "К сожалению, список ваших задач сейчас пуст")
		bot.Send(msg)
	}

}
