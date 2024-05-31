package data

import (
	"database/sql"
	"fmt"
)

func New(dsn string) (*OrderStorage, error) {
	const op = "data.sqlite.New"

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &OrderStorage{DB: db}, nil
}
