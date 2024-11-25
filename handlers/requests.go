package handlers

import (
	"GPT_Backend/db"
	"GPT_Backend/models"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
)

// Обработчик для сохранения запроса
func SaveRequestHandler(w http.ResponseWriter, r *http.Request) {
	var request models.SubmitRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Сохранение запроса в базу данных
	result, err := db.DB.Exec("INSERT INTO submit_requests (request) VALUES (?)", request.Request)
	if err != nil {
		http.Error(w, "Failed to save request", http.StatusInternalServerError)
		return
	}

	id, _ := result.LastInsertId()
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "Request saved with ID %d", id)
}

// Обработчик для получения всех запросов
func GetRequestsHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.DB.Query("SELECT id, request, processed FROM submit_requests")
	if err != nil {
		http.Error(w, "Failed to retrieve requests", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var requests []models.SubmitRequest
	for rows.Next() {
		var request models.SubmitRequest
		if err := rows.Scan(&request.ID, &request.Request, &request.Processed); err != nil {
			http.Error(w, "Failed to scan request", http.StatusInternalServerError)
			return
		}
		requests = append(requests, request)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(requests)
}

// Обработчик для получения последнего необработанного запроса
func GetLatestUnprocessedRequestHandler(w http.ResponseWriter, r *http.Request) {
	var request models.SubmitRequest

	// Запрос для получения последнего необработанного запроса
	row := db.DB.QueryRow("SELECT id, request, processed FROM submit_requests WHERE processed = 0 ORDER BY id DESC LIMIT 1")

	// Сканируем результат
	err := row.Scan(&request.ID, &request.Request, &request.Processed)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "No unprocessed requests found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to retrieve request", http.StatusInternalServerError)
		}
		return
	}

	// Возвращаем результат в формате JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(request)
}

// Обработчик для пометки записи как обработанной по указанному ID
func MarkProcessedHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Проверка на обработанный запрос
	row := db.DB.QueryRow("SELECT processed FROM submit_requests WHERE id = ?", id)
	var processed int
	err := row.Scan(&processed)
	if err == nil {
		if processed == 1 {
			http.Error(w, "Item is already processed", http.StatusConflict)
			return
		}
		_, err = db.DB.Exec("UPDATE submit_requests SET processed = 1 WHERE id = ?", id)
	} else {
		// Запрос не найден в submit_requests, проверяем таблицу pages
		row = db.DB.QueryRow("SELECT processed FROM pages WHERE id = ?", id)
		err = row.Scan(&processed)
		if err == nil {
			if processed == 1 {
				http.Error(w, "Item is already processed", http.StatusConflict)
				return
			}
			_, err = db.DB.Exec("UPDATE pages SET processed = 1 WHERE id = ?", id)
		}
	}

	if err != nil {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Item with ID %s marked as processed", id)
}
