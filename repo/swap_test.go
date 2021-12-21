package repo

import (
	"context"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/google/go-cmp/cmp"
	"github.com/ledgerwatch/erigon-lib/kv"
)

func TestDB_PutSwap(t *testing.T) {
	t.Parallel()

	type fields struct {
		d kv.RwDB
	}
	type args struct {
		tx kv.RwTx
		s  Swap
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
				s: Swap{
					TxHash:    common.BytesToHash([]byte("tx")),
					Wallet:    common.BytesToAddress([]byte("address")),
					TokenAddr: common.BytesToAddress([]byte("token")),
					Price:     &big.Int{},
					Value:     &big.Int{},
				},
			},
			wantErr: false,
		},
		{
			name:   "test1",
			fields: fields{newTestDB(t)},
			args: args{
				tx: nil,
				s: Swap{
					TxHash:    common.BytesToHash([]byte("transaction hash")),
					Wallet:    common.BytesToAddress([]byte("wallet address")),
					TokenAddr: common.BytesToAddress([]byte("token address")),
					Price:     &big.Int{},
					Value:     &big.Int{},
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
			tt.args.tx, _ = db.BeginRw(context.Background())
			if err := db.PutSwap(tt.args.tx, tt.args.s); (err != nil) != tt.wantErr {
				t.Errorf("DB.PutSwap() error = %v, wantErr %v", err, tt.wantErr)
			}
			tt.args.tx.Commit()

			roTx, err := db.BeginRo(context.Background())
			if err != nil {
				t.Fatal(err)
			}

			got, err := db.PeekSwap(roTx, tt.args.s.TxHash)
			if err != nil {
				t.Fatal(err)
			}
			roTx.Commit()

			if !cmp.Equal(got, tt.args.s, cmp.AllowUnexported(big.Int{})) {
				t.Errorf("DB.PeekSwap() got = %v, want = %v", got, tt.args.s)
			}
		})
	}
}

