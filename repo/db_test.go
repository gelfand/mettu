package repo

import (
	"testing"

	"github.com/ledgerwatch/erigon-lib/kv"
	"github.com/ledgerwatch/erigon-lib/kv/mdbx"
)

func newTestDBInMem() kv.RwDB {
	db, err := mdbx.NewMDBX(nil).WithTablessCfg(
		func(defaultBuckets kv.TableCfg) kv.TableCfg {
			return kvTablesCfg
		},
	).InMem().Open()
	if err != nil {
		panic(err)
	}

	return db
}

func newTestDB(t *testing.T) kv.RwDB {
	db, err := mdbx.NewMDBX(nil).Path(t.TempDir()).WithTablessCfg(
		func(defaultBuckets kv.TableCfg) kv.TableCfg {
			return kvTablesCfg
		}).Open()
	if err != nil {
		panic(err)
	}
	return db
}
