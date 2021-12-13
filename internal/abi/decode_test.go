package abi

import (
	"math/big"
	"reflect"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

func TestDecode(t *testing.T) {
	t.Parallel()

	type args struct {
		tx *types.Transaction
	}
	tests := []struct {
		name    string
		args    args
		want    TxDat
		wantErr bool
	}{
		{
			name: "swapExactETHForTokens0",
			args: args{
				tx: types.NewTx(&types.DynamicFeeTx{
					Data:  common.FromHex("0x7ff36ab50000000000000000000000000000000000000000000000003d91ae3365ec0cef000000000000000000000000000000000000000000000000000000000000008000000000000000000000000018b5b77fe9660b79f0283b1fc98097d97f9cb4a70000000000000000000000000000000000000000000000000000000061b5fb0d0000000000000000000000000000000000000000000000000000000000000002000000000000000000000000c02aaa39b223fe8d0a0e5c4f27ead9083c756cc2000000000000000000000000da5cae7bf4815e6ce3b2488ee102e67403245679"),
					Value: big.NewInt(9e17),
				}),
			},
			want: TxDat{
				AmountIn:  big.NewInt(9e17),
				AmountOut: big.NewInt(4436518643713182959),
				TokenIn:   common.HexToAddress("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2"),
				TokenOut:  common.HexToAddress("0xdA5caE7Bf4815e6cE3B2488Ee102E67403245679"),
			},
			wantErr: false,
		},
		{
			name: "swapETHForExactTokens",
			args: args{
				tx: types.NewTx(&types.DynamicFeeTx{
					Data:  common.FromHex("0xfb3bdb4100000000000000000000000000000000000000000000001043561a882930000000000000000000000000000000000000000000000000000000000000000000800000000000000000000000002ac3b47e7bc9d42822c1db3e6948c1a47051e8050000000000000000000000000000000000000000000000000000000061b5fb870000000000000000000000000000000000000000000000000000000000000003000000000000000000000000c02aaa39b223fe8d0a0e5c4f27ead9083c756cc2000000000000000000000000a0b86991c6218b36c1d19d4a2e9eb0ce3606eb480000000000000000000000002653891204f463fb2a2f4f412564b19e955166ae"),
					Value: big.NewInt(38e17),
				}),
			},
			want: TxDat{
				AmountIn:  big.NewInt(38e17),
				AmountOut: new(big.Int).SetBytes(common.FromHex("1043561a8829300000")),
				TokenIn:   common.HexToAddress("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2"),
				TokenOut:  common.HexToAddress("0x2653891204F463fb2a2F4f412564b19e955166aE"),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := Decode(tt.args.tx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Decode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Decode() = %v, want %v", got, tt.want)
			}
		})
	}
}
