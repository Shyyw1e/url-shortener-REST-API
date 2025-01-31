package posql

import (
	"database/sql"
	"fmt"

	"url-shorter-REST-API/internal/storage"

	"github.com/lib/pq" // Драйвер PostgreSQL
)

type Storage struct {
	db *sql.DB
}

// New создает подключение к PostgreSQL
func New(dsn string) (*Storage, error) {
	const op = "storage.postgres.New"

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Проверяем подключение
	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Создаем таблицу, если её нет
	stmt := `
	CREATE TABLE IF NOT EXISTS url (
		id SERIAL PRIMARY KEY,
		alias TEXT NOT NULL UNIQUE,
		url TEXT NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_alias ON url(alias);
	`
	_, err = db.Exec(stmt)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: db}, nil
}

// SaveURL сохраняет URL с заданным алиасом
func (s *Storage) SaveURL(urlToSave, alias string) error {
	const op = "storage.postgres.SaveURL"

	stmt := `INSERT INTO url (url, alias) VALUES ($1, $2)`
	_, err := s.db.Exec(stmt, urlToSave, alias)
	if err != nil {
		// Проверяем ошибку на нарушение уникальности
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" { // 23505 = unique_violation
			return fmt.Errorf("%s: %w", op, storage.ErrURLExists)
		}
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

// GetURL получает оригинальный URL по алиасу
func (s *Storage) GetURL(alias string) (string, error) {
	const op = "storage.postgres.GetURL"

	var url string
	stmt := `SELECT url FROM url WHERE alias = $1`
	err := s.db.QueryRow(stmt, alias).Scan(&url)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("%s: %w", op, storage.ErrURLNotFound)
		}
		return "", fmt.Errorf("%s: %w", op, err)
	}
	return url, nil
}

// Close закрывает соединение с БД
func (s *Storage) Close() error {
	return s.db.Close()
}
