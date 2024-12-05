package botfunc

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/AVick23/ToDo-Bot/database"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var (
	response  string
	buttons   [][]tgbotapi.InlineKeyboardButton
	userState = make(map[int64]string)
)

func NewInlineKeyboard(text string, command string) tgbotapi.InlineKeyboardButton {
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
			processingCallbackQuery(bot, update, db)
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
				NewInlineKeyboard("Мой день", "day"),
				NewInlineKeyboard("Запланированное", "planned"),
			),
			tgbotapi.NewInlineKeyboardRow(
				NewInlineKeyboard("Задачи", "tasks"),
				NewInlineKeyboard("Создать свой список", "create_list"),
			),
		)
		msg.ReplyMarkup = keyboard
		bot.Send(msg)
	} else if update.Message.Command() == "add" {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Хотите добавить новую задачу?\n Напишите задачу: ")
		bot.Send(msg)
		userState[update.Message.Chat.ID] = "state"
	} else if userState[update.Message.Chat.ID] == "state" {
		task := update.Message.Text
		username := fmt.Sprintf("%v", update.Message.Chat.ID)

		id, err := database.SaveUser(db, username)
		if err != nil {
			fmt.Printf("Не получилось сохранить в базу данных ID пользователя %v", err)
		}

		err = database.SaveTasks(db, id, task)
		if err != nil {
			response := "Кажется произошла ошибка, попробойти ещё раз"
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, response)
			bot.Send(msg)
			fmt.Printf("Произошла ошибка %v", err)
			return
		}

		respponse := fmt.Sprintf("Ваша задача успешно сохранена: (%v)", task)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, respponse)
		bot.Send(msg)
		userState[update.Message.Chat.ID] = ""
	}
}

func processingCallbackQuery(bot *tgbotapi.BotAPI, update tgbotapi.Update, db *sql.DB) {
	if update.CallbackQuery != nil {
		username := fmt.Sprintf("%v", update.CallbackQuery.From.ID)

		// Проверка нажатия на кнопку задачи
		if len(update.CallbackQuery.Data) > 5 && update.CallbackQuery.Data[:5] == "task_" {
			task := update.CallbackQuery.Data[5:]
			handleTaskAction(bot, update, task)
			return
		} else if len(update.CallbackQuery.Data) > 9 && update.CallbackQuery.Data[:9] == "complete_" {
			task := update.CallbackQuery.Data[9:]
			completeTasks(bot, update, username, task, db)
			return
		} else if len(update.CallbackQuery.Data) > 7 && update.CallbackQuery.Data[:7] == "delete_" {
			task := update.CallbackQuery.Data[7:]
			deleteTasks(bot, update, username, task, db)
			return
		}

		tasks, err := database.GetTasks(db, username)
		if err != nil {
			response := "Не удалось получить задания, попробуйте ещё раз"
			log.Printf("Ошибка получения задач: %v", err)
			msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, response)
			bot.Send(msg)
			return
		}

		switch update.CallbackQuery.Data {
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
			msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, response)
			bot.Send(msg)
		}
	}
}

func handleTaskAction(bot *tgbotapi.BotAPI, update tgbotapi.Update, task string) {
	response = fmt.Sprintf("Что вы хотите сделать с задачей: %s?", task)
	buttons := [][]tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardRow(
			NewInlineKeyboard("Выполнил задачу", "complete_"+task),
			NewInlineKeyboard("Удалить задачу", "delete_"+task),
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
