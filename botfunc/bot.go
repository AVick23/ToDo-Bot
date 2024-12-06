package botfunc

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	"github.com/AVick23/ToDo-Bot/database"
	"github.com/AVick23/ToDo-Bot/models"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var (
	response         string
	buttons          [][]tgbotapi.InlineKeyboardButton
	userState        = make(map[int64]string)
	userStateTask    = make(map[int64]string)
	inputDescription = make(map[int64]string)
	inputDate        = make(map[int64]string)
	inputTime        = make(map[int64]string)
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
		saveTaskUser(bot, update, db)
	}
}

func saveTaskUser(bot *tgbotapi.BotAPI, update tgbotapi.Update, db *sql.DB) {
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
	saveTaskDate(chatID, update)
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
	saveTaskTime(chatID, update)
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
}

func saveTaskDate(chatID int64, update tgbotapi.Update) {
	if update.CallbackQuery != nil && update.CallbackQuery.Data == "skip_date" {
		inputDate[chatID] = ""
	}
	userStateTask[chatID] = "time"
}

func saveTaskTime(chatID int64, update tgbotapi.Update) {
	if update.CallbackQuery != nil && update.CallbackQuery.Data == "skip_time" {
		inputTime[chatID] = ""
	}
}

func processCallbackQuery(bot *tgbotapi.BotAPI, update tgbotapi.Update, db *sql.DB) {
	if update.CallbackQuery == nil {
		return
	}

	username := fmt.Sprintf("%v", update.CallbackQuery.From.ID)
	callbackData := update.CallbackQuery.Data

	switch {
	case strings.HasPrefix(callbackData, "task_"):
		handleTaskAction(bot, update, callbackData[5:])
	case strings.HasPrefix(callbackData, "complete_"):
		completeTasks(bot, update, username, callbackData[9:], db)
	case strings.HasPrefix(callbackData, "delete_"):
		deleteTasks(bot, update, username, callbackData[7:], db)
	default:
		handleDefaultCallback(bot, update, db, username, callbackData)
	}
}

func handleDefaultCallback(bot *tgbotapi.BotAPI, update tgbotapi.Update, db *sql.DB, username string, callbackData string) {
	tasks, err := database.GetTasks(db, username)
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
	case "planned":
		response = "Это ваши запланированные задачи"
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
		msg := tgbotapi.NewEditMessageReplyMarkup(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, replyMarkup)
		bot.Send(msg)
	} else {
		replyMarkup := tgbotapi.NewInlineKeyboardMarkup([]tgbotapi.InlineKeyboardButton{})
		msg := tgbotapi.NewEditMessageReplyMarkup(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, replyMarkup)
		bot.Send(msg)

		noTask := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "К сожалению список ваших задач сейчас пуст")
		bot.Send(noTask)
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

func completeTasks(bot *tgbotapi.BotAPI, update tgbotapi.Update, username string, task string, db *sql.DB) {
	err := database.CompleteTasksDB(db, username, task)
	if err != nil {
		response := "Кажется произошла ошибка, попробуйте ещё раз"
		msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, response)
		bot.Send(msg)
		log.Printf("Произошла ошибка: %v", err)
		return
	}

	response := fmt.Sprintf("Ваша задача успешно выполнена и сохранена: (%v)", task)
	msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, response)
	bot.Send(msg)
}

func deleteTasks(bot *tgbotapi.BotAPI, update tgbotapi.Update, username string, task string, db *sql.DB) {
	err := database.DeleteTaskSQL(db, task, username)
	if err != nil {
		response := "Кажется произошла ошибка, попробуйте ещё раз"
		msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, response)
		bot.Send(msg)
		log.Printf("Произошла ошибка: %v", err)
		return
	}

	response := fmt.Sprintf("Ваша задача успешно удалена: (%v)", task)
	msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, response)
	bot.Send(msg)
}
