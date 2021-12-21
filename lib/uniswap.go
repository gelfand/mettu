package lib

import "math/big"

var big10 = big.NewInt(10)

type Reserves struct {
	In, Out *big.Int
}

func CalculatePrice(amountOut *big.Int, reserves []Reserves) *big.Int {
	value := new(big.Int).Set(amountOut)

	for _, reserve := range reserves {
		value = new(big.Int).Mul(value, reserve.In)
		value = new(big.Int).Div(value, reserve.Out)
	}

	return value
}
