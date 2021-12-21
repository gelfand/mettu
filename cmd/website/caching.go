package main

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"math/big"
	"os"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gelfand/mettu/cmd/website/util"
	"github.com/gelfand/mettu/internal/ethclient"
	"github.com/gelfand/mettu/lib"
	"github.com/gelfand/mettu/repo"
)

var client, _ = ethclient.DialContext(context.Background(), rpcAddr)

// Templates.
var (
	tokensTmpl    = template.Must(template.New("tokens.tmpl.html").ParseFiles("./static/templates/tokens.tmpl.html"))
	patternsTmpl  = template.Must(template.New("patterns.tmpl.html").ParseFiles("./static/templates/patterns.tmpl.html"))
	walletsTmpl   = template.Must(template.New("wallets.tmpl.html").ParseFiles("./static/templates/wallets.tmpl.html"))
	exchangesTmpl = template.Must(template.New("exchanges.tmpl.html").ParseFiles("./static/templates/exchanges.tmpl.html"))
	swapsTmpl     = template.Must(template.New("swaps.tmpl.html").ParseFiles("./static/templates/swaps.tmpl.html"))

	tokenDenominator = big.NewInt(1e18)
)

var rat0 = new(big.Rat).SetInt64(0)

// Atomic caching.
var (
	walletsDat   atomic.Value
	tokensDat    atomic.Value
	patternsDat  atomic.Value
	swapsDat     atomic.Value
	exchangesDat atomic.Value
)

func caching(ctx context.Context, db *repo.DB) {
	ticker := time.NewTicker(1 * time.Minute)

	wg := sync.WaitGroup{}

	cache := func(ctx context.Context, db *repo.DB, wg *sync.WaitGroup) error {
		wg.Add(5)

		go cacheSwaps(ctx, db, wg)
		go cacheExchanges(ctx, db, wg)
		go cachePatterns(ctx, db, wg)
		go cacheTokens(ctx, db, wg)
		go cacheWallets(ctx, db, wg)

		wg.Wait()
		return nil
	}

	cache(ctx, db, &wg)
	for {
		select {
		case <-ctx.Done():
			os.Exit(0)
		case <-ticker.C:
			cache(ctx, db, &wg)
		}
	}
}

// cacheTokens retrieves all tokens data and atomicly stores it in tokensDat.
func cacheTokens(ctx context.Context, db *repo.DB, wg *sync.WaitGroup) {
	defer wg.Done()

	tx, err := db.BeginRo(ctx)
	if err != nil {
		handleErr(&tokensDat, "could not begin read-only transaction", err)
		return
	}
	defer tx.Rollback()

	swaps, err := db.AllSwaps(tx)
	if err != nil {
		handleErr(&tokensDat, "could not retrieve all swaps data", err)
		return
	}

	tokens := make(map[common.Address]repo.Swap)
	for i := range swaps {
		if _, ok := tokens[swaps[i].TokenAddr]; ok {
			continue
		}
		tokens[swaps[i].TokenAddr] = swaps[i]
	}

	type tokenData struct {
		Token       repo.Token
		TotalBought *big.Int
		Price       string
		CurrPrice   string
		Diff        string
		DiffRat     *big.Rat
	}

	i := 0
	tt := make([]tokenData, len(tokens))
	for tokenAddr, v := range tokens {
		t, err := db.PeekToken(tx, tokenAddr)
		if err != nil {
			panic(err)
		}

		priceRat := new(big.Rat).SetInt(t.Price)
		priceRat = new(big.Rat).Quo(priceRat, new(big.Rat).SetInt64(1e18))

		r, err := client.GetReservesPath(v.Factory, v.Path)
		if err != nil {
			panic(err)
		}

		currPrice := lib.CalculatePrice(t.Denominator(), r)
		currPriceRat := new(big.Rat).SetInt(currPrice)
		currPriceRat = new(big.Rat).Quo(currPriceRat, new(big.Rat).SetInt64(1e18))

		diff := new(big.Rat).Set(rat0)
		if currPriceRat.Cmp(priceRat) > 0 && priceRat.Cmp(rat0) != 0 {
			diff = new(big.Rat).Quo(currPriceRat, priceRat)
		}

		tDat := tokenData{
			Token:       t,
			TotalBought: util.NormalizePrecision(t.TotalBought),
			Price:       priceRat.FloatString(6),
			CurrPrice:   currPriceRat.FloatString(6),
			Diff:        diff.FloatString(2),
			DiffRat:     diff,
		}
		tt[i] = tDat
		i++
	}
	if err := tx.Commit(); err != nil {
		handleErr(&tokensDat, "could not commit read-only transaction", err)
		return
	}

	// var tokensData []tokenData

	// for _, token := range t {
	//        r, err := client.GetReserves()

	// 	tData := tokenData{
	// 		Token:       token,
	// 		Price:       priceRat.FloatString(6),
	// 		TotalBought: util.NormalizePrecision(token.TotalBought),
	// 		CurrPrice:     ,
	// 		Change:      "",
	// 	}
	// }

	defer tx.Rollback()
	sort.SliceStable(tt, func(i, j int) bool {
		diffRat0, diffRat1 := tt[i].DiffRat, tt[j].DiffRat
		if diffRat0 == nil || diffRat1 == nil {
			fmt.Println(tt[i])
			return true
		}

		return diffRat0.Cmp(diffRat1) > 0
	})

	var buf bytes.Buffer
	if err := tokensTmpl.Execute(&buf, tt); err != nil {
		handleErr(&tokensDat, "could not execute token template", err)
		return
	}

	tokensDat.Store(buf.Bytes())
}

