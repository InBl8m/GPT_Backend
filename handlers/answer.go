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

func SaveAnswerHandler(w http.ResponseWriter, r *http.Request) {
	var answer models.Answer
	err := json.NewDecoder(r.Body).Decode(&answer)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	result, err := db.DB.Exec("INSERT INTO answers (answer) VALUES (?)", answer.Answer)
	if err != nil {
		http.Error(w, "Failed to save answer", http.StatusInternalServerError)
		return
	}

	id, _ := result.LastInsertId()
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "Answer saved with ID %d", id)
}

func GetAnswersHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.DB.Query("SELECT id, answer, processed FROM answers")
	if err != nil {
		http.Error(w, "Failed to retrieve answers", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var answers []models.Answer
	for rows.Next() {
		var answer models.Answer
		if err := rows.Scan(&answer.ID, &answer.Answer, &answer.Processed); err != nil {
			http.Error(w, "Failed to scan answer", http.StatusInternalServerError)
			return
		}
		answers = append(answers, answer)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(answers)
}

func GetLatestUnprocessedAnswerHandler(w http.ResponseWriter, r *http.Request) {
	var answer models.Answer

	row := db.DB.QueryRow("SELECT id, answer, processed FROM answers WHERE processed = 0 ORDER BY id ASC LIMIT 1")
	err := row.Scan(&answer.ID, &answer.Answer, &answer.Processed)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "No unprocessed answers found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to retrieve answer", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(answer)
}

func MarkAnswerProcessedHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	row := db.DB.QueryRow("SELECT processed FROM answers WHERE id = ?", id)
	var processed int
	err := row.Scan(&processed)
	if err == nil {
		if processed == 1 {
			http.Error(w, "Answer is already processed", http.StatusConflict)
			return
		}
		_, err = db.DB.Exec("UPDATE answers SET processed = 1 WHERE id = ?", id)
	}

	if err != nil {
		http.Error(w, "Answer not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Answer with ID %s marked as processed", id)
}
