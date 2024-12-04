package main

import (
	"fmt"
	"log"

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

	for update := range updates {
		if update.Message != nil {
			if update.Message.Command() == "start" {
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
				userId := update.Message.Chat.ID
				respponse := fmt.Sprintf("Вот ваша задача: (%v) \nВот ваш ID: (%v)", task, userId)
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, respponse)
				bot.Send(msg)
				userState[update.Message.Chat.ID] = ""
			}
		} else if update.CallbackQuery != nil {
			var response string

			switch update.CallbackQuery.Data {
			case "day":
				response = "Это ваши задачи на сегодня"
			case "planned":
				response = "Это ваши запланированные задачи"
			case "tasks":
				response = "Вот список ваших задач"
			case "create_list":
				response = "Введите название нового списка"
			}

			msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, response)
			bot.Send(msg)
		}
	}
}
