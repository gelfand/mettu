package main

import (
	"net/http"
	"os"
)

// walletsHandler atomicly loads cached response for exchanges and writes it.
func walletsHandler(w http.ResponseWriter, r *http.Request) {
	v, ok := walletsDat.Load().([]byte)
	if !ok {
		return
	}
	w.Write(v)
}

// exchangesHandler atomicly loads cached response for exchanges and writes it.
func exchangesHandler(w http.ResponseWriter, _ *http.Request) {
	v, ok := exchangesDat.Load().([]byte)
	if !ok {
		return
	}

	w.Write(v)
}

func patternsHandler(w http.ResponseWriter, _ *http.Request) {
	v, ok := patternsDat.Load().([]byte)
	if !ok {
		return
	}
	w.Write(v)
}

// tokensHandler atomicly loads cached response for tokens and writes it.
func tokensHandler(w http.ResponseWriter, _ *http.Request) {
	v, ok := tokensDat.Load().([]byte)
	if !ok {
		return
	}
	w.Write(v)
}

// swapsHandler atomicly loads cached response for swaps and writes it.
func swapsHandler(w http.ResponseWriter, _ *http.Request) {
	v, ok := swapsDat.Load().([]byte)
	if !ok {
		return
	}
	w.Write(v)
}

func authHandler(w http.ResponseWriter, r *http.Request) {
	f, _ := os.Open("./static/login.html")
	buf := make([]byte, 8192)
	n, _ := f.Read(buf)
	w.Write(buf[:n])
}

// mainHandler is router handler.
func mainHandler(w http.ResponseWriter, r *http.Request) {
	f, _ := os.Open("./static/index.html")
	buf := make([]byte, 8192)
	n, _ := f.Read(buf)
	w.Write(buf[:n])
}
