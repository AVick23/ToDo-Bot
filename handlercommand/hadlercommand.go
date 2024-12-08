package handlercommand

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"time"

	"github.com/AVick23/ToDo-Bot/database"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var (
	response  string
	buttons   [][]tgbotapi.InlineKeyboardButton
	emojiList = []string{"😊", "🚀", "✨", "🌟", "💡", "✅", "📅", "📌", "🕒", "🎯"}
)

func HandleDefaultCallback(bot *tgbotapi.BotAPI, update tgbotapi.Update, db *sql.DB, username string, callbackData string) {

	tasks, err := database.GetTasks(db, username)
	if err != nil {
		log.Printf("Ошибка получения задач: %v", err)
		msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "❌ Произошла ошибка при загрузке задач. Пожалуйста, попробуйте ещё раз.")
		bot.Send(msg)
		return
	}

	tasksplan, err := database.GetTasksPlanned(db, username)
	if err != nil {
		log.Printf("Ошибка получения запланированных задач: %v", err)
		msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "❌ Не удалось загрузить запланированные задачи. Попробуйте позже.")
		bot.Send(msg)
		return
	}

	taskday, err := database.GetTasksDay(db, username)
	if err != nil {
		log.Printf("Ошибка получения задач на сегодня: %v", err)
		msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "❌ Не удалось загрузить задачи на сегодня. Попробуйте позже.")
		bot.Send(msg)
		return
	}

	buttons = [][]tgbotapi.InlineKeyboardButton{}

	switch callbackData {
	case "day":
		response = "📅 *Ваши задачи на сегодня:*\n"
		for _, task := range taskday {
			buttonText := fmt.Sprintf("%s %s (до %s)", getRandomEmoji(), task.Task, task.Date)
			buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(buttonText, "task_"+task.Task),
			))
		}
	case "planned":
		response = "🗓️ *Ваши запланированные задачи:*\n"
		for _, task := range tasksplan {
			buttonText := fmt.Sprintf("%s %s (до %s)", getRandomEmoji(), task.Task, task.Date)
			if task.Notification.Valid {
				buttonText += fmt.Sprintf(" 🔔 %s", task.Notification.String)
			}
			buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(buttonText, "task_"+task.Task),
			))
		}
	case "tasks":
		response = "📋 *Вот ваш список задач:*\n"
		for _, task := range tasks {
			buttonText := fmt.Sprintf("%s %s", getRandomEmoji(), task)
			buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(buttonText, "task_"+task),
			))
		}
	case "create_list":
		response = "✏️ Пожалуйста, введите название для нового списка задач:"
	}

	if len(buttons) > 0 {
		replyMarkup := tgbotapi.NewInlineKeyboardMarkup(buttons...)
		msg := tgbotapi.NewEditMessageTextAndMarkup(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, response, replyMarkup)
		bot.Send(msg)
	} else {
		msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "ℹ️ К сожалению, у вас пока нет задач.")
		bot.Send(msg)
	}
}

func getRandomEmoji() string {
	rand.Seed(time.Now().UnixNano())
	return emojiList[rand.Intn(len(emojiList))]
}

func CheckAndSendReminders(db *sql.DB, bot *tgbotapi.BotAPI) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	loc, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		log.Fatalf("Ошибка загрузки часового пояса: %v", err)
	}

	for range ticker.C {

		currentTime := time.Now().In(loc)
		currentDate := currentTime.Format("02.01.2006")
		currentTimeStr := currentTime.Format("15:04")

		rows, err := db.Query(`
            SELECT u.username, t.tasks, t.date, t.notification 
            FROM tasks t 
            INNER JOIN users u ON t.user_id = u.id 
            WHERE t.date = $1 AND t.notification = $2 AND t.notification IS NOT NULL`,
			currentDate, currentTimeStr)
		if err != nil {
			log.Printf("Ошибка запроса задач: %v", err)
			continue
		}
		defer rows.Close()

		for rows.Next() {
			var username, task, date, notification string
			if err := rows.Scan(&username, &task, &date, &notification); err != nil {
				log.Printf("Ошибка сканирования: %v", err)
				continue
			}

			chatID, err := toInt64(username)
			if err != nil {
				log.Printf("Ошибка преобразования username в chatID: %v", err)
				continue
			}

			message := fmt.Sprintf("🔔 Напоминание: %s\n📅 Дата: %s\n⏰ Время: %s", task, date, notification)
			msg := tgbotapi.NewMessage(chatID, message)

			if _, err := bot.Send(msg); err != nil {
				log.Printf("Ошибка отправки сообщения: %v", err)
			}
		}
	}
}

func toInt64(username string) (int64, error) {
	chatID, err := strconv.ParseInt(username, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("не удалось преобразовать username: %s", username)
	}
	return chatID, nil
}
