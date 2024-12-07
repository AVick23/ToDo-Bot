package models

import "database/sql"

type Task struct {
	Description string
	Date        *string
	Time        *string
}

type Tasks struct {
	Task         string
	Date         string
	Notification sql.NullString
}

type Day struct {
	Task string
	Date string
}
