package repo

import (
	"context"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/ledgerwatch/erigon-lib/kv"
	"github.com/ledgerwatch/erigon-lib/kv/mdbx"
)

type DB struct {
	d kv.RwDB
}

const (
	exchangeStorage = "ExchangeStorage"
	patternStorage  = "PatternStorage"
	tokenStorage    = "TokenStorage"
	accountStorage  = "AccountStorage"
	swapStorage     = "SwapStorage"
)

var kvTables = []string{
	accountStorage,
	exchangeStorage,
	patternStorage,
	tokenStorage,
	swapStorage,
}

var kvTablesCfg = kv.TableCfg{
	accountStorage:  kv.TableCfgItem{},
	exchangeStorage: kv.TableCfgItem{},
	patternStorage:  kv.TableCfgItem{},
	tokenStorage:    kv.TableCfgItem{},
	swapStorage:     kv.TableCfgItem{},
}

func NewDB(path string) (*DB, error) {
	db, err := mdbx.NewMDBX(nil).Path(path).WithTablessCfg(
		func(defaultBuckets kv.TableCfg) kv.TableCfg {
			return kvTablesCfg
		}).Open()
	if err != nil {
		return nil, err
	}
	return &DB{db}, nil
}

func NewDBReadOnly(path string) (*DB, error) {
	db, err := mdbx.NewMDBX(nil).Path(path).WithTablessCfg(
		func(defaultBuckets kv.TableCfg) kv.TableCfg {
			return kvTablesCfg
		}).Readonly().Open()
	if err != nil {
		return nil, err
	}
	return &DB{db}, nil
}

// BeginRo begins read-only transaction.
func (db *DB) BeginRo(ctx context.Context) (kv.Tx, error) {
	return db.d.BeginRo(ctx)
}

// BeginRw begins read-write transaction.
func (db *DB) BeginRw(ctx context.Context) (kv.RwTx, error) {
	return db.d.BeginRw(ctx)
}

// Close closes the DB.
func (db *DB) Close() {
	db.d.Close()
}

// Update starts read-write transaction, for doing many-things.
func (db *DB) Update(ctx context.Context, f func(tx kv.RwTx) error) (err error) {
	return db.d.Update(ctx, f)
}

// View starts read-only transaction, for doing many-things.
func (db *DB) View(ctx context.Context, f func(tx kv.Tx) error) (err error) {
	return db.d.View(ctx, f)
}

func (db *DB) FlushBucket(tx kv.RwTx, table string) error {
	return tx.ClearBucket(table)
}

// func (db *DB) InsertToken(ctx context.Context) error {
// 	tx, err := db.d.BeginTx(ctx, &sql.TxOptions{ReadOnly: false})
// 	if err != nil {
// 		return err
// 	}
// 	defer tx.Rollback()

// 	if _, err = tx.ExecContext(ctx, ""); err != nil {
// 		return fmt.Errorf("unable to execute query")
// 	}

// 	return tx.Commit()
// }

// func pgxConnPool() k