func TestDB_PeekSwap(t *testing.T) {
	type fields struct {
		d kv.RwDB
	}
	type args struct {
		tx     kv.Tx
		txHash common.Hash
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    Swap
		wantErr bool
	}{
		{
			name:   "test0",
			fields: fields{testDB_swap(t)},
			args: args{
				tx:     nil,
				txHash: common.BytesToHash([]byte("tx0")),
			},
			want: Swap{
				TxHash:    common.BytesToHash([]byte("tx0")),
				Wallet:    common.BytesToAddress([]byte("wallet1")),
				TokenAddr: common.BytesToAddress([]byte("token")),
				Price:     &big.Int{},
				Value:     &big.Int{},
			},

			wantErr: false,
		},
		{
			name:   "test1",
			fields: fields{testDB_swap(t)},
			args: args{
				tx:     nil,
				txHash: common.BytesToHash([]byte("tx1")),
			},
			want: Swap{
				TxHash:    common.BytesToHash([]byte("tx1")),
				Wallet:    common.BytesToAddress([]byte("wallet1")),
				TokenAddr: common.BytesToAddress([]byte("token1")),
				Price:     &big.Int{},
				Value:     &big.Int{},
			},
			wantErr: false,
		},
		{
			name:   "test2",
			fields: fields{testDB_swap(t)},
			args: args{
				tx:     nil,
				txHash: common.BytesToHash([]byte("tx2")),
			},
			want: Swap{
				TxHash:    common.BytesToHash([]byte("tx2")),
				Wallet:    common.BytesToAddress([]byte("wallet2")),
				TokenAddr: common.BytesToAddress([]byte("token1")),
				Price:     &big.Int{},
				Value:     &big.Int{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &DB{
				d: tt.fields.d,
			}
			var err error
			tt.args.tx, err = db.BeginRo(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			defer tt.args.tx.Rollback()

			got, err := db.PeekSwap(tt.args.tx, tt.args.txHash)
			if (err != nil) != tt.wantErr {
				t.Errorf("DB.PeekSwap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !cmp.Equal(got, tt.want, cmp.AllowUnexported(big.Int{})) {
				t.Errorf("DB.PeekSwap() = %v, want %v", got, tt.want)
			}
		})
	}
}

// func TestDB_AllSwapsMap(t *testing.T) {
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
// 		want    map[common.Address][]SwapWithToken
// 		wantErr bool
// 	}{
// 		{
// 			name:   "test0",
// 			fields: fields{testDB_swap(t)},
// 			args:   args{},
// 			want: map[common.Address][]SwapWithToken{
// 				common.BytesToAddress([]byte("wallet1")): {
// 					{
// 						TxHash: common.BytesToHash([]byte("tx0")),
// 						Wallet: Account{
// 							Address:       common.BytesToAddress([]byte("wallet1")),
// 							TotalReceived: &big.Int{},
// 							TotalSpent:    &big.Int{},
// 							Exchange:      "",
// 						},
// 						Token: Token{
// 							Address:     common.BytesToAddress([]byte("token")),
// 							Symbol:      "",
// 							Decimals:    0,
// 							TotalBought: &big.Int{},
// 							TimesBought: 0,
// 						},
// 						Price: &big.Int{},
// 						Value: &big.Int{},
// 					},
// 					{
// 						TxHash: common.BytesToHash([]byte("tx1")),
// 						Wallet: Account{
// 							Address:       common.BytesToAddress([]byte("wallet1")),
// 							TotalReceived: &big.Int{},
// 							TotalSpent:    &big.Int{},
// 							Exchange:      "",
// 						},
// 						Token: Token{
// 							Address:     common.BytesToAddress([]byte("token1")),
// 							Symbol:      "",
// 							Decimals:    0,
// 							TotalBought: &big.Int{},
// 							TimesBought: 0,
// 						},
// 						Price: &big.Int{},
// 						Value: &big.Int{},
// 					},
// 				},
// 				common.BytesToAddress([]byte("wallet2")): {{
// 					TxHash: common.BytesToHash([]byte("tx2")),
// 					Wallet: Account{
// 						Address:       common.BytesToAddress([]byte("wallet2")),
// 						TotalReceived: &big.Int{},
// 						TotalSpent:    &big.Int{},
// 						Exchange:      "",
// 					},
// 					Token: Token{
// 						Address:     common.BytesToAddress([]byte("token1")),
// 						Symbol:      "",
// 						Decimals:    0,
// 						TotalBought: &big.Int{},
// 						TimesBought: 0,
// 					},
// 					Price: &big.Int{},
// 					Value: &big.Int{},
// 				}},
// 			},
// 			wantErr: false,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			db := &DB{
// 				d: tt.fields.d,
// 			}
// 			var err error
// 			tt.args.tx, err = db.BeginRo(context.Background())
// 			if err != nil {
// 				t.Fatal(err)
// 			}
// 			got, err := db.AllSwapsMap(tt.args.tx)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("DB.AllSwapsMap() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}

// 			if !cmp.Equal(got, tt.want, cmpopts.SortMaps(func(x, y common.Address) bool {
// 				xBig, yBig := new(big.Int).SetBytes(x[:]), new(big.Int).SetBytes(y[:])
// 				return xBig.Cmp(yBig) > 0
// 			}), cmp.AllowUnexported(big.Int{})) {
// 				t.Errorf("DB.AllSwapsMap() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

func testDB_swap(t *testing.T) kv.RwDB {
	d := newTestDB(t)
	db := &DB{
		d: d,
	}

	tx, err := db.BeginRw(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	type args struct {
		s Swap
	}

	tokens := []Token{
		{
			Address:     common.BytesToAddress([]byte("token")),
			Symbol:      "",
			Decimals:    0,
			Price:       big.NewInt(1e18),
			TotalBought: &big.Int{},
			TimesBought: 0,
		},
		{
			Address:     common.BytesToAddress([]byte("token1")),
			Price:       big.NewInt(1e18),
			Symbol:      "",
			Decimals:    0,
			TotalBought: &big.Int{},
			TimesBought: 0,
		},
	}

	accounts := []Account{
		{
			Address:       common.BytesToAddress([]byte("wallet1")),
			TotalReceived: &big.Int{},
			TotalSpent:    &big.Int{},
			Exchange:      "",
		},
		{
			Address:       common.BytesToAddress([]byte("wallet2")),
			TotalReceived: &big.Int{},
			TotalSpent:    &big.Int{},
			Exchange:      "",
		},
	}

	swaps := []Swap{
		{
			TxHash:    common.BytesToHash([]byte("tx0")),
			Wallet:    common.BytesToAddress([]byte("wallet1")),
			TokenAddr: common.BytesToAddress([]byte("token")),
			Path:      []common.Address{},
			Factory:   [20]byte{},
			Price:     &big.Int{},
			Value:     &big.Int{},
		},
		{
			TxHash:    common.BytesToHash([]byte("tx1")),
			Wallet:    common.BytesToAddress([]byte("wallet1")),
			TokenAddr: common.BytesToAddress([]byte("token1")),
			Path:      []common.Address{},
			Factory:   [20]byte{},
			Price:     &big.Int{},
			Value:     &big.Int{},
		},
		{
			TxHash:    common.BytesToHash([]byte("tx2")),
			Wallet:    common.BytesToAddress([]byte("wallet2")),
			TokenAddr: common.BytesToAddress([]byte("token1")),
			Path:      []common.Address{},
			Factory:   [20]byte{},
			Price:     &big.Int{},
			Value:     &big.Int{},
		},
	}

	for i := range swaps {
		if err := db.PutSwap(tx, swaps[i]); err != nil {
			t.Fatal(err)
		}
	}

	for i := range tokens {
		if err := db.PutToken(tx, tokens[i]); err != nil {
			t.Fatal(err)
		}
	}

	for i := range accounts {
		if err := db.PutAccount(tx, accounts[i]); err != nil {
			t.Fatal(err)
		}
	}

	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	return d
}
