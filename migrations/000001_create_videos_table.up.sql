CREATE TABLE videos (
                        author_id SERIAL PRIMARY KEY,
                        title TEXT NOT NULL,
                        file_path TEXT NOT NULL,
                        user_id INTEGER NOT NULL,
                        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);