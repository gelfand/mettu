package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/gelfand/log"
	"github.com/gelfand/mettu/cmd/website/internal/config"
	"github.com/gelfand/mettu/cmd/website/internal/server"
)

const (
	serverAddr = "0.0.0.0:443"
	rpcAddr    = "ws://127.0.0.1:8545"
)

var creds = map[string]string{
	"test":   "test",
	"Eugene": "9b1e8a94fcdb88c8391ec1200718b3ddd73fb631b9c6b5d56619852a47833665",
	"Sats":   "75408122edc81988b92988054d2b4339f88e01d3efb7ec55cd275a558be71ac2",
	"user1":  "0a041b9462caa4a31bac3567e0b6e6fd9100787db2ab433d96f6d178cabfce90",
	"user2":  "6025d18fe48abd45168528f18a82e265dd98d421a7084aa09f61b341703901a3",
	"user3":  "5860faf02b6bc6222ba5aca523560f0e364ccd8b67bee486fe8bf7c01d492ccb",
	"user4":  "5269ef980de47819ba3d14340f4665262c41e933dc92c1a27dd5d01b047ac80e",
	"user5":  "5a39bead318f306939acb1d016647be2e38c6501c58367fdb3e9f52542aa2442",
}

var (
	homedir, _ = os.UserHomeDir()

	datadirFlag = flag.String("datadir", homedir+"/.mettu", "path to the database.")
	certFlag    = flag.String("cert", homedir+"/.x509/cert.pem", "path to the certificate.")
	certkeyFlag = flag.String("key", homedir+"/.x509/key.pem", "path to the certificate key.")
	addrFlag    = flag.String("addr", serverAddr, "server address.")
)

func usage() {
	fmt.Fprintln(os.Stderr, "Usage of website...")
	flag.PrintDefaults()
}

func main() {
	flag.Usage = usage
	flag.Parse()

	cfg := &config.Config{
		Addr:    "127.0.0.1:8080",
		RPCAddr: rpcAddr,
		DBPath:  *datadirFlag,
	}
	server, err := server.NewServer(cfg)
	if err != nil {
		log.Fatal(err)
	}
	server.Install()
	server.ListenAndServeTLS("127.0.0.1:4444", *certFlag, *certkeyFlag)

	// ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	// go func() {
	// 	<-ctx.Done()
	// 	os.Exit(0)
	// }()
	// defer cancel()

	// c, err := newCrawler(context.Background(), nil)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// go c.run(ctx)
	// mux := chi.NewMux()
	// mux.Route("/patterns", func(r chi.Router) {
	// })

	// listAccounts := func(w http.ResponseWriter, r *http.Request) {
	// 	accs, ok := c.Accounts()
	// 	if !ok {
	// 		w.Write([]byte("Not ready yet..."))
	// 	}

	// 	if err := accountsTemplate.Execute(w, accs); err != nil {
	// 		panic(err)
	// 	}
	// }

	// mux.Route("/accounts", func(r chi.Router) {
	// 	r.Get("/", listAccounts)
	// })

	// // r.Get("/swaps", swapsHandler)
	// // r.Get("/patterns", patternsHandler)
	// // r.Get("/tokens", tokensHandler)
	// // r.Get("/wallets", walletsHandler)

	// server := &http.Server{
	// 	Addr:    *addr,
	// 	Handler: mux,
	// }
	// log.Fatal(server.ListenAndServeTLS(*cert, *certkey))
}

func paginate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// just a stub.. some ideas are to look at URL query params for something like
		// the page number, or the limit, and send a query cursor down the chain
		next.ServeHTTP(w, r)
	})
}
