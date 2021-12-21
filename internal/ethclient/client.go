package ethclient

import (
	"context"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
)

var defaultTimeout = 10 * time.Second

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

// // PriceAt calculates price of end token by it's swap path.
// func (c *Client) PriceAt(factoryAddr common.Address, path []repo.Token) (*big.Int, error) {
// 	f, err := factory.NewFactoryCaller(factoryAddr, c)
// 	if err != nil {
// 		return nil, fmt.Errorf("unable to resolve factory related to the given router: %w", err)
// 	}

// 	val := new(big.Rat).SetFloat64(0.97)
// 	for i := 0; i < len(path)-1; i++ {
// 		token0, token1 := path[i], path[i+1]

// 		// flag represents token0 > token1
// 		r, err := c.reservesAt(f, token0, token1)
// 		if err != nil {
// 			return nil, fmt.Errorf("could not retrieve reserves: %w", err)
// 		}

// 		inRat, outRat := new(big.Rat).SetInt(r.in), new(big.Rat).SetInt(r.out)
// 		pricePerOne := new(big.Rat).Quo(inRat, outRat)
// 		val = val.Mul(val, pricePerOne)
// 	}

// 	val = val.Mul(val, big.NewRat(1e18, 1))
// 	v := new(big.Int).Div(val.Num(), val.Denom())
// 	return v, nil
// }

// func (c *Client) FetchSwapsData(swaps []repo.SwapWithToken) ([]repo.SwapWithToken, error) {
// 	for i := range swaps {
// 		var err error
// 		swaps[i].CurrPrice, err = c.PriceAt(swaps[i].Factory, swaps[i].Path)
// 		if err != nil {
// 			continue
// 		}

// 		if swaps[i].Price.Cmp(common.Big0) == 0 {
// 			continue
// 		}
// 		amount := new(big.Int).Div(swaps[i].Value, swaps[i].Price)
// 		swaps[i].Profit = new(big.Int).Sub(new(big.Int).Mul(amount, swaps[i].CurrPrice), swaps[i].Value)

// 		valRat := new(big.Rat).SetInt(swaps[i].Value)
// 		profitRat := new(big.Rat).SetInt(swaps[i].Profit)

// 		profitabilityRat := new(big.Rat).Quo(profitRat, valRat)
// 		profitabilityRat = profitabilityRat.Mul(profitabilityRat, new(big.Rat).SetFloat64(100.0))
// 		profitability, _ := profitabilityRat.Float64()
// 		swaps[i].Profitability = int(profitability)

// 		priceRat := new(big.Rat).SetInt(swaps[i].Price)
// 		priceRat = priceRat.Quo(priceRat, new(big.Rat).SetUint64(1e18))
// 		swaps[i].PriceFloat, _ = priceRat.Float64()

// 		currPriceRat := new(big.Rat).SetInt(swaps[i].CurrPrice)
// 		currPriceRat = currPriceRat.Quo(currPriceRat, new(big.Rat).SetUint64(1e18))
// 		swaps[i].CurrPriceFloat, _ = currPriceRat.Float64()

// 	}
// 	return swaps, nil
// }

// func (c *Client) FetchSwapsDataMap(swapsDat map[common.Address][]repo.SwapWithToken) (map[common.Address][]repo.SwapWithToken, error) {
// 	for k, swaps := range swapsDat {
// 		for i := 0; i < len(swaps); i++ {
// 			var err error
// 			swaps[i].CurrPrice, err = c.PriceAt(swaps[i].Factory, swaps[i].Path)
// 			if err != nil {
// 				return nil, err
// 			}

// 			amount := new(big.Int).Div(swaps[i].Value, swaps[i].Price)
// 			swaps[i].Profit = new(big.Int).Sub(new(big.Int).Mul(amount, swaps[i].CurrPrice), swaps[i].Value)

// 			valRat := new(big.Rat).SetInt(swaps[i].Value)
// 			profitRat := new(big.Rat).SetInt(swaps[i].Profit)

// 			profitabilityRat := new(big.Rat).Quo(profitRat, valRat)
// 			profitabilityRat = profitabilityRat.Mul(profitabilityRat, new(big.Rat).SetFloat64(100.0))
// 			profitability, _ := profitabilityRat.Float64()
// 			swaps[i].Profitability = int(profitability)

// 			priceRat := new(big.Rat).SetInt(swaps[i].Price)
// 			priceRat = priceRat.Quo(priceRat, new(big.Rat).SetUint64(1e18))
// 			swaps[i].PriceFloat, _ = priceRat.Float64()

// 			currPriceRat := new(big.Rat).SetInt(swaps[i].CurrPrice)
// 			currPriceRat = currPriceRat.Quo(currPriceRat, new(big.Rat).SetUint64(1e18))
// 			swaps[i].CurrPriceFloat, _ = currPriceRat.Float64()
// 		}
// 		swapsDat[k] = swaps
// 	}
// 	return swapsDat, nil
// }

// func (c *Client) reservesAt(f *factory.FactoryCaller, token0 repo.Token, token1 repo.Token) (reserves, error) {
// 	flag := cmpAddresses(token0.Address, token1.Address)
// 	if flag {
// 		token0, token1 = token1, token0
// 	}

// 	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
// 	defer cancel()
// 	pairAddr, err := f.GetPair(&bind.CallOpts{Context: ctx}, token0.Address, token1.Address)
// 	if err != nil {
// 		return reserves{}, err
// 	}

// 	p, err := pair.NewPairCaller(pairAddr, c)
// 	if err != nil {
// 		return reserves{}, fmt.Errorf("invalid pair address: %w", err)
// 	}

// 	ctx, cancel = context.WithTimeout(context.Background(), defaultTimeout)
// 	defer cancel()
// 	reservesV, err := p.GetReserves(&bind.CallOpts{Context: ctx})
// 	if err != nil {
// 		return reserves{}, err
// 	}
// 	in, out := reservesV.Reserve0, reservesV.Reserve1

// 	in = in.Div(in, token0.Denominator())
// 	out = out.Div(out, token1.Denominator())

// 	in = toEthPrecision(in)
// 	out = toEthPrecision(out)

// 	if flag {
// 		in, out = out, in
// 	}
// 	return reserves{
// 		in:  in,
// 		out: out,
// 	}, nil
// }

// func toEthPrecision(v *big.Int) *big.Int {
// 	v = v.Mul(v, oneEther)
// 	return v
// }
