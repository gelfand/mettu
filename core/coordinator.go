package core

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/gelfand/log"
	abintr "github.com/gelfand/mettu/internal/abi"
	"github.com/gelfand/mettu/lib"

	ethclient "github.com/gelfand/mettu/internal/ethclient"
	"github.com/gelfand/mettu/repo"
)

type Coordinator struct {
	// TODO: maybe make use of this lock.
	lock sync.Mutex

	db     *repo.DB
	client *ethclient.Client
	signer types.Signer

	exchanges map[common.Address]repo.Exchange
	headersCh chan *types.Header
	txsChan   chan []*types.Transaction

	exitCh chan struct{}
}

// NewCoordinator creates new Coordinator.
func NewCoordinator(ctx context.Context, dbPath string, rpcAddr string) (*Coordinator, error) {
	db, err := repo.NewDB(dbPath)
	if err != nil {
		return nil, fmt.Errorf("unable to open database: %w", err)
	}

	client, err := ethclient.DialContext(ctx, rpcAddr)
	if err != nil {
		return nil, fmt.Errorf("unable to establlish connection with Ethereum RPC: %w", err)
	}

	chainID, err := client.ChainID(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve chainID: %w", err)
	}

	tx, err := db.BeginRo(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to begin database transaction: %w", err)
	}
	defer tx.Rollback()

	exchanges, err := db.AllExchangesMap(tx)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve all exchanges from the database: %w", err)
	}

	c := &Coordinator{
		db:        db,
		client:    client,
		signer:    types.LatestSignerForChainID(chainID),
		exchanges: exchanges,
		headersCh: make(chan *types.Header),
		txsChan:   make(chan []*types.Transaction),
		exitCh:    make(chan struct{}, 1),
	}
	fmt.Println(len(c.exchanges))
	return c, tx.Commit()
}

