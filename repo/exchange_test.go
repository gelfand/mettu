package repo

import (
	"context"
	"encoding/json"
	"math/big"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/ledgerwatch/erigon-lib/kv"
)

func testDB_PutExchange(t *testing.T, db *DB, e Exchange, wantErr bool) {
	tx, err := db.BeginRw(context.Background())
	if err != nil {
		t.Errorf("error = %v, wantErr %v", err, wantErr)
	}
	defer tx.Rollback()

	if err = db.PutExchange(tx, e); (err != nil) != wantErr {
		t.Errorf("DB.PutExchange() error = %v, wantErr %v", err, wantErr)
	}
	if err = tx.Commit(); err != nil {
		t.Error(err)
	}
}

func testDB_PutManyExchanges(t *testing.T, db *DB, exchanges []Exchange) {
	tx, err := db.BeginRw(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()

	for _, e := range exchanges {
		if err = db.PutExchange(tx, e); err != nil {
			t.Errorf("DB.PutExchange() error = %v", err)
		}
	}

	tx.Commit()
}

func TestDB_PutPeekExchange(t *testing.T) {
	t.Parallel()
	type fields struct {
		d kv.RwDB
	}
	type args struct {
		tx   kv.Tx
		addr common.Address
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    Exchange
		wantErr bool
	}{
		{
			name:   "test0",
			fields: fields{newTestDB(t)},
			args:   args{addr: common.BytesToAddress([]byte("binance0"))},
			want: Exchange{
				Name:    "Binance0",
				Address: common.BytesToAddress([]byte("binance0")),
			},
			wantErr: false,
		},
		{
			name:   "test1",
			fields: fields{newTestDB(t)},
			args:   args{addr: common.BytesToAddress([]byte("binance1"))},
			want: Exchange{
				Name:    "Binance1",
				Address: common.BytesToAddress([]byte("binance1")),
			},
			wantErr: false,
		},
		{
			name:   "test2",
			fields: fields{newTestDB(t)},
			args:   args{addr: common.BytesToAddress([]byte("binance2"))},
			want: Exchange{
				Name:    "Binance2",
				Address: common.BytesToAddress([]byte("binance2")),
			},
			wantErr: false,
		},
		{
			name:   "test3",
			fields: fields{newTestDB(t)},
			args:   args{addr: common.BytesToAddress([]byte("binance3"))},
			want: Exchange{
				Name:    "Binance3",
				Address: common.BytesToAddress([]byte("binance3")),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db := &DB{
				d: tt.fields.d,
			}

			testDB_PutExchange(t, db, tt.want, false)

			var err error
			tt.args.tx, err = db.BeginRo(context.Background())
			if err != nil {
				t.Error(err)
			}
			defer tt.args.tx.Rollback()

			got, err := db.PeekExchange(tt.args.tx, tt.args.addr)
			if (err != nil) != tt.wantErr {
				t.Errorf("DB.PeekExchange() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !cmp.Equal(got, tt.want) {
				t.Errorf("DB.PeekExchange() = %v, want %v", got, tt.want)
			}

			tt.args.tx.Commit()
		})
	}
}

func testdata_sliceOfExchanges() []Exchange {
	f, err := os.Open("./testdata/exchanges.json")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	var exchanges []Exchange
	dec := json.NewDecoder(f)
	if err = dec.Decode(&exchanges); err != nil {
		panic(err)
	}

	return exchanges
}

func testdata_mapOfExchanges() map[common.Address]Exchange {
	exchanges := make(map[common.Address]Exchange)
	exch := testdata_sliceOfExchanges()
	for i := range exch {
		exchanges[exch[i].Address] = exch[i]
	}
	return exchanges
}

func TestDB_AllExchanges(t *testing.T) {
	type fields struct {
		d kv.RwDB
	}
	type args struct {
		tx kv.RwTx
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []Exchange
		wantErr bool
	}{
		{
			name:    "allExchanges",
			fields:  fields{newTestDB(t)},
			want:    testdata_sliceOfExchanges(),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &DB{
				d: tt.fields.d,
			}

			testDB_PutManyExchanges(t, db, tt.want)

			tx, err := db.BeginRo(context.Background())
			if err != nil {
				t.Error(err)
			}
			defer tx.Rollback()

			got, err := db.AllExchanges(tx)
			if (err != nil) != tt.wantErr {
				t.Errorf("DB.AllExchanges() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !cmp.Equal(got, tt.want, cmpopts.SortSlices(func(x, y Exchange) bool {
				xBig, yBig := new(big.Int).SetBytes(x.Address[:]), new(big.Int).SetBytes(y.Address[:])
				return xBig.Cmp(yBig) >= 0
			})) {
				t.Errorf("DB.AllExchanges() = %v, want %v", got, tt.want)
			}
			tx.Commit()
		})
	}
}

func TestDB_AllExchangesMap(t *testing.T) {
	type fields struct {
		d kv.RwDB
	}
	type args struct {
		tx kv.RwTx
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    map[common.Address]Exchange
		wantErr bool
	}{
		{
			name:   "test0",
			fields: fields{newTestDB(t)},
			want:   testdata_mapOfExchanges(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &DB{
				d: tt.fields.d,
			}

			testDB_PutManyExchanges(t, db, testdata_sliceOfExchanges())

			tx, err := db.BeginRo(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			defer tx.Rollback()

			got, err := db.AllExchangesMap(tx)
			if (err != nil) != tt.wantErr {
				t.Errorf("DB.AllExchangesMap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !cmp.Equal(got, tt.want, cmpopts.SortMaps(func(x, y common.Address) bool {
				xBig, yBig := new(big.Int).SetBytes(x[:]), new(big.Int).SetBytes(y[:])
				return xBig.Cmp(yBig) >= 0
			})) {
				t.Errorf("DB.AllExchangesMap() = %v, want %v", got, tt.want)
			}
			tx.Commit()
		})
	}
}
