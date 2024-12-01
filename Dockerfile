FROM golang:1.23.1

WORKDIR /app

# Копируем модули (для кэширования)
COPY go.mod go.sum ./

# Загружаем зависимости
RUN go mod tidy

# Копируем оставшиеся файлы
COPY . .

# Компилируем приложения
RUN go build -o main main.go
RUN go build -o unprocessed unprocessed_page.go

CMD ["sh", "-c", "./main & ./unprocessed"]
