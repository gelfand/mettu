package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"runtime/pprof"

	"github.com/gelfand/log"
	"github.com/gelfand/mettu/core"
	_ "github.com/gelfand/mettu/internal/abi"
	"github.com/gelfand/mettu/repo"
)

var homedir, _ = os.UserHomeDir()

var (
	doInit  = flag.Bool("init", false, "initialize new database")
	rpcAddr = flag.String("rpc.addr", "ws://127.0.0.1:8545", "Ethereum RPC address")
	datadir = flag.String("datadir", homedir+"/.mettu/", "path to the mettu database")

	cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")
	memprofile = flag.String("memprofile", "", "write memory profile to `file`")
)

func usage() {
	fmt.Fprintln(os.Stderr, "Usage of mettu...")
	flag.PrintDefaults()
}

func main() {
	flag.Usage = usage
	flag.Parse()

	os.UserHomeDir()

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Error("Could not create CPU profile", "err", err)
		}
		defer f.Close() // error handling omitted for example
		runtime.SetCPUProfileRate(1000)
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Error("Could not start CPU profile", "err", err)
		}
		defer pprof.StopCPUProfile()
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	dbPath, err := filepath.Abs(*datadir)
	if err != nil {
		log.Error("Invalid path", "err", err)
		return
	}

	if *doInit {
		if err := initDB(ctx, dbPath); err != nil {
			log.Error("Unable to initialize new database", "err", err)
			return
		}
	}

	log.Info("Successfully initialized new db")

	coordinator, err := core.NewCoordinator(ctx, dbPath, *rpcAddr)
	if err != nil {
		log.Error("Unable to create new Coordinator", "err", err)
		return
	}
	coordinator.Run(ctx)
	<-ctx.Done()

	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			log.Error("could not create memory profile", "err", err)
		}
		defer f.Close() // error handling omitted for example
		runtime.GC()    // get up-to-date statistics
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Error("could not write memory profile", "err", err)
		}
	}
}

func initDB(ctx context.Context, dbPath string) error {
	db, err := repo.NewDB(dbPath)
	if err != nil {
		return fmt.Errorf("unable to initialize new database, err=%w", err)
	}
	defer db.Close()

	f, err := os.Open("exchanges.json")
	if err != nil {
		return fmt.Errorf("unable to open `exchanges.json`, err=%w", err)
	}
	defer f.Close()

	var exchanges []repo.Exchange
	dec := json.NewDecoder(f)
	if err := dec.Decode(&exchanges); err != nil {
		return fmt.Errorf("unable to unmarshal data of `exchanges.json`, err=%w", err)
	}

	tx, err := db.BeginRw(ctx)
	if err != nil {
		return fmt.Errorf("unable to begin read-write transaction, err=%w", err)
	}
	defer tx.Rollback()

	for i := range exchanges {
		if err := db.PutExchange(tx, exchanges[i]); err != nil {
			return fmt.Errorf(fmt.Sprintf("unable to insert %d exchange: %v,", i, exchanges[i])+"err=%w", err)
		}
	}

	return tx.Commit()
}
