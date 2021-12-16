package repo

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type Account struct {
	Address       common.Address
	TotalReceived *big.Int
	TotalSpent    *big.Int
}
