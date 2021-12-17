package repo

import (
	"context"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
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
			if !cmp.Equal(got2, tt.args.acc, cmp.AllowUnexported(big.Int{})) {
				t.Errorf("DB.PutAccount(), got = %v, want = %v", got2, tt.args.acc)
			}
		})
	}
}

func TestDB_HasAccount(t *testing.T) {
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
		want    bool
		wantErr bool
	}{
		{
			name: "test0",
			fields: fields{
				d: newTestDB(t),
			},
			args: args{
				tx:   nil,
				addr: common.BytesToAddress([]byte("has account test0")),
			},
			want:    true,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &DB{
				d: tt.fields.d,
			}
			done := make(chan bool, 1)
			go func() {
				tx, _ := db.BeginRw(context.Background())
				db.PutAccount(tx, Account{
					Address:       common.BytesToAddress([]byte("has account test0")),
					TotalReceived: &big.Int{},
					TotalSpent:    &big.Int{},
					FromExchanges: map[string]bool{},
				})
				tx.Commit()
				done <- true
			}()
			<-done

			tt.args.tx, _ = db.BeginRo(context.Background())
			got, err := db.HasAccount(tt.args.tx, tt.args.addr)
			if (err != nil) != tt.wantErr {
				t.Errorf("DB.HasAccount() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("DB.HasAccount() = %v, want %v", got, tt.want)
			}
		})
	}
}

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

func TestDB_AllAccountsMap(t *testing.T) {
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
		want    map[common.Address]Account
		wantErr bool
	}{
		{
			name:   "test0",
			fields: fields{d: newTestDB(t)},
			args:   args{},
			want: map[common.Address]Account{
				{0xff}: {
					Address:       common.Address{0xff},
					TotalReceived: big.NewInt(0),
					TotalSpent:    big.NewInt(0),
					FromExchanges: map[string]bool{"": true},
				},
				{0xfe, 0xfe}: {
					Address:       common.Address{0xfe, 0xfe},
					TotalReceived: big.NewInt(0),
					TotalSpent:    big.NewInt(0),
					FromExchanges: map[string]bool{"": true},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
		})
		db := &DB{
			d: tt.fields.d,
		}

		accountsTestdata(db)
		tt.args.tx, _ = db.BeginRo(context.Background())
		got, err := db.AllAccountsMap(tt.args.tx)
		if (err != nil) != tt.wantErr {
			t.Errorf("DB.AllAccountsMap() error = %v, wantErr %v", err, tt.wantErr)
			return
		}

		if !cmp.Equal(tt.want, got, cmpopts.SortMaps(func(x, y common.Address) bool {
			xBig, yBig := new(big.Int).SetBytes(x[:]), new(big.Int).SetBytes(y[:])
			return xBig.Cmp(yBig) == -1
		}), cmp.AllowUnexported(big.Int{})) {
			t.Errorf("DB.AllAccountsMap() got = %v, want = %v", got, tt.want)
		}
	}
}

func accountsTestdata(db *DB) {
	tx, err := db.BeginRw(context.Background())
	if err != nil {
		panic(err)
	}
	defer tx.Rollback()

	db.PutAccount(tx, Account{
		Address:       common.Address{0xff},
		TotalReceived: big.NewInt(0),
		TotalSpent:    big.NewInt(0),
		FromExchanges: map[string]bool{"": true},
	})

	db.PutAccount(tx, Account{
		Address:       common.Address{0xfe, 0xfe},
		TotalReceived: big.NewInt(0),
		TotalSpent:    big.NewInt(0),
		FromExchanges: map[string]bool{"": true},
	})
	if err := tx.Commit(); err != nil {
		panic(err)
	}
}

func must(thing func() (interface{}, interface{})) {
	if _, err := thing(); err != nil {
		panic(err)
	}
}
