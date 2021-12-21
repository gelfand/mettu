package repo

import (
	"bytes"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gelfand/mettu/internal/cbor"
	"github.com/ledgerwatch/erigon-lib/kv"
)

type SwapWithToken struct {
	TxHash  common.Hash
	Wallet  Account
	Token   Token
	Path    []Token
	Factory common.Address
	Value   *big.Int

	Price      *big.Int
	PriceFloat float64

	CurrPrice      *big.Int
	CurrPriceFloat float64
}

type Swap struct {
	TxHash    common.Hash
	Wallet    common.Address
	TokenAddr common.Address
	Path      []common.Address
	Factory   common.Address
	Price     *big.Int
	Value     *big.Int
}

type _swap struct {
	Wallet    common.Address
	TokenAddr common.Address
	Path      []common.Address
	Factory   common.Address
	Price     []byte
	Value     []byte
}

func (db *DB) PutSwap(tx kv.RwTx, s Swap) error {
	swapVal := _swap{
		Wallet:    s.Wallet,
		TokenAddr: s.TokenAddr,
		Path:      s.Path,
		Factory:   s.Factory,
		Price:     s.Price.Bytes(),
		Value:     s.Value.Bytes(),
	}

	var buf bytes.Buffer
	if err := cbor.Marshal(&buf, swapVal); err != nil {
		return fmt.Errorf("unable to encode swap record: %w", err)
	}

	if err := tx.Put(swapStorage, s.TxHash.Bytes(), buf.Bytes()); err != nil {
		return fmt.Errorf("unable to put swap record: %w", err)
	}

	return nil
}

func (db *DB) PeekSwap(tx kv.Tx, txHash common.Hash) (Swap, error) {
	val, err := tx.GetOne(swapStorage, txHash.Bytes())
	if err != nil {
		return Swap{}, fmt.Errorf("could not peek swap record: %w", err)
	}

	var swapVal _swap
	if err := cbor.Unmarshal(bytes.NewReader(val), &swapVal); err != nil {
		return Swap{}, fmt.Errorf("could not unmarshal swap value: %w", err)
	}

	s := Swap{
		TxHash:    txHash,
		Wallet:    swapVal.Wallet,
		TokenAddr: swapVal.TokenAddr,
		Path:      swapVal.Path,
		Factory:   swapVal.Factory,
		Price:     new(big.Int).SetBytes(swapVal.Price),
		Value:     new(big.Int).SetBytes(swapVal.Value),
	}

	return s, nil
}

// DeleteSwap deletes swap record from swapStorage.
func (db *DB) DeleteSwap(tx kv.RwTx, txHash common.Hash) error {
	v, err := tx.GetOne(swapStorage, txHash.Bytes())
	if err != nil {
		return err
	}

	if err := tx.Delete(swapStorage, txHash.Bytes(), v); err != nil {
		return err
	}
	return nil
}

func (db *DB) AllSwaps(tx kv.Tx) ([]Swap, error) {
	// 	wallets, err := db.AllAccountsMap(tx)
	// 	if err != nil {
	// 		return nil, fmt.Errorf("unable to retrieve all accounts: %w", err)
	// 	}

	// 	tokens, err := db.AllTokensMap(tx)
	// 	if err != nil {
	// 		return nil, fmt.Errorf("unable to retrieve all tokens")
	// 	}

	var swaps []Swap
	if err := tx.ForEach(swapStorage, []byte{}, func(k, v []byte) error {
		txHash := common.BytesToHash(k)

		var swapVal _swap
		if err := cbor.Unmarshal(bytes.NewReader(v), &swapVal); err != nil {
			return fmt.Errorf("unable to decode swap record: %w", err)
		}
		s := Swap{
			TxHash:    txHash,
			Wallet:    swapVal.Wallet,
			TokenAddr: swapVal.TokenAddr,
			Path:      swapVal.Path,
			Factory:   swapVal.Factory,
			Price:     new(big.Int).SetBytes(swapVal.Price),
			Value:     new(big.Int).SetBytes(swapVal.Value),
		}

		swaps = append(swaps, s)
		return nil
	}); err != nil {
		return nil, fmt.Errorf("could not retrieve all swap records: %w", err)
	}

	return swaps, nil
}

// func (db *DB) AllSwapsMap(tx kv.Tx) (map[common.Address][]SwapWithToken, error) {
// 	wallets, err := db.AllAccountsMap(tx)
// 	if err != nil {
// 		return nil, fmt.Errorf("unable to retrieve all accounts: %w", err)
// 	}

// 	tokens, err := db.AllTokensMap(tx)
// 	if err != nil {
// 		return nil, fmt.Errorf("unable to retrieve all tokens")
// 	}

// 	swaps := make(map[common.Address][]SwapWithToken)
// 	if err := tx.ForEach(swapStorage, []byte{}, func(k, v []byte) error {
// 		txHash := common.BytesToHash(k)

// 		var swapVal _swap
// 		if err := cbor.Unmarshal(bytes.NewReader(v), &swapVal); err != nil {
// 			return fmt.Errorf("unable to decode swap record: %w", err)
// 		}

// 		var path []Token
// 		for _, tokenAddr := range swapVal.Path {
// 			path = append(path, tokens[tokenAddr])
// 		}

// 		s := SwapWithToken{
// 			TxHash:  txHash,
// 			Wallet:  wallets[swapVal.Wallet],
// 			Token:   tokens[swapVal.TokenAddr],
// 			Path:    path,
// 			Factory: swapVal.Factory,
// 			Price:   new(big.Int).SetBytes(swapVal.Price),
// 			Value:   new(big.Int).SetBytes(swapVal.Value),
// 		}
// 		if _, ok := swaps[swapVal.Wallet]; !ok {
// 			swaps[swapVal.Wallet] = []SwapWithToken{}
// 		}
// 		swaps[swapVal.Wallet] = append(swaps[swapVal.Wallet], s)

// 		return nil
// 	}); err != nil {
// 		return nil, fmt.Errorf("could not retrieve all swap records: %w", err)
// 	}

// 	return swaps, nil
// }
