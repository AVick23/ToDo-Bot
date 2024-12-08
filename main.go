package main

import (
	"fmt"
	"log"

	"github.com/AVick23/ToDo-Bot/botfunc"
	"github.com/AVick23/ToDo-Bot/database"
	"github.com/AVick23/ToDo-Bot/handlercommand"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	bot, err := botfunc.NewBot("7699728760:AAGsMWGdlQsyI0q7dxR5by1pJaHBApj45_k")
	if err != nil {
		log.Fatal(err)
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	db, err := database.ConnectDB()
	if err != nil {
		log.Fatalf("Не получилось подключиться к БД: %v", err)
	}
	defer db.Close()

	fmt.Println("Всё работает, ежи братка.")

	go handlercommand.CheckAndSendReminders(db, bot)

	botfunc.RunProcess(bot, updates, db)
}
