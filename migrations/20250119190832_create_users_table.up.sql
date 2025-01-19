CREATE TABLE users (
                       id SERIAL PRIMARY KEY,                      -- Уникальный идентификатор пользователя
                       username VARCHAR(255) UNIQUE NOT NULL,      -- Уникальное имя пользователя
                       password VARCHAR(255) NOT NULL,             -- Хешированный пароль
                       email VARCHAR(255) UNIQUE NOT NULL,         -- Уникальный email
                       created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, -- Время создания записи
                       updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, -- Время последнего обновления записи
                       deleted_at TIMESTAMP DEFAULT NULL           -- Время удаления (для мягкого удаления)
);