// TODO: Parallelize.
func (c *Coordinator) processTransactions(ctx context.Context, txs []*types.Transaction) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	tx, err := c.db.BeginRw(ctx)
	if err != nil {
		return fmt.Errorf("unexpected error: could not begin database transaction: %w", err)
	}
	defer tx.Rollback()

	for _, txn := range txs {
		if txn.To() == nil || txn.Value().Cmp(big.NewInt(1e18)) == -1 {
			continue
		}
		if _, ok := c.exchanges[*txn.To()]; ok {
			continue
		}

		from, _ := types.Sender(c.signer, txn)
		cex, ok := c.exchanges[from]
		if ok {
			log.Info("Detected new CEX transfer", "To", *txn.To(), "Value", new(big.Int).Div(txn.Value(), big.NewInt(1e18)).String()+" ETH", "From", cex.Name)

			ok, err = c.db.HasAccount(tx, *txn.To())
			if err != nil {
				return fmt.Errorf("could not check if key exists in the db: %w", err)
			}
			var acc repo.Account
			if ok {
				acc, err = c.db.PeekAccount(tx, *txn.To())
				if err != nil {
					return fmt.Errorf("could not peek account: %w", err)
				}
				acc.TotalReceived = new(big.Int).Add(acc.TotalReceived, txn.Value())
			} else {
				acc = repo.Account{
					Address:       *txn.To(),
					TotalReceived: txn.Value(),
					TotalSpent:    big.NewInt(0),
					Exchange:      cex.Name,
				}
			}

			if err := c.db.PutAccount(tx, acc); err != nil {
				return fmt.Errorf("could not put account into key value storage: %w", err)
			}
			continue
		}

		ok, err := c.db.HasAccount(tx, from)
		if err != nil {
			return fmt.Errorf("could not check if account exists in the db: %w", err)
		}
		if !ok || len(txn.Data()) < 4 {
			continue
		}

		acc, err := c.db.PeekAccount(tx, from)
		if err != nil {
			return fmt.Errorf("could not peek account: %w", err)
		}
		methodID := [4]byte{}
		copy(methodID[:], txn.Data()[:4])
		if methodID != abintr.SwapETHForExactTokensID && methodID != abintr.SwapExactETHForTokensID {
			continue
		}

		txData, err := abintr.Decode(txn)
		if err != nil {
			log.Debug("Unable to decode transaction", "err", err)
			continue
		}
		factoryAddr, err := c.client.FactoryAt(*txn.To())
		if err != nil {
			log.Debug(err.Error())
			continue
		}

		var tokens []repo.Token
		for _, tokenAddr := range txData.Path {
			log.Info("Token", "address", tokenAddr)
			ok, err = c.db.HasToken(tx, tokenAddr)
			if err != nil {
				return fmt.Errorf("could not check if token exists in the db: %w", err)
			}
			var token repo.Token
			if ok {
				token, err = c.db.PeekToken(tx, tokenAddr)
				if err != nil {
					return fmt.Errorf("could not peek token: %w", err)
				}
				tokens = append(tokens, token)
				continue
			}

			token, err = c.client.TokenAt(tokenAddr)
			if err != nil {
				continue
			}
			tokens = append(tokens, token)
		}

		tokenOut := tokens[len(tokens)-1]
		reserves, err := c.client.GetReservesPath(factoryAddr, txData.Path)
		if err != nil {
			log.Debug(err.Error())
			continue
		}
		price := lib.CalculatePrice(tokenOut.Denominator(), reserves)

		ok, err = c.db.HasToken(tx, tokenOut.Address)
		if err != nil {
			return err
		}
		if !ok {
			tokenOut.Price = price
		} else {
			var tOut repo.Token
			tOut, err = c.db.PeekToken(tx, tokenOut.Address)
			if err != nil {
				return err
			}
			if tOut.Price.Cmp(common.Big0) == 0 {
				tokenOut.Price = price
			}
		}

		tokenOut.TotalBought = new(big.Int).Add(tokenOut.TotalBought, txn.Value())
		tokenOut.TimesBought++
		tokens[len(tokens)-1] = tokenOut

		ok, err = c.db.HasPattern(tx, tokenOut.Address, acc.Exchange)
		if err != nil {
			return fmt.Errorf("unable to check if pattern exists in the storage: %w", err)
		}
		var pattern repo.Pattern
		if ok {
			pattern, err = c.db.PeekPattern(tx, tokenOut.Address, acc.Exchange)
			if err != nil {
				return fmt.Errorf("unable to peek pattern: %w", err)
			}
		} else {
			pattern = repo.Pattern{
				TokenAddr:    tokenOut.Address,
				ExchangeName: acc.Exchange,
				Value:        big.NewInt(0),
				TimesOccured: 0,
			}
		}

		pattern.TimesOccured++
		pattern.Value = new(big.Int).Add(pattern.Value, txn.Value())

		acc.TotalSpent = new(big.Int).Add(acc.TotalSpent, txn.Value())

		if err = c.db.PutAccount(tx, acc); err != nil {
			return fmt.Errorf("unable to put updated account data: %w", err)
		}

		if err = c.db.PutPattern(tx, pattern); err != nil {
			return fmt.Errorf("unable to put updated pattern data: %w", err)
		}

		s := repo.Swap{
			TxHash:    txn.Hash(),
			Wallet:    from,
			TokenAddr: tokenOut.Address,
			Price:     price,
			Path:      txData.Path,
			Factory:   factoryAddr,
			Value:     txn.Value(),
		}

		if err = c.db.PutSwap(tx, s); err != nil {
			return fmt.Errorf("unable to put swap record: %w", err)
		}

		for _, token := range tokens {
			if err = c.db.PutToken(tx, token); err != nil {
				return fmt.Errorf("unable to put updated token data: %w", err)
			}
			log.Info("Successfully updated Token statistics for", "addr", token.Address, "symbol", token.Symbol, "totalBought", token.TotalBought.String()+" ETH", "timesBought", token.TimesBought)
		}
	}
	return tx.Commit()
}

func (c *Coordinator) proccessorLifecycle(ctx context.Context) {
	log.Info("Successfully started Processor lifecycle")
	cycleCounter := 0
	for {
		select {
		case <-ctx.Done():
			return
		case txs := <-c.txsChan:
			log.Info("Running lifecycle", "cycle", cycleCounter)
			if err := c.processTransactions(ctx, txs); err != nil {
				panic(err)
			}
			cycleCounter++
		}
	}
}

func (c *Coordinator) Run(ctx context.Context) {
	defer c.db.Close()
	go c.proccessorLifecycle(ctx)

	sub, err := c.client.SubscribeNewHead(ctx, c.headersCh)
	if err != nil {
		log.Error("unable to subscribe new headers, exiting", "err", err)
		os.Exit(1)
	}

	for {
		select {
		case <-ctx.Done():
			return
		case err := <-sub.Err():
			log.Error("unable to handle header subscription, exiting", "err", err)
			os.Exit(1)
		case header := <-c.headersCh:
			ctxWithTimeout, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			block, err := c.client.BlockByHash(ctxWithTimeout, header.Hash())
			if err != nil {
				log.Debug("Unable to retrieve block by hash", "err", err)
				cancel()
				continue
			}
			cancel()

			c.txsChan <- block.Transactions()
		}
	}
}

// func (c *Coordinator) dumpCache() {
// 	f, err := os.Create("seen_accounts.json")
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer f.Close()
// 	enc := json.NewEncoder(f)
// 	c.lock.Lock()

// 	if err := enc.Encode(c.seenAccounts); err != nil {
// 		panic(err)
// 	}
// 	c.lock.Unlock()
// }
