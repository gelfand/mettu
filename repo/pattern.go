package repo

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ledgerwatch/erigon-lib/kv"
)

type Pattern struct {
	TokenAddress common.Address
	FromAddress  common.Address
	TotalAmount  *big.Int
	TimesOccured int
}

type _pattern struct {
	TokenAddress common.Address
	FromAddress  common.Address
	TimesOccured int
	TotalAmount  []byte
}

func (db *DB) PutPattern(tx kv.RwTx) (Pattern, error) {
	return Pattern{}, nil
}
