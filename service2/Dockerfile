# Базовый образ
FROM golang:1.19.5

# Установка директории приложения
WORKDIR /app

# Копируем все файлы в директорию приложения
COPY app/go.mod ./
COPY app/go.sum ./

# Установка зависимостей
RUN go mod download

COPY app/*.go ./

# Компиляция приложения
RUN go build -o main .

# Запуск приложения
CMD ["/app/main"]