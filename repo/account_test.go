package repo

import (
	"context"
	"math/big"
	"reflect"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ledgerwatch/erigon-lib/kv"
)

func TestDB_PutAccount(t *testing.T) {
	t.Parallel()

	bigNumber := big.NewInt(0)
	bigNumber, _ = bigNumber.SetString("115792089237316195423570985008687907853269984665640564039457584007913129639936", 10)
	hugeNumber := new(big.Int).Mul(bigNumber, big.NewInt(1<<20))

	exchanges := make(map[string]bool)
	for _, v := range testdata_mapOfExchanges() {
		exchanges[v.Name] = true
	}

	type fields struct {
		d kv.RwDB
	}
	type args struct {
		tx  kv.RwTx
		acc Account
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "test0",
			fields: fields{newTestDB(t)},
			args: args{
				tx: nil,
				acc: Account{
					Address:       common.BytesToAddress([]byte("acc0")),
					TotalReceived: hugeNumber,
					TotalSpent:    hugeNumber,
					FromExchanges: exchanges,
				},
			},
			wantErr: false,
		},
		{
			name:   "test1",
			fields: fields{newTestDB(t)},
			args: args{
				tx: nil,
				acc: Account{
					Address:       common.BytesToAddress([]byte("qwerty")),
					TotalReceived: hugeNumber,
					TotalSpent:    hugeNumber,
					FromExchanges: exchanges,
				},
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
			tx, err := db.BeginRw(context.Background())
			if err != nil {
				t.Errorf("db.BeginRw(): %v", err)
			}

			if err := db.PutAccount(tx, tt.args.acc); (err != nil) != tt.wantErr {
				t.Errorf("DB.PutAccount() error = %v, wantErr %v", err, tt.wantErr)
			}
			tx.Commit()

			roTx, err := db.BeginRo(context.Background())
			if err != nil {
				t.Errorf("db.BeginRo(); %v", err)
			}
			defer roTx.Rollback()

			got1, err := db.HasAccount(roTx, tt.args.acc.Address)
			if err != nil {
				t.Errorf("DB.HasAccount() error = %v", err)
			}
			if !got1 {
				t.Errorf("DB.HasAccount(), got = %v, want = %v", false, true)
			}

			got2, err := db.PeekAccount(roTx, tt.args.acc.Address)
			if err != nil {
				t.Errorf("DB.PeekAccount() error = %v", err)
			}
			if !reflect.DeepEqual(got2, tt.args.acc) {
				t.Errorf("DB.PutAccount(), got = %v, want = %v", got2, tt.args.acc)
			}
		})
	}
}

// func TestDB_HasAccount(t *testing.T) {
// 	type fields struct {
// 		d kv.RwDB
// 	}
// 	type args struct {
// 		tx   kv.Tx
// 		addr common.Address
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
// 			got, err := db.HasAccount(tt.args.tx, tt.args.addr)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("DB.HasAccount() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if got != tt.want {
// 				t.Errorf("DB.HasAccount() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

// func TestDB_PeekAccount(t *testing.T) {
// 	type fields struct {
// 		d kv.RwDB
// 	}
// 	type args struct {
// 		tx      kv.Tx
// 		address common.Address
// 	}
// 	tests := []struct {
// 		name    string
// 		fields  fields
// 		args    args
// 		want    Account
// 		wantErr bool
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			db := &DB{
// 				d: tt.fields.d,
// 			}
// 			got, err := db.PeekAccount(tt.args.tx, tt.args.address)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("DB.PeekAccount() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("DB.PeekAccount() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

// func TestDB_AllAccounts(t *testing.T) {
// 	type fields struct {
// 		d kv.RwDB
// 	}
// 	type args struct {
// 		tx kv.Tx
// 	}
// 	tests := []struct {
// 		name    string
// 		fields  fields
// 		args    args
// 		want    []Account
// 		wantErr bool
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			db := &DB{
// 				d: tt.fields.d,
// 			}
// 			got, err := db.AllAccounts(tt.args.tx)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("DB.AllAccounts() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("DB.AllAccounts() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

// func TestDB_AllAccountsMap(t *testing.T) {
// 	type fields struct {
// 		d kv.RwDB
// 	}
// 	type args struct {
// 		tx kv.Tx
// 	}
// 	tests := []struct {
// 		name    string
// 		fields  fields
// 		args    args
// 		want    map[common.Address]Account
// 		wantErr bool
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			db := &DB{
// 				d: tt.fields.d,
// 			}
// 			got, err := db.AllAccountsMap(tt.args.tx)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("DB.AllAccountsMap() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("DB.AllAccountsMap() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }
