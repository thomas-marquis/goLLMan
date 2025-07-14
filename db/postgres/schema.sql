CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE books (
   id SERIAL PRIMARY KEY,
   title TEXT NOT NULL,
   author TEXT NOT NULL,
   metadata JSONB
);

CREATE TABLE book_index (
    id SERIAL PRIMARY KEY,
    content TEXT NOT NULL,
    book_id INTEGER NOT NULL,
    embedding vector(1024) NOT NULL,
    FOREIGN KEY (book_id) REFERENCES books(id)
);