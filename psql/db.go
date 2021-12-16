package psql

import (
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
)

type DB struct {
	d *sqlx.DB
}

func NewDB(databaseURI string) (*DB, error) {
	d, err := sqlx.Open("pgx", databaseURI)
	if err != nil {
		return nil, err
	}
	return &DB{d}, nil
}
