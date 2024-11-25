package main

import (
	"database/sql"
	"fmt"
	"golang.org/x/net/html"
	"log"
	_ "modernc.org/sqlite"
	"strings"
	"sync"
	"time"
)

var (
	assistantMessages []string   // Глобальный слайс для хранения всех уникальных сообщений
	mu                sync.Mutex // Мьютекс для синхронизации доступа к `assistantMessages`
	previousMsg       string
)

// Функция для получения последней необработанной страницы из базы данных
func getLastUnprocessedPage(db *sql.DB) (int, string, error) {
	var id int
	var htmlContent string

	// Запрос для получения последней необработанной страницы
	row := db.QueryRow("SELECT id, html FROM pages WHERE processed = 0 ORDER BY id DESC LIMIT 1")
	err := row.Scan(&id, &htmlContent)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, "", fmt.Errorf("no unprocessed pages found")
		}
		return 0, "", fmt.Errorf("failed to retrieve page: %v", err)
	}
	return id, htmlContent, nil
}

// Функция для обновления страницы как обработанной
func markPageAsProcessed(db *sql.DB, pageID int) error {
	_, err := db.Exec("UPDATE pages SET processed = 1 WHERE id = ?", pageID)
	return err
}

// Рекурсивный обход HTML-дерева для поиска div с атрибутом data-message-author-role="assistant"
func findDivsWithRole(n *html.Node) string {
	if n.Type == html.ElementNode && n.Data == "div" {
		for _, attr := range n.Attr {
			if attr.Key == "data-message-author-role" && attr.Val == "assistant" {
				return getElementText(n)
			}
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		text := findDivsWithRole(c)
		if text != "" {
			return text
		}
	}
	return ""
}

// Функция для извлечения текстового содержимого из узла HTML
func getElementText(n *html.Node) string {
	var sb strings.Builder
	var getText func(*html.Node)
	getText = func(n *html.Node) {
		if n.Type == html.TextNode {
			sb.WriteString(n.Data)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			getText(c)
		}
	}
	getText(n)
	return sb.String()
}

// Проверка, существует ли сообщение в массиве
func messageExists(message string) bool {
	for _, m := range assistantMessages {
		if m == message {
			return true
		}
	}
	return false
}

// Функция для сохранения сообщения в таблицу answers
func saveMessageToDB(db *sql.DB, message string) error {
	for i := 0; i < 5; i++ { // Попробуем 5 раз
		_, err := db.Exec("INSERT INTO answers (answer) VALUES (?)", message)
		if err != nil {
			if strings.Contains(err.Error(), "database is locked") {
				time.Sleep(100 * time.Millisecond) // Задержка перед повтором
				continue
			}
			return err // Любая другая ошибка
		}
		return nil // Успешно записали
	}
	return fmt.Errorf("failed to save message after multiple attempts")
}

// Функция для периодической печати сообщений и сохранения в БД
func printMessages(db *sql.DB) {
	for {
		mu.Lock()
		if len(assistantMessages) > 0 {
			currentMsg := assistantMessages[0]
			fmt.Println("Length of assistantMessages:", len(assistantMessages))

			// Вычисляем отличающуюся часть
			diff := getDifference(previousMsg, currentMsg)
			fmt.Println("Message difference:", diff)

			// Сохраняем текущее сообщение как предыдущее
			previousMsg = currentMsg

			// Если есть отличие, сохраняем его в базу данных
			if diff != "" {
				err := saveMessageToDB(db, diff)
				if err != nil {
					log.Printf("Failed to save message to DB: %v", err)
				}
			}

			// Удаляем первый элемент после обработки
			assistantMessages = assistantMessages[1:]
		}
		mu.Unlock()

		// Задержка между обработкой сообщений
		time.Sleep(2 * time.Second)
	}
}

func getDifference(prev, current string) string {
	for i := 0; i < len(prev) && i < len(current); i++ {
		if prev[i] != current[i] {
			return current[i:]
		}
	}
	if len(prev) < len(current) {
		return current[len(prev):]
	}
	return ""
}

func main() {
	// Подключаемся к базе данных SQLite
	db, err := sql.Open("sqlite", "app.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Запускаем горутину для печати сообщений и сохранения в БД
	go printMessages(db)

	for {
		// Получаем последнюю необработанную страницу из базы данных
		pageID, htmlContent, err := getLastUnprocessedPage(db)
		if err != nil {
			if err.Error() == "no unprocessed pages found" {
				fmt.Println("No new unprocessed pages. Retrying...")
			} else {
				log.Printf("Error fetching page: %v", err)
			}
			time.Sleep(5 * time.Second)
			continue
		}

		// Парсинг HTML
		doc, err := html.Parse(strings.NewReader(htmlContent))
		if err != nil {
			log.Printf("Error parsing HTML: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}

		// Ищем нужный div и получаем текст
		currentText := findDivsWithRole(doc)
		if currentText != "" && !messageExists(currentText) {
			mu.Lock()
			assistantMessages = append(assistantMessages, currentText)
			mu.Unlock()

			// Обновляем статус страницы как обработанной
			err := markPageAsProcessed(db, pageID)
			if err != nil {
				log.Printf("Failed to mark page as processed: %v", err)
			}
		} else if messageExists(currentText) {
			fmt.Println("Message already processed.")
		}

		// Задержка перед следующим обходом
		time.Sleep(5 * time.Second)
	}
}
