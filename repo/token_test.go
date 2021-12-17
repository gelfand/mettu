package repo

import (
	"context"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/google/go-cmp/cmp"
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
	t.Parallel()

	bigNumber := big.NewInt(0)
	bigNumber, _ = bigNumber.SetString("115792089237316195423570985008687907853269984665640564039457584007913129639936", 10)
	hugeNumber := new(big.Int).Mul(bigNumber, big.NewInt(128))

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
			fields: fields{newTestDB(t)},
			args: args{
				addr: common.HexToAddress("0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2"),
			},
			want: Token{
				Address:     common.HexToAddress("0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2"),
				Symbol:      "WETH",
				Decimals:    18,
				TimesBought: 25,
				TotalBought: hugeNumber,
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

			if !cmp.Equal(got, tt.want, cmp.AllowUnexported(big.Int{})) {
				t.Errorf("DB.PeekToken() = %v, want %v", got, tt.want)
			}
		})
	}
}
