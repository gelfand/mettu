package util

import "math/big"

var decPrecision = big.NewInt(1e18)

func NormalizePrecision(v *big.Int) *big.Int {
	if v == nil {
		v = new(big.Int).SetInt64(0)
	}
	v = v.Div(v, decPrecision)
	return v
}
