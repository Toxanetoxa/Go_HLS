CREATE TABLE videos (
                        id SERIAL PRIMARY KEY,          -- Автоинкрементируемый первичный ключ
                        title TEXT NOT NULL,            -- Название видео
                        file_path TEXT NOT NULL,        -- Путь к файлу
                        author_id INTEGER NOT NULL,     -- ID автора
                        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, -- Время создания записи
                        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP -- Время обновления записи
);