package abi

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/gelfand/mettu/uniswap/router"
)

var (
	swapExactETHForTokens *abi.Method
	swapETHForExactTokens *abi.Method
)

var (
	swapExactETHForTokensID = [4]byte{0x7f, 0xf3, 0x6a, 0xb5}
	swapETHForExactTokensID = [4]byte{0xfb, 0x3b, 0xdb, 0x41}
)

func init() {
	routerABI, _ := abi.JSON(strings.NewReader(router.RouterABI))
	swapExactETHForTokens, _ = routerABI.MethodById(swapExactETHForTokensID[:])
	swapETHForExactTokens, _ = routerABI.MethodById(swapETHForExactTokensID[:])
}

type TxDat struct {
	AmountIn  *big.Int
	AmountOut *big.Int
	TokenIn   common.Address
	TokenOut  common.Address
}

func Decode(tx *types.Transaction) (TxDat, error) {
	methodID := [4]byte{}
	copy(methodID[:], tx.Data()[:4])

	switch methodID {
	case swapExactETHForTokensID:
		inputData := map[string]interface{}{}
		if err := swapExactETHForTokens.Inputs.UnpackIntoMap(inputData, tx.Data()[4:]); err != nil {
			return TxDat{}, fmt.Errorf("unable to decode swapExactETHForTokens: %w", err)
		}

		amountIn := tx.Value()
		amountOut := inputData["amountOutMin"].(*big.Int)

		path := inputData["path"].([]common.Address)
		tokenIn := path[0]
		tokenOut := path[len(path)-1]

		return TxDat{
			AmountIn:  amountIn,
			AmountOut: amountOut,
			TokenIn:   tokenIn,
			TokenOut:  tokenOut,
		}, nil
	case swapETHForExactTokensID:
		inputData := map[string]interface{}{}
		if err := swapETHForExactTokens.Inputs.UnpackIntoMap(inputData, tx.Data()[4:]); err != nil {
			return TxDat{}, fmt.Errorf("unable to decode swapETHForExactTokens: %w", err)
		}

		amountIn := tx.Value()
		amountOut := inputData["amountOut"].(*big.Int)

		path := inputData["path"].([]common.Address)
		tokenIn := path[0]
		tokenOut := path[len(path)-1]

		return TxDat{
			AmountIn:  amountIn,
			AmountOut: amountOut,
			TokenIn:   tokenIn,
			TokenOut:  tokenOut,
		}, nil
	default:
		return TxDat{}, nil
	}
}
