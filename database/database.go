package database

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq" // Подключаем драйвер для PostgreSQL
)

type DB struct {
	*sql.DB
}

// ConnectDB подключается к PostgreSQL
func ConnectDB() (*sql.DB, error) {
	// Строка подключения к PostgreSQL
	connStr := "postgres://avick123:super123@127.0.0.1:5432/todo_db?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("ошибка подключения к базе данных: %w", err)
	}

	// Проверяем подключение
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("не удалось установить соединение: %w", err)
	}

	// Запросы для создания таблиц
	createTableDB := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			username TEXT NOT NULL
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
	}

	// Выполняем каждый запрос
	for _, query := range createTableDB {
		_, err := db.Exec(query)
		if err != nil {
			return nil, fmt.Errorf("ошибка при выполнении запроса: %w", err)
		}
	}

	fmt.Println("Подключение к PostgreSQL успешно!")
	return db, nil
}
