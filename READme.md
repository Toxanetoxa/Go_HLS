##  Сборка и запуск приложения:
В проекте используется Make все команды можно посмотреть 
```shell
  make help 
```
## Рекомендуемая команда при первом запуске:
```shell
   make start
```

## Опционально:

### Команда бля билда докер образа:
```shell
  make build
```
### Команда для поднятия контейнеров в дев режиме
```shell
  make up
```

#### Важно добавить в **etc/hosts** строчку
```
127.0.0.1   backend.app.loc
```

## Приложение будет доступно по адресу (если вы следовали инструкции):
### [backend.app.loc](http://backend.app.loc)

## Если зашли на http://backend.app.loc и вас редиректнуло на гифку значит бекенд работает)

и на 80 порту (точнее его будет проксировать Nginx) и на порту 8080 (сам контейнер приложения)
(*важно чтобы у вас ничего другого не крутилось в фоновом режиме на этих портах)
Redis будет крутиться в network приложения и он будет доступен по redis:6379

## Основные компоненты:
1. Само приложение на Go:
   * Go 1.22
   * Gin  HTTP-веб-фреймворк
2. Nginx - в качестве прокси-сервера
3. Docker и Docker-Compose для разработки, доставки и запуска приложения в изолированном контейнере.

# ТЗ:

## Цель
Создать прототип api-first (backend отдельно от фронтенд) мини-сервиса на Go (версии 1.21+) для видеостриминга (условный «мини-YouTube») со следующими возможностями:
- Регистрация и авторизация
- Получение и загрузка видеоконтента.
- Подсчёт просмотров и отображение количества текущих зрителей в реальном времени.

## Задачи
1.Авторизация и регистрация
Реализовать регистрацию и авторизацию пользователей на базе JWT.
2. Загрузка видео
Поддержать загрузку хотя бы одного небольшого файла.
Хранить метаданные видео (название, путь к файлу, автор и т. п.) в базе данных PostgreSQL.
3. Просмотр видео с частичной загрузкой
Создать эндпоинт, позволяющий «пролистывать» видео без полной перезагрузки.
4. Счётчик просмотров и активных зрителей
Отслеживать в реальном времени, сколько пользователей прямо сейчас смотрит видео.
Отображать общее количество просмотров (как минимум увеличивать счётчик при первом просмотре). Напишите как бы вы решили проблему борьбы с накруткой счетчика просмотра в реальном проекте.
5. Минимальный фронтенд (опционально, но приветствуется)
Будет плюсом, если для демонстрации будет сделан простой HTML-интерфейс (без сложных стилей).


# Маршруты:
- POST /auth/register — регистрация (email + хешированный пароль).
- POST /auth/login — получение JWT-токена.
- GET /auth/me — получение данных пользователя (с проверкой токена).
- POST /videos/upload — загрузка видео (multipart/form-data).
- GET /videos/{id}/stream 

# Технологии
- Backend: Go 1.21 + Gin
- Database: PostgreSQL (gorm)
- Caching: Redis (активные зрители)
- JWT: github.com/golang-jwt/jwt/v5
- Storage: Локально (storage/) или MinIO (S3)
- WebSocket: github.com/gorilla/websocket

### Тестовый запрос:
curl -X POST http://backend.app.loc/videos/upload \
-F "file=@/Users/antonsotnik/Documents/meine/test.mov" \
-F "title=My Video" \
-F "author_id=1"