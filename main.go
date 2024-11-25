package main

import (
	"GPT_Backend/db"
	"GPT_Backend/handlers"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

// Middleware для включения CORS
func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")                                // Разрешаем доступ с любых доменов
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS") // Разрешенные методы
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")     // Разрешенные заголовки
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	// Инициализация базы данных
	db.InitDB()
	defer db.CloseDB()

	// Настройка роутера
	r := mux.NewRouter()
	r.Use(enableCORS) // Применяем middleware для всех маршрутов

	r.HandleFunc("/save_page", handlers.SavePageHandler).Methods("POST")
	r.HandleFunc("/save_request", handlers.SaveRequestHandler).Methods("POST")
	r.HandleFunc("/pages", handlers.GetPagesHandler).Methods("GET")
	r.HandleFunc("/requests", handlers.GetRequestsHandler).Methods("GET")
	r.HandleFunc("/latest_unprocessed_request", handlers.GetLatestUnprocessedRequestHandler).Methods("GET") // Новый маршрут
	r.HandleFunc("/mark_processed/{id}", handlers.MarkProcessedHandler).Methods("POST")
	r.HandleFunc("/save_answer", handlers.SaveAnswerHandler).Methods("POST")
	r.HandleFunc("/answers", handlers.GetAnswersHandler).Methods("GET")
	r.HandleFunc("/latest_unprocessed_answer", handlers.GetLatestUnprocessedAnswerHandler).Methods("GET")
	r.HandleFunc("/mark_answer_processed/{id}", handlers.MarkAnswerProcessedHandler).Methods("POST")

	// Запуск сервера
	fmt.Println("Server is running on port 8080")
	log.Fatal(http.ListenAndServe("0.0.0.0:8080", r))
}
