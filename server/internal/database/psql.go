package psql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type Repository interface {
	GetConnection() *sqlx.DB
}

type repository struct {
	Client *sqlx.DB
}

var (
	EmptyInsert error = errors.New("no affected rows")
	EmptySelect error = errors.New("no select rows")
)

func New(dsn string) (Repository, error) {
	db, err := sqlx.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("Connect to DB error: %v", err)
	}

	repo := repository{Client: db}
	if err = repo.Client.Ping(); err != nil {
		return &repo, fmt.Errorf("DB is not ready error: %v", err)
	}

	return &repo, nil
}

func (c repository) GetConnection() *sqlx.DB {
	return c.Client
}

func Insert(ctx context.Context, conn *sqlx.DB, text, hash string) (int64, error) {
	row := conn.QueryRow("INSERT INTO hashes (text, hash) VALUES ( $1, $2) ON CONFLICT (text) DO NOTHING RETURNING id", text, hash)
	var i int64
	err := row.Scan(&i)
	if err != nil && err != sql.ErrNoRows {
		return 0, fmt.Errorf("Insert error: %v", err)
	}
	if err == sql.ErrNoRows {
		return 0, EmptyInsert
	}
	return i, nil
}

func Select(ctx context.Context, conn *sqlx.DB, id int64) (string, error) {
	row := conn.QueryRow("SELECT hash FROM hashes WHERE id=$1;", id)
	var hash string
	err := row.Scan(&hash)
	if err != nil && err != sql.ErrNoRows {
		return "", fmt.Errorf("Select error: %v", err)
	}
	if err == sql.ErrNoRows {
		return "", EmptySelect
	}
	return hash, nil
}
