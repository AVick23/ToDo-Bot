package complete

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/AVick23/ToDo-Bot/database"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func CompleteTasks(bot *tgbotapi.BotAPI, update tgbotapi.Update, username string, task string, db *sql.DB) {
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
