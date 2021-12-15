package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/gelfand/mettu/core"
	_ "github.com/gelfand/mettu/internal/abi"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	coordinator, err := core.NewCoordinator(ctx, "./database", "ws://127.0.0.1:8545")
	if err != nil {
		log.Fatal(err)
	}
	go coordinator.Run(ctx)
	<-ctx.Done()
}
