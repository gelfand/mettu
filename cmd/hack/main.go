package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/gelfand/mettu/repo"
)

var path = flag.String("path", "./database", "database path")

func main() {
	flag.Parse()

	ctx := context.Background()
	db, _ := repo.NewDB("./mettu_db")
	defer db.Close()

	tx, err := db.BeginRo(ctx)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(db.AllExchanges(tx))
}
