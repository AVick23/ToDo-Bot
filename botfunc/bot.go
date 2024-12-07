package botfunc

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/AVick23/ToDo-Bot/complete"
	"github.com/AVick23/ToDo-Bot/deletetask"
	"github.com/AVick23/ToDo-Bot/handlercommand"
	"github.com/AVick23/ToDo-Bot/savetask"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var (
	response  string
	userState = make(map[int64]string)
)

func newInlineKeyboard(text string, command string) tgbotapi.InlineKeyboardButton {
	return tgbotapi.NewInlineKeyboardButtonData(text, command)
}

func NewBot(token string) (*tgbotapi.BotAPI, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("не удалось установить подключение к боту: %v", err)
	}
	bot.Debug = true
	fmt.Printf("Удалось успешно подключиться к боту: %v", bot.Self.UserName)
	return bot, nil
}

func RunProcess(bot *tgbotapi.BotAPI, updates tgbotapi.UpdatesChannel, db *sql.DB) {
	for update := range updates {
		if update.Message != nil {
			processingMessage(bot, update, db)

		} else if update.CallbackQuery != nil {
			processCallbackQuery(bot, update, db)
		}
	}
}

func processingMessage(bot *tgbotapi.BotAPI, update tgbotapi.Update, db *sql.DB) {
	if update.Message.Command() == "start" {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Я ваш личный помощник, который будет записывать все ваши задачи. Чтобы узнать, какие команды доступны, воспользуйтесь командой «/help» или просто начните вводить символ «/». Это откроет меню команд, расположенное слева от поля ввода.")
		bot.Send(msg)
	} else if update.Message.Command() == "tasks" {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Выберите опцию:")
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				newInlineKeyboard("Мой день", "day"),
				newInlineKeyboard("Запланированное", "planned"),
			),
			tgbotapi.NewInlineKeyboardRow(
				newInlineKeyboard("Задачи", "tasks"),
				newInlineKeyboard("Создать свой список", "create_list"),
			),
		)
		msg.ReplyMarkup = keyboard
		bot.Send(msg)
	} else if update.Message.Command() == "add" {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Хотите добавить новую задачу?\n Напишите задачу: ")
		bot.Send(msg)
		userState[update.Message.Chat.ID] = "state"
	} else if userState[update.Message.Chat.ID] == "state" {
		savetask.SaveTaskUser(bot, update, db)
	}
}

func processCallbackQuery(bot *tgbotapi.BotAPI, update tgbotapi.Update, db *sql.DB) {
	if update.CallbackQuery == nil {
		return
	}

	chatID := update.CallbackQuery.Message.Chat.ID
	username := fmt.Sprintf("%v", update.CallbackQuery.From.ID)
	callbackData := update.CallbackQuery.Data

	switch {
	case strings.HasPrefix(callbackData, "skip_date"):
		savetask.SaveTaskDate(bot, chatID)
	case strings.HasPrefix(callbackData, "skip_time"):
		savetask.SaveTaskTime(bot, chatID, db)
	case strings.HasPrefix(callbackData, "task_"):
		handleTaskAction(bot, update, callbackData[5:])
	case strings.HasPrefix(callbackData, "complete_"):
		complete.CompleteTasks(bot, update, username, callbackData[9:], db)
	case strings.HasPrefix(callbackData, "delete_"):
		deletetask.DeleteTasks(bot, update, username, callbackData[7:], db)
	default:
		handlercommand.HandleDefaultCallback(bot, update, db, username, callbackData)
	}
}

func handleTaskAction(bot *tgbotapi.BotAPI, update tgbotapi.Update, task string) {
	response = fmt.Sprintf("Что вы хотите сделать с задачей: %s?", task)
	buttons := [][]tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardRow(
			newInlineKeyboard("Выполнил задачу", "complete_"+task),
			newInlineKeyboard("Удалить задачу", "delete_"+task),
		),
	}
	replyMarkup := tgbotapi.NewInlineKeyboardMarkup(buttons...)
	msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, response)
	msg.ReplyMarkup = replyMarkup
	bot.Send(msg)
}
