package repo

import (
	"bytes"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gelfand/mettu/internal/cbor"
	"github.com/ledgerwatch/erigon-lib/kv"
)

type FullPattern struct {
	Token    Token
	Exchange string
	Value    *big.Int
	Counter  int
}

type Pattern struct {
	TokenAddr    common.Address
	ExchangeName string
	Value        *big.Int
	TimesOccured int
}

type _patternKey struct {
	TokenAddr    common.Address
	ExchangeName string
}

type _patternValue struct {
	Value        []byte
	TimesOccured int
}

func (db *DB) PutPattern(tx kv.RwTx, p Pattern) error {
	key := _patternKey{
		TokenAddr:    p.TokenAddr,
		ExchangeName: p.ExchangeName,
	}
	value := _patternValue{
		Value:        p.Value.Bytes(),
		TimesOccured: p.TimesOccured,
	}

	var keyBuf bytes.Buffer
	var valueBuf bytes.Buffer

	if err := cbor.Marshal(&keyBuf, key); err != nil {
		return fmt.Errorf("unable to marshal pattern key value: %w", err)
	}
	if err := cbor.Marshal(&valueBuf, value); err != nil {
		return fmt.Errorf("unable to marshal pattern value: %w", err)
	}

	if err := tx.Put(patternStorage, keyBuf.Bytes(), valueBuf.Bytes()); err != nil {
		return fmt.Errorf("unable to put key value pattern: %w", err)
	}
	return nil
}

func (db *DB) HasPattern(tx kv.Tx, token common.Address, exchangeName string) (bool, error) {
	key := _patternKey{
		TokenAddr:    token,
		ExchangeName: exchangeName,
	}
	var keyBuf bytes.Buffer
	if err := cbor.Marshal(&keyBuf, key); err != nil {
		return false, fmt.Errorf("unable to marshal key value: %w", err)
	}

	return tx.Has(patternStorage, keyBuf.Bytes())
}

func (db *DB) PeekPattern(tx kv.Tx, token common.Address, exchangeName string) (Pattern, error) {
	key := _patternKey{
		TokenAddr:    token,
		ExchangeName: exchangeName,
	}
	var keyBuf bytes.Buffer
	if err := cbor.Marshal(&keyBuf, key); err != nil {
		return Pattern{}, fmt.Errorf("unable to marshal key value: %w", err)
	}

	val, err := tx.GetOne(patternStorage, keyBuf.Bytes())
	if err != nil {
		return Pattern{}, fmt.Errorf("unable to tx.GetOne in PeekPattern: %w", err)
	}

	var value _patternValue
	if err := cbor.Unmarshal(bytes.NewReader(val), &value); err != nil {
		return Pattern{}, fmt.Errorf("unable to unmarshal pattern value: %w", err)
	}

	return Pattern{
		TokenAddr:    token,
		ExchangeName: exchangeName,
		Value:        new(big.Int).SetBytes(value.Value),
		TimesOccured: value.TimesOccured,
	}, nil
}

func (db *DB) AllPatterns(tx kv.Tx) ([]Pattern, error) {
	var patterns []Pattern
	if err := tx.ForEach(patternStorage, []byte{}, func(k, v []byte) error {
		var (
			value _patternValue
			key   _patternKey
		)

		if err := cbor.Unmarshal(bytes.NewReader(k), &key); err != nil {
			return fmt.Errorf("could not unmarshal pattern key: %w", err)
		}
		if err := cbor.Unmarshal(bytes.NewReader(v), &value); err != nil {
			return fmt.Errorf("could not unmarshal pattern value: %w", err)
		}

		pattern := Pattern{
			TokenAddr:    key.TokenAddr,
			ExchangeName: key.ExchangeName,
			Value:        new(big.Int).SetBytes(value.Value),
			TimesOccured: value.TimesOccured,
		}
		patterns = append(patterns, pattern)
		return nil
	}); err != nil {
		return nil, fmt.Errorf("unable to iterate through all patterns: %w", err)
	}

	return patterns, nil
}

func (db *DB) AllPatternsData(tx kv.Tx) ([]FullPattern, error) {
	tokensMap, err := db.AllTokensMap(tx)
	if err != nil {
		return nil, err
	}

	var patterns []FullPattern
	if err := tx.ForEach(patternStorage, []byte{}, func(k, v []byte) error {
		var (
			value _patternValue
			key   _patternKey
		)

		if err := cbor.Unmarshal(bytes.NewReader(k), &key); err != nil {
			return fmt.Errorf("could not unmarshal pattern key: %w", err)
		}
		if err := cbor.Unmarshal(bytes.NewReader(v), &value); err != nil {
			return fmt.Errorf("could not unmarshal pattern value: %w", err)
		}

		pattern := FullPattern{
			Token:    tokensMap[key.TokenAddr],
			Exchange: key.ExchangeName,
			Value:    new(big.Int).SetBytes(value.Value),
			Counter:  value.TimesOccured,
		}
		patterns = append(patterns, pattern)
		return nil
	}); err != nil {
		return nil, fmt.Errorf("unable to iterate through all patterns: %w", err)
	}

	return patterns, nil
}
