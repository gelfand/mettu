package ethclient

import (
	"context"
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gelfand/mettu/erc20"
	"github.com/gelfand/mettu/lib"
	"github.com/gelfand/mettu/repo"
	"github.com/gelfand/mettu/uniswap/factory"
	"github.com/gelfand/mettu/uniswap/pair"
	"github.com/gelfand/mettu/uniswap/router"
)

// errors.
var (
	errIdenticalAddresses = errors.New("identical addresses")
	errZeroAddress        = errors.New("zero address")
)

var zeroAddress = big.NewInt(0)

// type pairCaller struct {
// 	caller *pair.PairCaller
// 	flag   bool
// }

// var (
// 	factoryCallers = sync.Map{}
// 	pairCallers    = sync.Map{}
// )

// func getReservesFast(p *pairCaller) (lib.Reserves, error) {
// 	fmt.Println(p)
// 	reserves, err := p.caller.GetReserves(&bind.CallOpts{})
// 	if err != nil {
// 		return lib.Reserves{}, nil
// 	}

// 	reserveA, reserveB := reserves.Reserve0, reserves.Reserve1
// 	if !p.flag {
// 		reserveA, reserveB = reserveB, reserveA
// 	}

// 	return lib.Reserves{
// 		In:  reserveA,
// 		Out: reserveB,
// 	}, nil
// }

func (c *Client) GetReservesPath(factoryAddr common.Address, path []common.Address) ([]lib.Reserves, error) {
	var r []lib.Reserves

	factoryCaller, err := factory.NewFactoryCaller(factoryAddr, c)
	if err != nil {
		return nil, err
	}

	for i := 1; i < len(path); i++ {
		tokenA, tokenB := path[i-1], path[i]

		flag, err := cmpAddresses(tokenA, tokenB)
		if err != nil {
			return nil, err
		}

		pairAddr, err := factoryCaller.GetPair(&bind.CallOpts{}, tokenA, tokenB)
		if err != nil {
			return nil, err
		}
		p, err := pair.NewPairCaller(pairAddr, c)
		if err != nil {
			return nil, err
		}

		reserves, err := p.GetReserves(&bind.CallOpts{})
		if err != nil {
			return nil, err
		}

		reserveA, reserveB := reserves.Reserve0, reserves.Reserve1
		if !flag {
			reserveA, reserveB = reserveB, reserveA
		}

		r = append(r, lib.Reserves{
			In:  reserveA,
			Out: reserveB,
		})
	}

	return r, nil
}

func (c *Client) GetReserves(factoryAddr, tokenA, tokenB common.Address) (lib.Reserves, error) {
	flag, err := cmpAddresses(tokenA, tokenB)
	if err != nil {
		return lib.Reserves{}, err
	}
	factoryCaller, err := factory.NewFactoryCaller(factoryAddr, c)
	if err != nil {
		return lib.Reserves{}, err
	}

	pairAddr, err := factoryCaller.GetPair(&bind.CallOpts{}, tokenA, tokenB)
	if err != nil {
		return lib.Reserves{}, err
	}
	p, err := pair.NewPairCaller(pairAddr, c)
	if err != nil {
		return lib.Reserves{}, err
	}

	reserves, err := p.GetReserves(&bind.CallOpts{})
	if err != nil {
		return lib.Reserves{}, err
	}

	reserveA, reserveB := reserves.Reserve0, reserves.Reserve1
	if !flag {
		reserveA, reserveB = reserveB, reserveA
	}

	return lib.Reserves{
		In:  reserveA,
		Out: reserveB,
	}, nil

	// factoryKey := [20]byte{}
	// copy(factoryKey[:], factoryAddr[:])

	// keyComparable := [60]byte{}
	// copy(keyComparable[:20], factoryAddr[:])
	// copy(keyComparable[20:40], tokenA[:])
	// copy(keyComparable[40:60], tokenB[:])

	// // 	NOTE: first try fast path, if pair caller is already cached.
	// if callerIface, ok := pairCallers.Load(keyComparable); ok {
	// 	return getReservesFast(callerIface.(*pairCaller))
	// }

	// flag, err := cmpAddresses(tokenA, tokenB)
	// if err != nil {
	// 	return lib.Reserves{}, err
	// }
	// // NOTE: a little bit slower path, if factory caller is already cached.
	// if factoryIface, ok := factoryCallers.Load(factoryKey); ok {
	// 	factoryCaller := factoryIface.(*factory.FactoryCaller)

	// 	pairAddr, err := factoryCaller.GetPair(&bind.CallOpts{}, tokenA, tokenB)
	// 	if err != nil {
	// 		return lib.Reserves{}, nil
	// 	}

	// 	p, err := pair.NewPairCaller(pairAddr, c)
	// 	if err != nil {
	// 		return lib.Reserves{}, nil
	// 	}

	// 	pCaller := &pairCaller{
	// 		caller: p,
	// 		flag:   flag,
	// 	}
	// 	pairCallers.Store(keyComparable, pCaller)

	// 	return getReservesFast(pCaller)
	// }

	// // slowest path if none of the above is possible
	// factoryCaller, err := factory.NewFactoryCaller(factoryAddr, c)
	// if err != nil {
	// 	return lib.Reserves{}, nil
	// }
	// // cache factory caller for future calls
	// factoryCallers.Store(factoryKey, factoryCaller)

	// pairAddr, err := factoryCaller.GetPair(&bind.CallOpts{}, tokenA, tokenB)
	// if err != nil {
	// 	return lib.Reserves{}, nil
	// }
	// p, err := pair.NewPairCaller(pairAddr, c)
	// if err != nil {
	// 	return lib.Reserves{}, nil
	// }

	// pCaller := &pairCaller{
	// 	caller: p,
	// 	flag:   flag,
	// }
	// pairCallers.Store(keyComparable, pCaller)

	// return getReservesFast(pCaller)
}

func (c *Client) FactoryAt(routerAddr common.Address) (common.Address, error) {
	r, err := router.NewRouterCaller(routerAddr, c)
	if err != nil {
		return common.Address{}, fmt.Errorf("could not resolve router at: %v, err: %w", routerAddr, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	factoryAddr, err := r.Factory(&bind.CallOpts{Context: ctx})
	if err != nil {
		return common.Address{}, fmt.Errorf("invalid router address: %w", err)
	}

	return factoryAddr, nil
}

func (c *Client) TokenAt(addr common.Address) (repo.Token, error) {
	t, err := erc20.NewErc20Caller(addr, c)
	if err != nil {
		return repo.Token{}, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
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

func cmpAddresses(x, y common.Address) (bool, error) {
	xBig, yBig := new(big.Int).SetBytes(x[:]), new(big.Int).SetBytes(y[:])
	if xBig.Cmp(zeroAddress) == 0 || yBig.Cmp(zeroAddress) == 0 {
		return false, errZeroAddress
	}
	if xBig.Cmp(yBig) == 0 {
		return true, errIdenticalAddresses
	}
	return xBig.Cmp(yBig) == -1, nil
}
