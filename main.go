package main

import (
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var targetNumber int
var attempts int

func main() {
	bot, err := tgbotapi.NewBotAPI("7699728760:AAGsMWGdlQsyI0q7dxR5by1pJaHBApj45_k")
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Авторизация под аккаунтом %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			switch update.Message.Text {
			case "/start":
				targetNumber = rand.New(rand.NewSource(time.Now().UnixNano())).Intn(100) + 1
				attempts = 10
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Привет! Я загадал число от 1 до 100. Попробуешь угадать его? У тебя есть 5 попыток.")
				inlineButton := tgbotapi.NewInlineKeyboardMarkup(
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("Начать угадывать", "start_game"),
					),
				)
				msg.ReplyMarkup = inlineButton
				bot.Send(msg)
			default:
				guess, err := strconv.Atoi(update.Message.Text)
				if err != nil {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Пожалуйста, введите число.")
					bot.Send(msg)
					continue
				}

				attempts--
				response := ""

				if guess == targetNumber {
					response = fmt.Sprintf("Поздравляю! Вы угадали число %d.", targetNumber)
				} else if attempts == 0 {
					response = fmt.Sprintf("Вы исчерпали все попытки. Я загадывал число %d.", targetNumber)
				} else {
					hint := "меньше"
					if guess < targetNumber {
						hint = "больше"
					}
					if abs(guess-targetNumber) <= 10 {
						response = fmt.Sprintf("Горячо! Число %s. Попробуйте еще раз.", hint)
					} else {
						response = fmt.Sprintf("Холодно. Число %s. Попробуйте еще раз.", hint)
					}
				}

				msg := tgbotapi.NewMessage(update.Message.Chat.ID, response)
				bot.Send(msg)
			}
		}

		if update.CallbackQuery != nil {
			switch update.CallbackQuery.Data {
			case "start_game":
				msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Введите ваше предположение:")
				bot.Send(msg)
			}

			callback := tgbotapi.NewCallback(update.CallbackQuery.ID, update.CallbackQuery.Data)
			if _, err := bot.Request(callback); err != nil {
				log.Println(err)
			}
		}
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
