package repo

import (
	"context"
	"math/big"
	"reflect"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ledgerwatch/erigon-lib/kv"
)

func TestDB_Pattern(t *testing.T) {
	bigNumber := big.NewInt(0)
	bigNumber, _ = bigNumber.SetString("115792089237316195423570985008687907853269984665640564039457584007913129639936", 10)
	hugeNumber := new(big.Int).Mul(bigNumber, big.NewInt(1<<20))

	type fields struct {
		d kv.RwDB
	}
	type args struct {
		tx kv.RwTx
		p  Pattern
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		want    Pattern
	}{
		{
			name:   "test0",
			fields: fields{newTestDB(t)},
			args: args{
				tx: nil,
				p: Pattern{
					TokenAddr:    common.BytesToAddress([]byte("its a token")),
					ExchangeName: "exchange",
					Value:        hugeNumber,
					TimesOccured: ^int(0),
				},
			},
			want: Pattern{
				TokenAddr:    common.BytesToAddress([]byte("its a token")),
				ExchangeName: "exchange",
				Value:        hugeNumber,
				TimesOccured: ^int(0),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &DB{
				d: tt.fields.d,
			}
			tx, err := db.BeginRw(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			if err := db.PutPattern(tx, tt.args.p); err != nil {
				t.Errorf("DB.PutPattern() error = %v, wantErr %v", err, tt.wantErr)
			}
			tx.Commit()

			roTx, err := db.BeginRo(context.Background())
			if err != nil {
				t.Fatal(err)
			}

			gotHas, err := db.HasPattern(roTx, tt.args.p.TokenAddr, tt.args.p.ExchangeName)
			if err != nil {
				t.Fatal()
			}
			if !gotHas {
				t.Errorf("DB.HasPattern() = %v, want %v", gotHas, true)
			}

			got, err := db.PeekPattern(roTx, tt.args.p.TokenAddr, tt.args.p.ExchangeName)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DB.PeekPattern() = %v, want %v", got, tt.want)
			}
			roTx.Commit()
		})
	}
}
