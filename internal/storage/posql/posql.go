package posql

import (
	"database/sql"
	"fmt"

	"url-shorter-REST-API/internal/storage"

	"github.com/lib/pq"
)

type Storage struct {
	db *sql.DB
}

// Подключаюсь к посгру
func New(dsn string) (*Storage, error) {
	const op = "storage.postgres.New"

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Проверка коннекта
	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Если нет таблицы, делаем
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
func (s *Storage) SaveURL(urlToSave, alias string) (int64, error) {
	const op = "storage.postgres.SaveURL"

	stmt := `INSERT INTO url (url, alias) VALUES ($1, $2)`
	_, err := s.db.Exec(stmt, urlToSave, alias)
	if err != nil {
		// проверка уникальности
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" { // 23505 = unique_violation
			return 0, fmt.Errorf("%s: %w", op, storage.ErrURLExists)
		}
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	return 0, nil
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

// func (s *Storage) DeleteURL(alias string) error {
// 	const op = "storage.postgres.DeleteURL"

// 	stmt := "DELETE FROM url WHERE alias = $1"
// 	res, err := s.db.Exec(stmt, alias)
// 	if err != nil {
// 		return fmt.Errorf("%s: %w", op, err)
// 	}

// 	rowsAffected, err := res.RowsAffected()
// 	if err != nil {
// 		return fmt.Errorf("%s: %w", op, err)
// 	}

// 	if rowsAffected == 0 {
// 		return fmt.Errorf("%s: %w", op, storage.ErrURLNotFound)
// 	}

// 	return nil
// }
