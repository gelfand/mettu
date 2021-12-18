package repo

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ledgerwatch/erigon-lib/kv"
)

type Swap struct {
	TxHash    common.Hash
	Wallet    common.Address
	TokenAddr common.Address
	Price     *big.Int
	Value     *big.Int
}

type _swap struct {
	Wallet    common.Address
	TokenAddr common.Address
	Price     []byte
	Value     []byte
}

func (db *DB) PutSwap(tx kv.RwTx, record Swap) error {
	return nil
}

func (db *DB) PeekSwap(tx kv.Tx, txHash common.Hash) (Swap, error) {
	return Swap{}, nil
}

func (db *DB) AllSwaps(tx kv.Tx)
