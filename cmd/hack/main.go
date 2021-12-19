package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gelfand/mettu/repo"
)

var path = flag.String("path", "./database", "database path")

var (
	homedir, _ = os.UserHomeDir()

	wethAddr   = common.HexToAddress("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2")
	usdcAddr   = common.HexToAddress("0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48")
	routerAddr = common.HexToAddress("0x7a250d5630B4cF539739dF2C5dAcb4c659F2488D")
	aaveAddr   = common.HexToAddress("0x7Fc66500c84A76Ad7e9c93437bFc5Ac33E2DDaE9")
)

func main() {
	flag.Parse()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	db, err := repo.NewDB(homedir + "/.mettu/")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	tx, err := db.BeginRw(ctx)
	defer tx.Rollback()
	if err != nil {
		log.Fatal(err)
	}

	if err := db.FlushBucket(tx, "SwapStorage"); err != nil {
		log.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		log.Fatal(err)
	}

	// ctx := context.Background()
	// defer db.Close()

	// tx, err := db.BeginRo(ctx)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// fmt.Println(db.AllExchanges(tx))
}
