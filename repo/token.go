package repo

import (
	"bytes"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gelfand/mettu/internal/cbor"
	"github.com/ledgerwatch/erigon-lib/kv"
)

// Token is Ethereum `token`.
type Token struct {
	Address     common.Address
	Symbol      string
	TimesBought int
	Decimals    *big.Int
	TotalBought *big.Int
}

// _token is an internal type, we use it to store big.Int in the database.
type _token struct {
	Address common.Address
	Symbol  string
	// Amount is stored as bytes.
	TimesBought int
	Decimals    []byte
	TotalBought []byte
}

// PutToken puts Token object into the storage.
func (db *DB) PutToken(tx kv.RwTx, t Token) error {
	tokenVal := _token{
		Address:     t.Address,
		Symbol:      t.Symbol,
		TimesBought: t.TimesBought,
		Decimals:    t.Decimals.Bytes(),
		TotalBought: t.TotalBought.Bytes(),
	}

	var buf bytes.Buffer
	if err := cbor.Marshal(&buf, tokenVal); err != nil {
		return fmt.Errorf("unable encode token=%v, err=%w", t, err)
	}

	if err := tx.Put(tokenStorage, t.Address.Bytes(), buf.Bytes()); err != nil {
		return fmt.Errorf("unable to put token=%v, err=%w", t, err)
	}

	return nil
}

// PeekToken returns Token from the key value storage by it's Address.
func (db *DB) PeekToken(tx kv.Tx, addr common.Address) (Token, error) {
	val, err := tx.GetOne(tokenStorage, addr.Bytes())
	if err != nil {
		return Token{}, fmt.Errorf("unable to get token by address=%v, err=%w", addr, err)
	}

	var tokenVal _token
	if err := cbor.Unmarshal(bytes.NewReader(val), &tokenVal); err != nil {
		return Token{}, fmt.Errorf("unable to decode token, err=%w", err)
	}

	t := Token{
		Address:     tokenVal.Address,
		Symbol:      tokenVal.Symbol,
		TimesBought: tokenVal.TimesBought,
		Decimals:    new(big.Int).SetBytes(tokenVal.Decimals),
		TotalBought: new(big.Int).SetBytes(tokenVal.TotalBought),
	}
	return t, nil
}
