package repo

import (
	"github.com/ledgerwatch/erigon-lib/kv"
	"github.com/ledgerwatch/erigon-lib/kv/mdbx"
)

func newTestDB() kv.RwDB {
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
