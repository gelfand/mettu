package main

import (
	"fmt"
	"html/template"
	"log"
	"math/big"
	"net/http"
	"os"
	"sort"

	"github.com/gelfand/mettu/repo"
)

var (
	db *repo.DB

	decPrecision = big.NewInt(1e18)
)

// templates
var (
	tokensTmpl    = template.Must(template.New("tokens.tmpl.html").ParseFiles("./templates/tokens.tmpl.html"))
	patternsTmpl  = template.Must(template.New("patterns.tmpl.html").ParseFiles("./templates/patterns.tmpl.html"))
	walletsTmpl   = template.Must(template.New("wallets.tmpl.html").ParseFiles("./templates/wallets.tmpl.html"))
	exchangesTmpl = template.Must(template.New("exchanges.tmpl.html").ParseFiles("./templates/exchanges.tmpl.html"))
)

func walletsHandler(w http.ResponseWriter, r *http.Request) {
	tx, err := db.BeginRo(r.Context())
	if err != nil {
		log.Fatal(err)
	}
	defer tx.Rollback()

	wallets, err := db.AllAccounts(tx)
	if err != nil {
		log.Fatal(err)
	}

	sort.SliceStable(wallets, func(i, j int) bool {
		return wallets[i].TotalSpent.Cmp(wallets[j].TotalSpent) == 1
	})

	// normalize big Ints to be in human readable format
	for i := range wallets {
		normalizePrecision(wallets[i].TotalSpent)
		normalizePrecision(wallets[i].TotalReceived)
	}

	if err := walletsTmpl.Execute(w, wallets); err != nil {
		log.Fatal(err)
	}

	fmt.Println(r.Header)

	tx.Commit()
}

func exchangesHandler(w http.ResponseWriter, r *http.Request) {
	tx, err := db.BeginRo(r.Context())
	if err != nil {
		log.Fatal(err)
	}
	defer tx.Rollback()

	exchanges, err := db.AllExchanges(tx)
	if err != nil {
		log.Fatal(err)
	}

	if err := exchangesTmpl.Execute(w, exchanges); err != nil {
		log.Fatal(err)
	}

	tx.Commit()
}

func patternsHandler(w http.ResponseWriter, r *http.Request) {
	tx, err := db.BeginRo(r.Context())
	if err != nil {
		log.Fatal(err)
	}
	defer tx.Rollback()

	patterns, err := db.AllPatternsData(tx)
	if err != nil {
		log.Fatal(err)
	}

	for i := range patterns {
		normalizePrecision(patterns[i].Value)
	}

	sort.SliceStable(patterns, func(i, j int) bool {
		return patterns[i].Value.Cmp(patterns[j].Value) == 1
	})

	if err := patternsTmpl.Execute(w, patterns); err != nil {
		log.Fatal(err)
	}

	tx.Commit()
}

func tokensHandler(w http.ResponseWriter, r *http.Request) {
	tx, err := db.BeginRo(r.Context())
	if err != nil {
		log.Fatal(err)
	}
	defer tx.Rollback()

	tokens, err := db.AllTokens(tx)
	if err != nil {
		log.Fatal(err)
	}

	sort.SliceStable(tokens, func(i, j int) bool {
		return tokens[i].TotalBought.Cmp(tokens[j].TotalBought) == 1
	})
	for i := range tokens {
		normalizePrecision(tokens[i].TotalBought)
	}

	if err := tokensTmpl.Execute(w, tokens); err != nil {
		log.Fatal(err)
	}

	tx.Commit()
}

func mainHandler(w http.ResponseWriter, r *http.Request) {
	f, _ := os.Open("./templates/main.html")
	buf := make([]byte, 8192)
	n, _ := f.Read(buf)
	w.Write(buf[:n])
}

func main() {
	var err error
	db, err = repo.NewDB("/Users/eugene/.mettu/")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	http.HandleFunc("/", mainHandler)
	http.HandleFunc("/wallets", walletsHandler)
	http.HandleFunc("/tokens", tokensHandler)
	http.HandleFunc("/exchanges", exchangesHandler)
	http.HandleFunc("/patterns", patternsHandler)
	http.ListenAndServe(":8080", nil)
}

func normalizePrecision(v *big.Int) *big.Int {
	v = v.Div(v, decPrecision)
	return v
}
