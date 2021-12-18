package repo

import (
	"context"
	"fmt"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
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
			fmt.Println(got)
			if !cmp.Equal(got, tt.want, cmp.AllowUnexported(big.Int{})) {
				t.Errorf("DB.PeekPattern() = %v, want %v", got, tt.want)
			}
			roTx.Commit()
		})
	}
}

// func TestDB_PutPattern(t *testing.T) {
// 	type fields struct {
// 		d kv.RwDB
// 	}
// 	type args struct {
// 		tx kv.RwTx
// 		p  Pattern
// 	}
// 	tests := []struct {
// 		name    string
// 		fields  fields
// 		args    args
// 		wantErr bool
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			db := &DB{
// 				d: tt.fields.d,
// 			}
// 			if err := db.PutPattern(tt.args.tx, tt.args.p); (err != nil) != tt.wantErr {
// 				t.Errorf("DB.PutPattern() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 		})
// 	}
// }

// func TestDB_HasPattern(t *testing.T) {
// 	type fields struct {
// 		d kv.RwDB
// 	}
// 	type args struct {
// 		tx           kv.Tx
// 		token        common.Address
// 		exchangeName string
// 	}
// 	tests := []struct {
// 		name    string
// 		fields  fields
// 		args    args
// 		want    bool
// 		wantErr bool
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			db := &DB{
// 				d: tt.fields.d,
// 			}
// 			got, err := db.HasPattern(tt.args.tx, tt.args.token, tt.args.exchangeName)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("DB.HasPattern() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if got != tt.want {
// 				t.Errorf("DB.HasPattern() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

// func TestDB_PeekPattern(t *testing.T) {
// 	type fields struct {
// 		d kv.RwDB
// 	}
// 	type args struct {
// 		tx           kv.Tx
// 		token        common.Address
// 		exchangeName string
// 	}
// 	tests := []struct {
// 		name    string
// 		fields  fields
// 		args    args
// 		want    Pattern
// 		wantErr bool
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			db := &DB{
// 				d: tt.fields.d,
// 			}
// 			got, err := db.PeekPattern(tt.args.tx, tt.args.token, tt.args.exchangeName)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("DB.PeekPattern() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("DB.PeekPattern() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

func mustPutPatterns(db *DB, patterns []Pattern) {
	tx, _ := db.BeginRw(context.Background())
	for i := range patterns {
		db.PutPattern(tx, patterns[i])
	}
	tx.Commit()
}

func TestDB_AllPatterns(t *testing.T) {
	bigNumber := big.NewInt(0)
	bigNumber, _ = bigNumber.SetString("115792089237316195423570985008687907853269984665640564039457584007913129639936", 10)
	hugeNumber := new(big.Int).Mul(bigNumber, big.NewInt(1<<20))

	type fields struct {
		d kv.RwDB
	}
	type args struct {
		tx kv.Tx
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []Pattern
		wantErr bool
	}{
		{
			name:   "test0",
			fields: fields{newTestDB(t)},
			args:   args{},
			want: []Pattern{
				{TokenAddr: common.BytesToAddress([]byte("pattern0")), ExchangeName: "FTX", Value: hugeNumber, TimesOccured: ^int(0)},
				{TokenAddr: common.BytesToAddress([]byte("pattern1")), ExchangeName: "Binance", Value: hugeNumber, TimesOccured: ^int(0)},
				{TokenAddr: common.BytesToAddress([]byte("pattern2")), ExchangeName: "Binance", Value: hugeNumber, TimesOccured: ^int(0)},
				{TokenAddr: common.BytesToAddress([]byte("pattern3")), ExchangeName: "Binance", Value: hugeNumber, TimesOccured: ^int(0)},
				{TokenAddr: common.BytesToAddress([]byte("pattern4")), ExchangeName: "Binance", Value: hugeNumber, TimesOccured: ^int(0)},
				{TokenAddr: common.BytesToAddress([]byte("pattern5")), ExchangeName: "Binance", Value: hugeNumber, TimesOccured: ^int(0)},
				{TokenAddr: common.BytesToAddress([]byte("pattern6")), ExchangeName: "Binance", Value: hugeNumber, TimesOccured: ^int(0)},
				{TokenAddr: common.BytesToAddress([]byte("pattern7")), ExchangeName: "Binance", Value: hugeNumber, TimesOccured: ^int(0)},
				{TokenAddr: common.BytesToAddress([]byte("pattern8")), ExchangeName: "Binance", Value: hugeNumber, TimesOccured: ^int(0)},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &DB{
				d: tt.fields.d,
			}

			mustPutPatterns(db, tt.want)
			tt.args.tx, _ = db.BeginRo(context.Background())
			got, err := db.AllPatterns(tt.args.tx)
			if (err != nil) != tt.wantErr {
				t.Errorf("DB.AllPatterns() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !cmp.Equal(got, tt.want, cmpopts.SortSlices(func(x, y Pattern) bool {
				xBig, yBig := new(big.Int).SetBytes(x.TokenAddr[:]), new(big.Int).SetBytes(y.TokenAddr[:])
				return xBig.Cmp(yBig) == -1
			}), cmp.AllowUnexported(big.Int{})) {
				t.Errorf("DB.AllPatterns() = %v, want %v", got, tt.want)
			}
		})
	}
}
