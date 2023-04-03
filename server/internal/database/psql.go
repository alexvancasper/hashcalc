package psql

import (
	"context"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type Repository interface {
	GetConnection() *sqlx.DB
}

type repository struct {
	Client *sqlx.DB
}

func New(dsn string) (Repository, error) {
	db, err := sqlx.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	repo := repository{Client: db}
	if err = repo.Client.Ping(); err != nil {
		return &repo, err
	}

	return &repo, nil
}

func (c repository) GetConnection() *sqlx.DB {
	return c.Client
}

func Insert(ctx context.Context, conn *sqlx.DB, text, hash string) (int64, error) {
	row := conn.QueryRow("INSERT INTO hashes (text, hash) VALUES ( $1, $2) RETURNING id;", text, hash)
	var i int64
	err := row.Scan(&i)
	if err != nil {
		return 0, err
	}
	return i, nil
}

func Select(ctx context.Context, conn *sqlx.DB, id string) (string, error) {
	row := conn.QueryRow("SELECT hash FROM hashes WHERE id=$1;", id)
	var hash string
	err := row.Scan(&hash)
	if err != nil {
		return "", err
	}
	return hash, nil
}
