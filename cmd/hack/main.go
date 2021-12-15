package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/gelfand/mettu/repo"
)

var path = flag.String("path", "./database", "database path")

func main() {
	flag.Parse()

	db, _ := repo.NewDB("./database")
	defer db.Close()
	f, err := os.Open("./repo/testdata/exchanges.json")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	tx, _ := db.BeginRo(context.Background())
	fmt.Println(db.AllExchangesMap(tx))
}
