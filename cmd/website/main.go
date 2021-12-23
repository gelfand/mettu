package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/gelfand/mettu/repo"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"golang.org/x/net/context"
)

var (
	db *repo.DB

	rpcAddr = "ws://127.0.0.1:8545"
	creds   = map[string]string{
		"Eugene": "9b1e8a94fcdb88c8391ec1200718b3ddd73fb631b9c6b5d56619852a47833665",
		"Sats":   "75408122edc81988b92988054d2b4339f88e01d3efb7ec55cd275a558be71ac2",
		"user1":  "0a041b9462caa4a31bac3567e0b6e6fd9100787db2ab433d96f6d178cabfce90",
		"user2":  "6025d18fe48abd45168528f18a82e265dd98d421a7084aa09f61b341703901a3",
		"user3":  "5860faf02b6bc6222ba5aca523560f0e364ccd8b67bee486fe8bf7c01d492ccb",
		"user4":  "5269ef980de47819ba3d14340f4665262c41e933dc92c1a27dd5d01b047ac80e",
		"user5":  "5a39bead318f306939acb1d016647be2e38c6501c58367fdb3e9f52542aa2442",
	}
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	go func() {
		<-ctx.Done()
		os.Exit(0)
	}()
	defer cancel()

	homedir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	db, err = repo.NewDB(homedir + "/.mettu/")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	go runCaching(ctx, db)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.CleanPath)
	r.Use(middleware.NoCache)
	r.Use(middleware.BasicAuth("Alpha Leak", creds))

	r.Mount("/static/css/", http.StripPrefix("/static/css/", http.FileServer(http.Dir("static/css"))))
	r.Mount("/static/js/", http.StripPrefix("/static/js/", http.FileServer(http.Dir("static/js"))))

	r.Get("/", func(w http.ResponseWriter, _ *http.Request) {
		f, err := os.Open("./static/index.html")
		log.Fatalf("Unable to read index html template: %v", err)

		buf := make([]byte, 8192)
		n, _ := f.Read(buf)
		w.Write(buf[:n])
	})

	r.Get("/exchanges", exchangesHandler)
	r.Get("/swaps", swapsHandler)
	r.Get("/patterns", patternsHandler)
	r.Get("/tokens", tokensHandler)
	r.Get("/wallets", walletsHandler)

	server := &http.Server{
		Addr:    "0.0.0.0:433",
		Handler: r,
	}
	log.Fatal(server.ListenAndServeTLS(homedir+"/.x509/cert.pem", homedir+"/.x509/key.pem"))
}
