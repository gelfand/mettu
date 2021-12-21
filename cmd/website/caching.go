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

var client, _ = ethclient.DialContext(context.Background(), rpcAddr) // Templates.
var (
	tokensTmpl    = template.Must(template.New("tokens.tmpl.html").ParseFiles("./templates/tokens.tmpl.html"))
	patternsTmpl  = template.Must(template.New("patterns.tmpl.html").ParseFiles("./templates/patterns.tmpl.html"))
	walletsTmpl   = template.Must(template.New("wallets.tmpl.html").ParseFiles("./templates/wallets.tmpl.html"))
	exchangesTmpl = template.Must(template.New("exchanges.tmpl.html").ParseFiles("./templates/exchanges.tmpl.html"))

	swapsTmpl = template.Must(template.New("swaps.tmpl.html").ParseFiles("./templates/swaps.tmpl.html"))
)

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

	tokens, err := db.AllTokens(tx)
	if err != nil {
		handleErr(&tokensDat, "could not retrieve all tokens data", err)
		return
	}

	if err := tx.Commit(); err != nil {
		handleErr(&tokensDat, "could not commit read-only transaction", err)
		return
	}

	sort.SliceStable(tokens, func(i, j int) bool {
		return tokens[i].TotalBought.Cmp(tokens[j].TotalBought) == 1
	})
	for i := range tokens {
		util.NormalizePrecision(tokens[i].TotalBought)
	}

	var buf bytes.Buffer
	if err := tokensTmpl.Execute(&buf, tokens); err != nil {
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
		Diff         int64
	}

	swapsData := make([]swapData, len(swaps))
	for i, swap := range swaps {
		token := tokens[swap.TokenAddr]

		v := swapData{
			TokenAddr:    util.AddressShort(swap.TokenAddr),
			Swap:         swap,
			Wallet:       util.AddressShort(swap.Wallet),
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
		denominator := new(big.Rat).SetInt(big.NewInt(1e18))
		priceRat := new(big.Rat).Quo(numerator, denominator)
		v.Price = priceRat.FloatString(6)

		price := lib.CalculatePrice(token.Denominator(), reserves)

		numerator = new(big.Rat).SetInt(price)
		priceRat = new(big.Rat).Quo(numerator, denominator)
		v.CurrentPrice = priceRat.FloatString(6)

		valRat := new(big.Rat).SetInt(swap.Value)
		valRat = valRat.Quo(valRat, denominator)
		v.Value = valRat.FloatString(3)

		var diff *big.Int

		if price.Cmp(swap.Price) > 0 && swap.Price.Cmp(common.Big0) != 0 {
			diff = new(big.Int).Div(price, swap.Price)
		} else if price.Cmp(common.Big0) != 0 {
			diff = new(big.Int).Div(swap.Price, price)
			diff = new(big.Int).Neg(diff)
		} else {
			diff = big.NewInt(0)
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
		v.Diff = diff.Int64()

		swapsData[i] = v
	}

	sort.SliceStable(swapsData, func(i, j int) bool {
		return swapsData[i].Diff > swapsData[j].Diff
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
