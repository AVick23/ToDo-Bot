package main

import (
	"fmt"
	"log"

	"github.com/AVick23/ToDo-Bot/database"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func NewInlineKeyboard(text string, command string) tgbotapi.InlineKeyboardButton {
	return tgbotapi.NewInlineKeyboardButtonData(text, command)
}

var userState = make(map[int64]string)

func main() {
	bot, err := tgbotapi.NewBotAPI("7699728760:AAGsMWGdlQsyI0q7dxR5by1pJaHBApj45_k")
	if err != nil {
		log.Fatalf("Не получилось подключиться к API, проверьте ошибку: %v", err)
	}

	bot.Debug = true

	log.Printf("Подключение к боту %v прошла успешна", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	db, err := database.ConnectDB()
	if err != nil {
		log.Fatalf("Не получилось подключиться к БД %v", err)
	}
	defer db.Close()

	fmt.Println("Всё работает, ежи братка.")

	for update := range updates {
		if update.Message != nil {
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
		} else if update.CallbackQuery != nil {
			var response string
			var buttons [][]tgbotapi.InlineKeyboardButton

			username := fmt.Sprintf("%v", update.CallbackQuery.From.ID)

			tasks, err := database.GetTasks(db, username)
			if err != nil {
				response = "Не удалось получить задания, попробуйте ещё раз"
				log.Printf("Ошибка получения задач: %v", err)
			} else {
				switch update.CallbackQuery.Data {
				case "day":
					response = "Это ваши задачи на сегодня:\n"
				case "planned":
					response = "Это ваши запланированные задачи"
				case "tasks":
					response = "Вот список ваших задач"
					for _, task := range tasks {
						buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(
							tgbotapi.NewInlineKeyboardButtonData(task, task),
						))
					}
				case "create_list":
					response = "Введите название нового списка"
				}

				if len(buttons) > 0 {
					replyMarkup := tgbotapi.NewInlineKeyboardMarkup(buttons...)
					msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, response)
					msg.ReplyMarkup = replyMarkup
					bot.Send(msg)
				} else {
					msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, response)
					bot.Send(msg)
				}
			}
		}
	}
}
