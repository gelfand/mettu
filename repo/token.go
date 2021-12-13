package repo

import (
	"bytes"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gelfand/mettu/internal/cbor"
	"github.com/ledgerwatch/erigon-lib/kv"
)

// Token is Ethereum `token`.
type Token struct {
	Symbol    string
	Decimals  int64
	Purchases int
	Address   common.Address
}

// PutToken puts Token object into the storage.
func (db *DB) PutToken(tx kv.RwTx, t Token) error {
	var buf bytes.Buffer
	if err := cbor.Marshal(&buf, t); err != nil {
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

	var t Token
	if err := cbor.Unmarshal(bytes.NewReader(val), &t); err != nil {
		return Token{}, fmt.Errorf("unable to decode token, err=%w", err)
	}

	return t, nil
}
