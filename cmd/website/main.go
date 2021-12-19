package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gelfand/mettu/cmd/website/util"
	"github.com/gelfand/mettu/internal/ethclient"
	"github.com/gelfand/mettu/repo"
	"golang.org/x/net/context"
)

var db *repo.DB

// templates
var (
	tokensTmpl    = template.Must(template.New("tokens.tmpl.html").ParseFiles("./templates/tokens.tmpl.html"))
	patternsTmpl  = template.Must(template.New("patterns.tmpl.html").ParseFiles("./templates/patterns.tmpl.html"))
	walletsTmpl   = template.Must(template.New("wallets.tmpl.html").ParseFiles("./templates/wallets.tmpl.html"))
	exchangesTmpl = template.Must(template.New("exchanges.tmpl.html").ParseFiles("./templates/exchanges.tmpl.html"))
	swapsTmpl     = template.Must(template.New("swaps.tmpl.html").ParseFiles("./templates/swaps.tmpl.html"))

	walletsDat   atomic.Value
	tokensDat    atomic.Value
	patternsDat  atomic.Value
	swapsDat     atomic.Value
	exchangesDat atomic.Value

	rpcAddr = "ws://127.0.0.1:8545"
)

func walletsHandler(w http.ResponseWriter, r *http.Request) {
	v, ok := walletsDat.Load().([]byte)
	if !ok {
		return
	}
	w.Write(v)
}

func exchangesHandler(w http.ResponseWriter, r *http.Request) {
	v, ok := exchangesDat.Load().([]byte)
	if !ok {
		return
	}
	w.Write(v)
}

func patternsHandler(w http.ResponseWriter, r *http.Request) {
	v, ok := patternsDat.Load().([]byte)
	if !ok {
		return
	}
	w.Write(v)
}

func tokensHandler(w http.ResponseWriter, r *http.Request) {
	v, ok := tokensDat.Load().([]byte)
	if !ok {
		return
	}
	w.Write(v)
}

func swapsHandler(w http.ResponseWriter, r *http.Request) {
	v, ok := swapsDat.Load().([]byte)
	if !ok {
		return
	}
	w.Write(v)
}

func mainHandler(w http.ResponseWriter, r *http.Request) {
	f, _ := os.Open("./templates/main.html")
	buf := make([]byte, 8192)
	n, _ := f.Read(buf)
	w.Write(buf[:n])
}

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

func cacheTokens(ctx context.Context, db *repo.DB, wg *sync.WaitGroup) {
	defer wg.Done()
	tx, err := db.BeginRo(ctx)
	if err != nil {
		panic(err)
	}
	tokens, err := db.AllTokens(tx)
	if err != nil {
		log.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		log.Fatal(err)
	}

	sort.SliceStable(tokens, func(i, j int) bool {
		return tokens[i].TotalBought.Cmp(tokens[j].TotalBought) == 1
	})
	for i := range tokens {
		util.NormalizePrecision(tokens[i].TotalBought)
	}

	var buf bytes.Buffer
	if err := tokensTmpl.Execute(&buf, tokens); err != nil {
		log.Fatal(err)
	}
	tokensDat.Store(buf.Bytes())
}

func cacheExchanges(ctx context.Context, db *repo.DB, wg *sync.WaitGroup) {
	defer wg.Done()
	tx, err := db.BeginRo(ctx)
	if err != nil {
		panic(err)
	}
	exchanges, err := db.AllExchanges(tx)
	if err != nil {
		log.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		log.Fatal(err)
	}

	var buf bytes.Buffer
	if err := exchangesTmpl.Execute(&buf, exchanges); err != nil {
		log.Fatal(err)
	}

	exchangesDat.Store(buf.Bytes())
}

func cacheWallets(ctx context.Context, db *repo.DB, wg *sync.WaitGroup) {
	defer wg.Done()
	tx, err := db.BeginRo(ctx)
	if err != nil {
		panic(err)
	}
	wallets, err := db.AllAccounts(tx)
	if err != nil {
		log.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		log.Fatal(err)
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
		log.Fatal(err)
	}

	walletsDat.Store(buf.Bytes())
}

func cachePatterns(ctx context.Context, db *repo.DB, wg *sync.WaitGroup) {
	defer wg.Done()
	tx, err := db.BeginRo(ctx)
	if err != nil {
		panic(err)
	}
	defer tx.Rollback()
	patterns, err := db.AllPatternsData(tx)
	if err != nil {
		log.Fatal(err)
	}

	if err := tx.Commit(); err != nil {
		log.Fatal(err)
	}

	for i := range patterns {
		util.NormalizePrecision(patterns[i].Value)
	}

	sort.SliceStable(patterns, func(i, j int) bool {
		return patterns[i].Value.Cmp(patterns[j].Value) == 1
	})

	var buf bytes.Buffer
	if err := patternsTmpl.Execute(&buf, patterns); err != nil {
		log.Fatal(err)
	}
	patternsDat.Store(buf.Bytes())
}

func cacheSwaps(ctx context.Context, db *repo.DB, wg *sync.WaitGroup) {
	defer wg.Done()
	tx, err := db.BeginRo(ctx)
	if err != nil {
		panic(err)
	}
	client, err := ethclient.DialContext(ctx, rpcAddr)
	if err != nil {
		panic(err)
	}
	swaps, err := db.AllSwaps(tx)
	if err != nil {
		panic(err)
	}

	swaps, err = client.FetchSwapsData(swaps)
	if err != nil {
		panic(err)
	}

	for i := range swaps {
		util.NormalizePrecision(swaps[i].Profit)
		util.NormalizePrecision(swaps[i].Value)
	}

	sort.SliceStable(swaps, func(i, j int) bool {
		return swaps[i].Profitability > swaps[j].Profitability
	})

	var buf bytes.Buffer
	if err := swapsTmpl.Execute(&buf, swaps); err != nil {
		panic(err)
	}

	swapsDat.Store(buf.Bytes())
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	var err error
	db, err = repo.NewDB("/Users/eugene/.mettu/")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	go caching(ctx, db)

	http.HandleFunc("/", mainHandler)
	http.HandleFunc("/wallets", walletsHandler)
	http.HandleFunc("/tokens", tokensHandler)
	http.HandleFunc("/exchanges", exchangesHandler)
	http.HandleFunc("/patterns", patternsHandler)
	http.HandleFunc("/swaps", swapsHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.ListenAndServe(":8080", nil)
}

func handleErr(w io.Writer, err error) {
	fmt.Fprintf(w, "ERROR: %v\n", err)
}
