package util

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

var decPrecision = big.NewInt(1e18)

func NormalizePrecision(v *big.Int) *big.Int {
	if v == nil {
		v = new(big.Int).SetInt64(0)
	}
	v = v.Div(v, decPrecision)
	return v
}

func AddressShort(addr common.Address) string {
	return addr.String()[:8] + "..." + addr.String()[36:]
}

func HashShort(hash common.Hash) string {
	return hash.String()[:8] + "..." + hash.String()[60:]
}
