package abi

import (
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/gelfand/mettu/uniswap/router"
)

var ErrUnknownMethod = errors.New("unknown transaction method")

var (
	swapExactETHForTokens *abi.Method
	swapETHForExactTokens *abi.Method
)

var (
	SwapExactETHForTokensID = [4]byte{0x7f, 0xf3, 0x6a, 0xb5}
	SwapETHForExactTokensID = [4]byte{0xfb, 0x3b, 0xdb, 0x41}
)

func init() {
	routerABI, _ := abi.JSON(strings.NewReader(router.RouterABI))
	swapExactETHForTokens, _ = routerABI.MethodById(SwapExactETHForTokensID[:])
	swapETHForExactTokens, _ = routerABI.MethodById(SwapETHForExactTokensID[:])
}

type TxDat struct {
	AmountIn  *big.Int
	AmountOut *big.Int
	TokenIn   common.Address
	TokenOut  common.Address
	Path      []common.Address
}

func Decode(tx *types.Transaction) (TxDat, error) {
	methodID := [4]byte{}
	copy(methodID[:], tx.Data()[:4])

	switch methodID {
	case SwapExactETHForTokensID:
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
			Path:      path,
		}, nil
	case SwapETHForExactTokensID:
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
			Path:      path,
		}, nil
	default:
		return TxDat{}, ErrUnknownMethod
	}
}
