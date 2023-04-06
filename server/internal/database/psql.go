package psql

import (
	"context"
	"errors"
	"fmt"
	"hashserver/pkg/hashcalc"
	"sync"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type Instance struct {
	Conn       *pgxpool.Pool
	PoolConfig *pgxpool.Config
}

var (
	EmptyInsert error = errors.New("no affected rows")
	EmptySelect error = errors.New("no select rows")
)

func New(dsn string, number int) (*Instance, error) {
	poolConfig, err := pgxpool.ParseConfig(dsn)

	if err != nil {
		return nil, fmt.Errorf("Connect to DB error: %v", err)
	}
	poolConfig.MaxConns = int32(number)

	c, err := pgxpool.ConnectConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, err
	}

	_, err = c.Exec(context.Background(), ";")
	if err != nil {
		return nil, fmt.Errorf("Ping failed: %v\n", err)
	}

	repo := Instance{Conn: c, PoolConfig: poolConfig}

	return &repo, nil

}

func (i *Instance) Insert(ctx context.Context, hash string) (int64, error) {
	row := i.Conn.QueryRow(ctx, "WITH t AS (INSERT INTO hashes (hash) VALUES ($1) ON CONFLICT (hash) DO NOTHING RETURNING id) SELECT * FROM t UNION ALL SELECT id FROM hashes WHERE hash = $1;", hash)
	var id int64
	err := row.Scan(&id)
	if err != nil && err != pgx.ErrNoRows {
		return 0, fmt.Errorf("Insert error: %v", err)
	}
	if err == pgx.ErrNoRows {
		return id, EmptyInsert
	}
	return id, nil
}

func (i *Instance) Select(ctx context.Context, id int64) (string, error) {
	row := i.Conn.QueryRow(ctx, "SELECT hash FROM hashes WHERE id=$1;", id)
	var hash string
	err := row.Scan(&hash)
	if err != nil && err != pgx.ErrNoRows {
		return "", fmt.Errorf("Select error: %v", err)
	}
	if err == pgx.ErrNoRows {
		return "", EmptySelect
	}
	return hash, nil
}

func (db *Instance) MultiHashInsert(ctx context.Context, hash []*hashcalc.Hash) error {
	var wg sync.WaitGroup
	wg.Add(len(hash))
	HashIDs := make(chan *hashcalc.Hash, len(hash))
	for i := 0; i < len(hash); i++ {
		go inserted(ctx, &wg, db.Conn, hash[i].Hash, HashIDs)
	}
	wg.Wait()

	for i := 0; i < len(hash); i++ {
		hash[i] = <-HashIDs
	}
	wg.Wait()
	return nil
}

func inserted(ctx context.Context, wg *sync.WaitGroup, conn *pgxpool.Pool, hash string, hashID chan<- *hashcalc.Hash) {
	defer wg.Done()
	row := conn.QueryRow(ctx, "WITH t AS (INSERT INTO hashes (hash) VALUES ($1) ON CONFLICT (hash) DO NOTHING RETURNING id) SELECT * FROM t UNION ALL SELECT id FROM hashes WHERE hash = $1;", hash)
	var id int64
	err := row.Scan(&id)
	if err != nil {
		fmt.Errorf("Insert error: %v", err)
	}
	h := &hashcalc.Hash{
		Id:   id,
		Hash: hash,
	}
	hashID <- h
}
