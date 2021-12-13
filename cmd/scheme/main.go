package main

import (
	"context"
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gelfand/mettu/repo"
)

type Token struct {
	Address  common.Address
	Symbol   string
	Decimals uint64
	Bought   uint64
}

func main() {
	ctx := context.Background()
	db, err := repo.NewDB()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(db.RwDB.AllBuckets())
	tx, err := db.RwDB.BeginRw(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer tx.Rollback()

	tx.Put("S")
}
