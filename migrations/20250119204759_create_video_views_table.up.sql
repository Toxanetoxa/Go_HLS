CREATE TABLE video_views (
                             id SERIAL PRIMARY KEY,          -- Автоинкрементируемый первичный ключ
                             video_id INTEGER NOT NULL,      -- ID видео
                             ip_address TEXT NOT NULL,       -- IP-адрес пользователя
                             created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, -- Время создания записи
                             FOREIGN KEY (video_id) REFERENCES videos(id) ON DELETE CASCADE
);