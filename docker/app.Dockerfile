# Используем образ с Go для сборки
FROM golang:1.22-bullseye AS builder

WORKDIR /app

# Копируем файлы для установки зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь проект
COPY . .

# Собираем исполняемый файл для основного приложения
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /usr/local/bin/app ./cmd/go_hls

# Собираем исполняемый файл для миграций
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /usr/local/bin/migrate ./cmd/migrate

# Проверяем, что файлы скомпилированы
RUN ls -la /usr/local/bin/app /usr/local/bin/migrate

# Используем минимальный образ для запуска приложения
FROM ubuntu:20.04 AS runner

# Копируем собранные исполняемые файлы из builder-образа
COPY --from=builder /usr/local/bin/app /usr/local/bin/app
COPY --from=builder /usr/local/bin/migrate /usr/local/bin/migrate

# Убедимся, что файлы имеют права на выполнение
RUN chmod +x /usr/local/bin/app /usr/local/bin/migrate

# Проверяем, что файлы скопированы и имеют права на выполнение
RUN ls -la /usr/local/bin

# Указываем переменную окружения для исполняемых файлов
ENV PATH="/usr/local/bin:$PATH"

EXPOSE 8080

# Запускаем приложение
CMD ["app"]