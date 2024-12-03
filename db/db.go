package db

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

var DB *sql.DB

// Инициализация базы данных
func InitDB() {
	var err error

	// Подключение к PostgreSQL
	connStr := "host=localhost port=5432 user=postgres password=password dbname=back sslmode=disable"
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}

	// Проверка соединения
	if err = DB.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	fmt.Println("Connected to PostgreSQL!")

	// Создаем таблицы, если они еще не существуют
	_, err = DB.Exec(`
		CREATE TABLE IF NOT EXISTS pages (
			id SERIAL PRIMARY KEY,
			html TEXT NOT NULL,
			processed BOOLEAN NOT NULL DEFAULT false
		);
	`)
	if err != nil {
		log.Fatalf("Failed to create pages table: %v", err)
	}

	_, err = DB.Exec(`
		CREATE TABLE IF NOT EXISTS answers (
			id SERIAL PRIMARY KEY,
			answer TEXT NOT NULL,
			processed BOOLEAN DEFAULT false
		);
	`)
	if err != nil {
		log.Fatalf("Failed to create answers table: %v", err)
	}

	_, err = DB.Exec(`
		CREATE TABLE IF NOT EXISTS submit_requests (
			id SERIAL PRIMARY KEY,
			request TEXT NOT NULL,
			processed BOOLEAN NOT NULL DEFAULT false
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
