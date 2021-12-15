package ethclient

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gelfand/mettu/erc20"
	"github.com/gelfand/mettu/repo"
)

type Client struct {
	*ethclient.Client
}

func DialContext(ctx context.Context, rawurl string) (*Client, error) {
	ec, err := ethclient.DialContext(ctx, rawurl)
	if err != nil {
		return nil, err
	}
	return &Client{ec}, nil
}

func (c *Client) TokenAt(addr common.Address) (repo.Token, error) {
	t, err := erc20.NewErc20Caller(addr, c)
	if err != nil {
		return repo.Token{}, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	symbol, err := t.Symbol(&bind.CallOpts{Context: ctx})
	if err != nil {
		return repo.Token{}, fmt.Errorf("unable to retrieve token symbol: %w", err)
	}

	decimals, err := t.Decimals(&bind.CallOpts{Context: ctx})
	if err != nil {
		return repo.Token{}, fmt.Errorf("unable to retrieve token decimals: %w", err)
	}

	return repo.Token{
		Address:     addr,
		Symbol:      symbol,
		Decimals:    int64(decimals),
		TotalBought: big.NewInt(0),
		TimesBought: 0,
	}, nil
}