// cacheExchanges retrieves all exchanges data and atomicly stores it in exchangesDat.
func cacheExchanges(ctx context.Context, db *repo.DB, wg *sync.WaitGroup) {
	defer wg.Done()

	tx, err := db.BeginRo(ctx)
	if err != nil {
		handleErr(&exchangesDat, "could not begin read-only transaction", err)
		return
	}
	defer tx.Rollback()

	exchanges, err := db.AllExchanges(tx)
	if err != nil {
		handleErr(&exchangesDat, "could not retrieve all exchanges from the db", err)
		return
	}
	if err := tx.Commit(); err != nil {
		handleErr(&exchangesDat, "could not commit read-only transaction", err)
	}

	var buf bytes.Buffer
	if err := exchangesTmpl.Execute(&buf, exchanges); err != nil {
		handleErr(&exchangesDat, "could not execute exchanges template", err)
		return
	}

	exchangesDat.Store(buf.Bytes())
}

func cacheWallets(ctx context.Context, db *repo.DB, wg *sync.WaitGroup) {
	defer wg.Done()

	tx, err := db.BeginRo(ctx)
	if err != nil {
		handleErr(&walletsDat, "could not begin read-only transaction", err)
		return
	}
	defer tx.Rollback()

	wallets, err := db.AllAccounts(tx)
	if err != nil {
		handleErr(&walletsDat, "could not retrieve all accounts from the db", err)
		return
	}

	if err := tx.Commit(); err != nil {
		handleErr(&walletsDat, "could not commit read-only transaction", err)
		return
	}

	sort.SliceStable(wallets, func(i, j int) bool {
		return wallets[i].TotalSpent.Cmp(wallets[j].TotalSpent) == 1
	})

	// normalize big Ints to be in human readable format
	for i := range wallets {
		util.NormalizePrecision(wallets[i].TotalSpent)
		util.NormalizePrecision(wallets[i].TotalReceived)
	}

	var buf bytes.Buffer
	if err := walletsTmpl.Execute(&buf, wallets); err != nil {
		handleErr(&walletsDat, "could not execute wallets template", err)
		return
	}

	walletsDat.Store(buf.Bytes())
}

func cachePatterns(ctx context.Context, db *repo.DB, wg *sync.WaitGroup) {
	defer wg.Done()

	tx, err := db.BeginRo(ctx)
	if err != nil {
		handleErr(&patternsDat, "could not begin read-only transaction", err)
		return
	}
	defer tx.Rollback()

	patterns, err := db.AllPatternsData(tx)
	if err != nil {
		handleErr(&patternsDat, "could not retieve all patterns data", err)
		return
	}

	if err := tx.Commit(); err != nil {
		handleErr(&patternsDat, "could not commit read-only transaction", err)
		return
	}

	for i := range patterns {
		util.NormalizePrecision(patterns[i].Value)
	}

	sort.SliceStable(patterns, func(i, j int) bool {
		return patterns[i].Value.Cmp(patterns[j].Value) == 1
	})

	var buf bytes.Buffer
	if err := patternsTmpl.Execute(&buf, patterns); err != nil {
		handleErr(&patternsDat, "could not execute patterns template", err)
	}

	patternsDat.Store(buf.Bytes())
}

