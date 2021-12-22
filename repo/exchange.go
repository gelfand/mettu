package repo

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ledgerwatch/erigon-lib/kv"
)

// Exchange is CEX.
type Exchange struct {
	// Name is value corresponding to the Address.
	Name string `json:"Name"`
	// We use Address as key in our storage layout.
	Address common.Address `json:"Address"`
}

// PutExchange inserts Exchange into the storage.
func (db *DB) PutExchange(tx kv.RwTx, e Exchange) error {
	if err := tx.Put(exchangeStorage, e.Address.Bytes(), []byte(e.Name)); err != nil {
		return fmt.Errorf("unable to put exchange=%v, err=%w", e, err)
	}

	return nil
}

// PeekExchange retrieves Exchange from the storage by give address.
func (db *DB) PeekExchange(tx kv.Tx, addr common.Address) (Exchange, error) {
	val, err := tx.GetOne(exchangeStorage, addr.Bytes())
	if err != nil {
		return Exchange{}, fmt.Errorf("unable to get exchange by address=%v, err=%w", addr, err)
	}

	return Exchange{
		Name:    string(val),
		Address: addr,
	}, nil
}

// AllExchanges returns all exchanges stored in the exchangeStorage.
func (db *DB) AllExchanges(tx kv.Tx) ([]Exchange, error) {
	var exchanges []Exchange
	if err := tx.ForEach(exchangeStorage, []byte{}, func(k, v []byte) error {
		exchanges = append(exchanges, Exchange{
			Name:    string(v),
			Address: common.BytesToAddress(k),
		})

		return nil
	}); err != nil {
		return nil, err
	}

	return exchanges, nil
}

// AllExchangesMap returns all exchanges in map being mapped to their addresses.
func (db *DB) AllExchangesMap(tx kv.Tx) (map[common.Address]Exchange, error) {
	exchanges := make(map[common.Address]Exchange)
	if err := tx.ForEach(exchangeStorage, []byte{}, func(k, v []byte) error {
		addr := common.BytesToAddress(k)

		exchanges[addr] = Exchange{
			Name:    string(v),
			Address: addr,
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return exchanges, nil
}
