package main

import (
	"net/http"
	"sync/atomic"

	"github.com/gelfand/mettu/repo"
)

var swaps atomic.Value

type swap struct {
	Swap repo.Swap
}

func ListSwaps(w http.ResponseWriter, r *http.Request) {
	v, ok := swapsDat.Load().([]swap)
	if !ok {
		return
	}
	_ = v
}

func ListSwapsByWallet(w http.ResponseWriter, r *http.Request) {
	v, ok := swaps.Load().([]swap)
	if !ok {
		w.Write([]byte("ERROR: Something isn't okay with cached data!"))
		return
	}
	_ = v
}

func paginateSwaps(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// just a stub.. some ideas are to look at URL query params for something like
		// the page number, or the limit, and send a query cursor down the chain
		next.ServeHTTP(w, r)
	})
}
