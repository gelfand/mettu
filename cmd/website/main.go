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
	go func() {
		<-ctx.Done()
		os.Exit(0)
	}()
	defer cancel()

	var err error
	db, err = repo.NewDB("/Users/eugene/.mettu/")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	go caching(ctx, db)

	mux := http.NewServeMux()
	mux.HandleFunc("/", mainHandler)
	mux.HandleFunc("/wallets", walletsHandler)
	mux.HandleFunc("/tokens", tokensHandler)
	mux.HandleFunc("/exchanges", exchangesHandler)
	mux.HandleFunc("/patterns", patternsHandler)
	mux.HandleFunc("/swaps", swapsHandler)
	mux.Handle("/static/css/", http.StripPrefix("/static/css/", http.FileServer(http.Dir("static/css"))))

	server := &http.Server{
		Addr:    "0.0.0.0:443",
		Handler: mux,
	}
	server.ListenAndServeTLS("certs/cdn.pem", "certs/cdn-key.pem")
}
