package handlers

import (
	"GPT_Backend/db"
	"GPT_Backend/models"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
)

// Обработчик для сохранения HTML-кода страницы
func SavePageHandler(w http.ResponseWriter, r *http.Request) {
	var page models.Page
	err := json.NewDecoder(r.Body).Decode(&page)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Проверяем последнюю запись в базе данных
	var lastHTML string
	err = db.DB.QueryRow("SELECT html FROM pages ORDER BY id DESC LIMIT 1").Scan(&lastHTML)
	if err != nil && err != sql.ErrNoRows {
		http.Error(w, "Failed to check last page", http.StatusInternalServerError)
		return
	}

	// Если последняя страница совпадает с текущей, пропускаем сохранение
	if lastHTML == page.HTML {
		http.Error(w, "Page is identical to the last one", http.StatusConflict)
		return
	}

	// Сохранение HTML в базу данных
	result, err := db.DB.Exec("INSERT INTO pages (html) VALUES (?)", page.HTML)
	if err != nil {
		http.Error(w, "Failed to save page", http.StatusInternalServerError)
		return
	}

	id, _ := result.LastInsertId()
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "Page saved with ID %d", id)
}

// Обработчик для получения всех страниц
func GetPagesHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.DB.Query("SELECT id, html, processed FROM pages")
	if err != nil {
		http.Error(w, "Failed to retrieve pages", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var pages []models.Page
	for rows.Next() {
		var page models.Page
		if err := rows.Scan(&page.ID, &page.HTML, &page.Processed); err != nil {
			http.Error(w, "Failed to scan page", http.StatusInternalServerError)
			return
		}
		pages = append(pages, page)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pages)
}
