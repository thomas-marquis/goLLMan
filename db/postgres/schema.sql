CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE books (
   id SERIAL PRIMARY KEY,
   title TEXT NOT NULL,
   author TEXT NOT NULL
);

CREATE TABLE book_index (
    id SERIAL PRIMARY KEY,
    content TEXT NOT NULL,
    hash TEXT NOT NULL UNIQUE,
    book_id INTEGER NOT NULL,
    embedding vector(1024) NOT NULL,
    FOREIGN KEY (book_id) REFERENCES books(id)
);