func cacheSwaps(ctx context.Context, db *repo.DB, wg *sync.WaitGroup) {
	defer wg.Done()

	tx, err := db.BeginRo(ctx)
	if err != nil {
		handleErr(&swapsDat, "could not begin read-only transaction", err)
		return
	}
	defer tx.Rollback()

	tokens, err := db.AllTokensMap(tx)
	if err != nil {
		panic(err)
	}
	swaps, err := db.AllSwaps(tx)
	if err != nil {
		handleErr(&swapsDat, "could not retrieve swaps data from the kv storage", err)
		return
	}

	type swapData struct {
		TokenAddr    string
		Swap         repo.Swap
		Wallet       string
		Symbol       string
		Value        string
		Price        string
		CurrentPrice string
		Diff         string
		DiffRat      *big.Rat
	}

	swapsData := make([]swapData, len(swaps))
	for i, swap := range swaps {
		token := tokens[swap.TokenAddr]

		v := swapData{
			TokenAddr:    swap.TokenAddr.String(),
			Swap:         swap,
			Wallet:       swap.Wallet.String(),
			Symbol:       token.Symbol,
			Value:        "",
			Price:        "",
			CurrentPrice: "",
		}

		reserves, err := client.GetReservesPath(swap.Factory, swap.Path)
		if err != nil {
			handleErr(&swapsDat, "could not get reserves", err)
		}

		numerator := new(big.Rat).SetInt(swap.Price)
		denominator := new(big.Rat).SetInt64(1e18)
		priceRat := new(big.Rat).Quo(numerator, denominator)
		v.Price = priceRat.FloatString(6)

		price := lib.CalculatePrice(token.Denominator(), reserves)

		numerator = new(big.Rat).SetInt(price)
		priceRat = new(big.Rat).Quo(numerator, denominator)
		v.CurrentPrice = priceRat.FloatString(6)

		v.Value = new(big.Int).Div(swap.Value, tokenDenominator).String()

		var diff *big.Rat
		if price.Cmp(swap.Price) > 0 && swap.Price.Cmp(common.Big0) != 0 {
			diff = new(big.Rat).Quo(new(big.Rat).SetInt(price), new(big.Rat).SetInt(swap.Price))
		} else {
			diff = new(big.Rat).Set(rat0)
		}
		//
		// diff := new(big.Int).Mul(Div(sum, sumDiv)
		// diff = diff.Mul(diff, )
		// if price.Cmp(swap.Price) > 0 {
		// 	increase := new(big.Int).Sub(price, swap.Price)
		// 	den := new(big.Int).Mul(swap.Price, big.NewInt(100))
		// 	den = den.Mul(den, token.Denominator())
		// 	diff = new(big.Int).Div(increase, den)
		// } else {
		// 	decrease := new(big.Int).Sub(swap.Price, price)
		// 	den := new(big.Int).Mul(swap.Price, big.NewInt(100))
		// 	den = den.Mul(den, token.Denominator())
		// 	diff = new(big.Int).Div(decrease, den)
		// }
		fmt.Println(diff)
		v.Diff = diff.FloatString(2)
		v.DiffRat = diff

		swapsData[i] = v
	}

	sort.SliceStable(swapsData, func(i, j int) bool {
		return swapsData[i].DiffRat.Cmp(swapsData[j].DiffRat) > 0
	})

	// swaps, err = client.FetchSwapsData(swaps)
	// if err != nil {
	// 	handleErr(&swapsDat, "could not fetch swaps data", err)
	// 	return
	// }

	// for i := range swaps {
	// 	util.NormalizePrecision(swaps[i].Profit)
	// 	util.NormalizePrecision(swaps[i].Value)
	// }

	// sort.SliceStable(swaps, func(i, j int) bool {
	// 	return swaps[i].Profitability > swaps[j].Profitability
	// })

	var buf bytes.Buffer
	if err := swapsTmpl.Execute(&buf, swapsData); err != nil {
		handleErr(&swapsDat, "could not execute swaps template", err)
		return
	}

	swapsDat.Store(buf.Bytes())
}
