package repo

import (
	"context"
	"fmt"
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

func Test_Has(t *testing.T) {
	db := newTestDBInMem()
	tx, err := db.BeginRw(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	tx.CreateBucket("bucket")
	tx.Commit()

	roTx, err := db.BeginRo(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(roTx.Has("bucket", []byte("key")))
}
