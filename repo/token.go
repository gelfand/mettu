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
	Decimals    int64
	Price       *big.Int
	TotalBought *big.Int
	TimesBought int
}

func (t Token) Denominator() *big.Int {
	return new(big.Int).Exp(big.NewInt(10), big.NewInt(t.Decimals), nil)
}

// _token is an internal type, we use it to store big.Int in the database.
type _token struct {
	Address common.Address
	Symbol  string
	// Amount is stored as bytes.
	Decimals    int64
	TimesBought int
	Price       []byte
	TotalBought []byte
}

// PutToken puts Token object into the storage.
func (db *DB) PutToken(tx kv.RwTx, t Token) error {
	tokenVal := _token{
		Address:     t.Address,
		Symbol:      t.Symbol,
		Decimals:    t.Decimals,
		TimesBought: t.TimesBought,
		Price:       big.NewInt(0).Bytes(),
		TotalBought: t.TotalBought.Bytes(),
	}
	if t.Price != nil {
		tokenVal.Price = t.Price.Bytes()
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

func (db *DB) HasToken(tx kv.Tx, addr common.Address) (bool, error) {
	return tx.Has(tokenStorage, addr.Bytes())
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
		Decimals:    tokenVal.Decimals,
		TotalBought: new(big.Int).SetBytes(tokenVal.TotalBought),
		TimesBought: tokenVal.TimesBought,
		Price:       new(big.Int).SetBytes(tokenVal.Price),
	}
	return t, nil
}

func (db *DB) AllTokens(tx kv.Tx) ([]Token, error) {
	var tokens []Token
	if tokensErr := tx.ForEach(tokenStorage, []byte{}, func(_, v []byte) error {
		var tokenVal _token
		if err := cbor.Unmarshal(bytes.NewReader(v), &tokenVal); err != nil {
			return fmt.Errorf("unable to decode token, err=%w", err)
		}

		t := Token{
			Address:     tokenVal.Address,
			Symbol:      tokenVal.Symbol,
			Decimals:    tokenVal.Decimals,
			Price:       new(big.Int).SetBytes(tokenVal.Price),
			TotalBought: new(big.Int).SetBytes(tokenVal.TotalBought),
			TimesBought: tokenVal.TimesBought,
		}
		tokens = append(tokens, t)

		return nil
	}); tokensErr != nil {
		return nil, fmt.Errorf("unable to retrieve all tokens: err=%w", tokensErr)
	}

	return tokens, nil
}

func (db *DB) AllTokensMap(tx kv.Tx) (map[common.Address]Token, error) {
	tokens := make(map[common.Address]Token)
	if tokensErr := tx.ForEach(tokenStorage, []byte{}, func(_, v []byte) error {
		var tokenVal _token
		if err := cbor.Unmarshal(bytes.NewReader(v), &tokenVal); err != nil {
			return fmt.Errorf("unable to decode token value, err=%w", err)
		}

		t := Token{
			Address:     tokenVal.Address,
			Symbol:      tokenVal.Symbol,
			TimesBought: tokenVal.TimesBought,
			Decimals:    tokenVal.Decimals,
			TotalBought: new(big.Int).SetBytes(tokenVal.TotalBought),
			Price:       new(big.Int).SetBytes(tokenVal.Price),
		}
		tokens[tokenVal.Address] = t

		return nil
	}); tokensErr != nil {
		return nil, fmt.Errorf("unable to retrieve all tokens: err=%w", tokensErr)
	}

	return tokens, nil
}
