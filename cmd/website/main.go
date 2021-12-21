package main

import (
	"net/http"
	"os"
	"os/signal"

	"github.com/gelfand/mettu/repo"
	"golang.org/x/net/context"
)

var (
	db *repo.DB

	rpcAddr = "ws://127.0.0.1:8545"
)

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

	http.ListenAndServeTLS(":8080", "certs/localhost.pem", "certs/localhost-key.pem", nil)
}
