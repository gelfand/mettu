package repo

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ledgerwatch/erigon-lib/kv"
)

func testDB_PutToken(t *testing.T, db *DB, token Token, wantErr bool) {
	tx, err := db.BeginRw(context.Background())
	if err != nil {
		t.Errorf("error = %v, wantErr %v", err, wantErr)
	}
	defer tx.Rollback()

	if err = db.PutToken(tx, token); (err != nil) != wantErr {
		t.Errorf("DB.PutToken() error = %v, wantErr %v", err, wantErr)
	}
	if err = tx.Commit(); err != nil {
		t.Error(err)
	}
}

func TestDB_PutPeekToken(t *testing.T) {
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
		want    Token
		wantErr bool
	}{
		{
			name:   "",
			fields: fields{newTestDB()},
			args: args{
				addr: common.HexToAddress("0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2"),
			},
			want: Token{
				Symbol:    "WETH",
				Decimals:  18,
				Purchases: 255,
				Address:   common.HexToAddress("0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2"),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &DB{
				d: tt.fields.d,
			}

			testDB_PutToken(t, db, tt.want, false)

			var err error
			tt.args.tx, err = db.BeginRo(context.Background())
			if err != nil {
				t.Error(err)
			}
			defer tt.args.tx.Rollback()

			got, err := db.PeekToken(tt.args.tx, tt.args.addr)
			if (err != nil) != tt.wantErr {
				t.Errorf("DB.PeekToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			fmt.Println(got)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DB.PeekToken() = %v, want %v", got, tt.want)
			}
		})
	}
}
