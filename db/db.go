package db

import (
	"database/sql"
	"log"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

// Инициализация базы данных
func InitDB() {
	var err error
	DB, err = sql.Open("sqlite", "app.db")
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}

	// Создаем таблицы, если они еще не существуют
	// Для каждой таблицы выполняем отдельную команду CREATE TABLE IF NOT EXISTS
	_, err = DB.Exec(`
		CREATE TABLE IF NOT EXISTS pages (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			html TEXT NOT NULL,
			processed BOOLEAN NOT NULL DEFAULT 0
		);
	`)
	if err != nil {
		log.Fatalf("Failed to create pages table: %v", err)
	}

	_, err = DB.Exec(`
		CREATE TABLE IF NOT EXISTS answers (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			answer TEXT NOT NULL,
			processed BOOLEAN DEFAULT 0
		);
	`)
	if err != nil {
		log.Fatalf("Failed to create answers table: %v", err)
	}

	_, err = DB.Exec(`
		CREATE TABLE IF NOT EXISTS submit_requests (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			request TEXT NOT NULL,
			processed BOOLEAN NOT NULL DEFAULT 0
		);
	`)
	if err != nil {
		log.Fatalf("Failed to create submit_requests table: %v", err)
	}
}

func CloseDB() {
	if err := DB.Close(); err != nil {
		log.Fatalf("Failed to close database: %v", err)
	}
}
