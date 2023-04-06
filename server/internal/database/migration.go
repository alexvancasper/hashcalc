package psql

import (
	"database/sql"

	_ "github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/pressly/goose"
)

func MigrationUP(db *Instance) error {
	mdb, err := sql.Open("postgres", db.PoolConfig.ConnString())
	if err != nil {
		return errors.WithStack(err)
	}
	err = mdb.Ping()
	if err != nil {
		return errors.WithStack(err)
	}
	if err := goose.Up(mdb, "migrations"); err != nil {
		return errors.WithStack(err)
	}
	mdb.Close()
	return nil
}
