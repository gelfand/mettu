package main

import (
	"context"
	"flag"
	"log"
	"math/big"
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
	uniAddr    = common.HexToAddress("0x1f9840a85d5af5bf1d1762f925bdaddc4201f984")
	wbtcAddr   = common.HexToAddress("0x2260fac5e5542a773aa44fbcfedf7c193bc2c599")
)

func main() {
	// flag.Parse()
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	// client, err := ethclient.DialContext(ctx, "ws://127.0.0.1:8545")
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// factoryAddr, err := client.FactoryAt(routerAddr)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Println(uniAddr)

	// path := []common.Address{wethAddr, usdcAddr, wbtcAddr, wethAddr, uniAddr}

	// // reservesA, err := client.GetReserves(factoryAddr, wethAddr, usdcAddr)
	// // if err != nil {
	// // 	log.Fatal(err)
	// // }
	// // reservesB, err := client.GetReserves(factoryAddr, usdcAddr, wbtcAddr)
	// // reservesC, err := client.GetReserves(factoryAddr, wbtcAddr, wethAddr)
	// // reservesD, err := client.GetReserves(factoryAddr, wethAddr, uniAddr)

	// // var reserves []lib.Reserves
	// // reserves = append(reserves, reservesA)
	// // reserves = append(reserves, reservesB)
	// // reserves = append(reserves, reservesC)
	// // reserves = append(reserves, reservesD)

	// reserves, err := client.GetReservesPath(factoryAddr, path)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// amountOut := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
	// price := lib.CalculatePrice(amountOut, reserves)
	// priceRat := new(big.Rat).SetInt(price)
	// priceRat = priceRat.Quo(priceRat, new(big.Rat).SetInt64(1e18))
	// fmt.Println(priceRat.Float64())

	// // fmt.Println(reserveA, reserveB, in)

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

	if err := tx.Commit(); err != nil {
		log.Fatal()
	}

	// ctx := context.Background()
	// defer db.Close()

	// tx, err := db.BeginRo(ctx)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// fmt.Println(db.AllExchanges(tx))
}

func quote(amountA, reserveA, reserveB *big.Int) *big.Int {
	amountA = amountA.Mul(amountA, reserveB)
	amountB := new(big.Int).Div(amountA, reserveA)
	return amountB
}

func getAmountIn(amountOut, reserveIn, reserveOut *big.Int) *big.Int {
	if amountOut.Cmp(common.Big0) == 0 {
		return nil
	}

	numerator := new(big.Int).Mul(reserveIn, big.NewInt(1000))
	denominator := new(big.Int).Sub(reserveOut, new(big.Int).Mul(amountOut, big.NewInt(997)))
	amountIn := new(big.Int).Div(numerator, denominator)
	amountIn = amountIn.Add(amountIn, common.Big1)
	return amountIn
}
