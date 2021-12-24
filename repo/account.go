package repo

import (
	"bytes"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gelfand/mettu/internal/cbor"
	"github.com/ledgerwatch/erigon-lib/kv"
)

type Account struct {
	Address  common.Address
	Balance  *big.Int
	Received *big.Int
	Spent    *big.Int
	Exchange string
}

type _account struct {
	Balance  []byte
	Received []byte
	Spent    []byte
	Exchange string
}

func (db *DB) PutAccount(tx kv.RwTx, acc Account) error {
	a := _account{
		Balance:  acc.Balance.Bytes(),
		Received: acc.Received.Bytes(),
		Spent:    acc.Spent.Bytes(),
		Exchange: acc.Exchange,
	}

	var accBuf bytes.Buffer
	if err := cbor.Marshal(&accBuf, a); err != nil {
		return fmt.Errorf("unable to marshal account value: %w", err)
	}

	if err := tx.Put(accountStorage, acc.Address.Bytes(), accBuf.Bytes()); err != nil {
		return fmt.Errorf("unable to put new account entry: %w", err)
	}
	return nil
}

func (db *DB) HasAccount(tx kv.Tx, addr common.Address) (bool, error) {
	return tx.Has(accountStorage, addr.Bytes())
}

func (db *DB) PeekAccount(tx kv.Tx, address common.Address) (Account, error) {
	val, err := tx.GetOne(accountStorage, address.Bytes())
	if err != nil {
		return Account{}, fmt.Errorf("unable to tx.GetOne: %w", err)
	}
	var a _account
	if err := cbor.Unmarshal(bytes.NewReader(val), &a); err != nil {
		return Account{}, fmt.Errorf("unable to unmarshal account value: %w", err)
	}

	return Account{
		Address:  address,
		Balance:  new(big.Int).SetBytes(a.Balance),
		Received: new(big.Int).SetBytes(a.Received),
		Spent:    new(big.Int).SetBytes(a.Spent),
		Exchange: a.Exchange,
	}, nil
}

func (db *DB) AllAccounts(tx kv.Tx) ([]Account, error) {
	var accounts []Account
	if err := tx.ForEach(accountStorage, []byte{}, func(k, v []byte) error {
		addr := common.BytesToAddress(k)
		var a _account
		if err := cbor.Unmarshal(bytes.NewReader(v), &a); err != nil {
			return fmt.Errorf("unable to unmarshal account value, address: %v, err: %w", addr, err)
		}

		acc := Account{
			Address:  addr,
			Balance:  new(big.Int).SetBytes(a.Balance),
			Received: new(big.Int).SetBytes(a.Received),
			Spent:    new(big.Int).SetBytes(a.Spent),
			Exchange: a.Exchange,
		}
		accounts = append(accounts, acc)
		return nil
	}); err != nil {
		return nil, fmt.Errorf("tx.ForEach: %w", err)
	}

	return accounts, nil
}

func (db *DB) AllAccountsMap(tx kv.Tx) (map[common.Address]Account, error) {
	accounts := make(map[common.Address]Account)
	if err := tx.ForEach(accountStorage, []byte{}, func(k, v []byte) error {
		addr := common.BytesToAddress(k)
		var a _account
		if err := cbor.Unmarshal(bytes.NewReader(v), &a); err != nil {
			return fmt.Errorf("unable to unmarshal account value, address: %v, err: %w", addr, err)
		}

		acc := Account{
			Address:  addr,
			Balance:  new(big.Int).SetBytes(a.Balance),
			Received: new(big.Int).SetBytes(a.Received),
			Spent:    new(big.Int).SetBytes(a.Spent),
			Exchange: a.Exchange,
		}
		accounts[addr] = acc
		return nil
	}); err != nil {
		return nil, fmt.Errorf("tx.ForEach: %w", err)
	}

	return accounts, nil
}
