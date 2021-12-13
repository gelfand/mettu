package repo

import (
	"context"
	"reflect"
	"testing"

	"github.com/ethereum/go-ethereum/common"
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
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DB.PeekExchange() = %v, want %v", got, tt.want)
			}

			tt.args.tx.Commit()
		})
	}
}
