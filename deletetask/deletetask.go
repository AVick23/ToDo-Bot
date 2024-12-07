package deletetask

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/AVick23/ToDo-Bot/database"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func DeleteTasks(bot *tgbotapi.BotAPI, update tgbotapi.Update, username string, task string, db *sql.DB) {
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
