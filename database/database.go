package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/AVick23/ToDo-Bot/models"
	_ "github.com/lib/pq"
)

type DB struct {
	*sql.DB
}

func CreateDB() (*sql.DB, error) {
	connStr := "postgres://avick123:super123@127.0.0.1:5432/todo_db?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("ошибка подключения к базе данных: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("не удалось установить соединение: %w", err)
	}

	createTableDB := []string{
		`CREATE TABLE IF NOT EXISTS users (
            id SERIAL PRIMARY KEY,
            username TEXT NOT NULL UNIQUE
        )`,
		`CREATE TABLE IF NOT EXISTS tasks (
            id SERIAL PRIMARY KEY,
            user_id INTEGER NOT NULL,
            tasks TEXT,
            date TEXT,
            list_name TEXT,
            notification TEXT,
            FOREIGN KEY(user_id) REFERENCES users(id)
        )`,
		`CREATE TABLE IF NOT EXISTS completed ( 
		id SERIAL PRIMARY KEY, 
		user_id BIGINT NOT NULL, 
		task TEXT NOT NULL, 
		FOREIGN KEY(user_id) REFERENCES users(id) 
		)`,
	}

	for _, query := range createTableDB {
		_, err := db.Exec(query)
		if err != nil {
			return nil, fmt.Errorf("ошибка при выполнении запроса: %w", err)
		}
	}
	return db, nil
}

func ConnectDB() (*sql.DB, error) {
	connStr := "postgres://avick123:super123@127.0.0.1:5432/todo_db?sslmode=disable"
	return sql.Open("postgres", connStr)
}

func SaveUser(db *sql.DB, username string) (int, error) {
	var id int
	err := db.QueryRow("INSERT INTO users (username) VALUES ($1) ON CONFLICT (username) DO UPDATE SET username=EXCLUDED.username RETURNING id", username).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("ошибка сохранения пользователя: %w", err)
	}
	return id, nil
}

func SaveTasks(db *sql.DB, userID int, task models.Task) error {

	if task.Description != "" {
		_, err := db.Exec("INSERT INTO tasks (user_id, tasks, date, notification) VALUES ($1, $2, $3, $4)", userID, task.Description, task.Date, task.Time)
		if err != nil {
			return fmt.Errorf("не удалось сохранить задачу %v", err)
		}
	}
	return nil
}

func GetTasks(db *sql.DB, username string) ([]string, error) {
	var tasks []string
	var id int
	err := db.QueryRow("SELECT id FROM users WHERE username = $1", username).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения userID: %w", err)
	}
	rows, err := db.Query("SELECT tasks FROM tasks WHERE user_id = $1", id)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнение задачи %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var task string
		err := rows.Scan(&task)
		if err != nil {
			return nil, fmt.Errorf("ошибка сканирования задач %v", err)
		}
		tasks = append(tasks, task)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при итерации %v", err)
	}
	return tasks, nil
}

func GetTasksPlanned(db *sql.DB, username string) ([]models.Tasks, error) {
	var tasks []models.Tasks
	var id int

	err := db.QueryRow("SELECT id FROM users WHERE username = $1", username).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения userID: %w", err)
	}

	rows, err := db.Query("SELECT tasks, date, notification FROM tasks WHERE user_id = $1 AND tasks IS NOT NULL AND date IS NOT NULL AND (notification IS NULL OR notification IS NOT NULL)", id)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения задачи: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var task models.Tasks
		err := rows.Scan(&task.Task, &task.Date, &task.Notification)
		if err != nil {
			return nil, fmt.Errorf("ошибка сканирования задач: %v", err)
		}

		tasks = append(tasks, task)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при итерации: %v", err)
	}
	return tasks, nil
}

func CompleteTasksDB(db *sql.DB, username string, task string) error {
	var id int

	err := db.QueryRow("SELECT id FROM users WHERE username = $1", username).Scan(&id)
	if err != nil {
		return fmt.Errorf("ошибка получения userID: %w", err)
	}

	deleteTaskSQL := "DELETE FROM tasks WHERE user_id = $1 AND tasks = $2"
	res, err := db.Exec(deleteTaskSQL, id, task)
	if err != nil {
		return fmt.Errorf("не удалось удалить задачу у пользователя: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("ошибка получения количества удаленных строк: %w", err)
	}
	if rowsAffected == 0 {
		fmt.Printf("Задача '%s' для пользователя '%s' не найдена\n", task, username)
		return fmt.Errorf("задача не найдена")
	}

	_, err = db.Exec("INSERT INTO completed (user_id, task) VALUES ($1, $2)", id, task)
	if err != nil {
		return fmt.Errorf("ошибка сохранения задачи: %w", err)
	}

	fmt.Printf("Задача '%s' для пользователя '%s' успешно выполнена и удалена\n", task, username)
	return nil
}

func DeleteTaskSQL(db *sql.DB, task string, username string) error {
	var id int
	var idtask int

	err := db.QueryRow("SELECT id FROM users WHERE username = $1", username).Scan(&id)
	if err != nil {
		return fmt.Errorf("ошибка получения id в таблице users: %w", err)
	}

	err = db.QueryRow("SELECT id FROM tasks WHERE user_id = $1 AND tasks = $2", id, task).Scan(&idtask)
	if err != nil {
		return fmt.Errorf("ошибка получения id в таблице tasks: %v", err)
	}

	res, err := db.Exec("DELETE FROM tasks WHERE id = $1", idtask)
	if err != nil {
		return fmt.Errorf("не удалось удалить задачу у пользователя: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("ошибка получения количества удаленных строк: %w", err)
	}
	if rowsAffected == 0 {
		fmt.Printf("Задача '%s' для пользователя '%s' не найдена\n", task, username)
		return fmt.Errorf("задача не найдена")
	}
	return nil
}

func GetTasksDay(db *sql.DB, username string) ([]models.Day, error) {
	var tasks []models.Day
	var id int

	err := db.QueryRow("SELECT id FROM users WHERE username = $1", username).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения userID: %w", err)
	}

	rows, err := db.Query("SELECT tasks, date FROM tasks WHERE user_id = $1 AND tasks IS NOT NULL AND date IS NOT NULL", id)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения задачи: %v", err)
	}
	defer rows.Close()

	today := time.Now().Format("02.01.2006")

	for rows.Next() {
		var task models.Day
		err := rows.Scan(&task.Task, &task.Date)
		if err != nil {
			return nil, fmt.Errorf("ошибка сканирования задач: %v", err)
		}

		if task.Date == today {
			tasks = append(tasks, task)
		} else {
			log.Printf("Некорректная дата или дата не совпадает с сегодняшней: %v", task.Date)
		}
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при итерации: %v", err)
	}

	return tasks, nil
}
