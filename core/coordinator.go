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
	ethclient "github.com/gelfand/mettu/internal/ethclient"
	"github.com/gelfand/mettu/repo"
)

type Account struct {
	Address common.Address `json:"Address"`
	Value   *big.Int       `json:"Value"`
	From    repo.Exchange  `json:"From"`
}

type Coordinator struct {
	// TODO: maybe make use of this lock.
	lock sync.Mutex

	db     *repo.DB
	client *ethclient.Client
	signer types.Signer

	seenAccounts map[common.Address]Account
	exchanges    map[common.Address]repo.Exchange
	headersCh    chan *types.Header
	txsChan      chan []*types.Transaction
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
		db:           db,
		client:       client,
		signer:       types.LatestSignerForChainID(chainID),
		seenAccounts: make(map[common.Address]Account),
		exchanges:    exchanges,
		headersCh:    make(chan *types.Header),
		txsChan:      make(chan []*types.Transaction),
	}
	return c, tx.Commit()
}

// TODO: Parallelize
func (c *Coordinator) processTransactions(ctx context.Context, txs []*types.Transaction) {
	c.lock.Lock()
	defer c.lock.Unlock()
	tx, err := c.db.BeginRw(ctx)
	if err != nil {
		log.Error("Unable to begin Database transaction", "err", err)
		return
	}
	defer tx.Rollback()

	for _, txn := range txs {
		if txn.To() == nil || txn.Value().Cmp(big.NewInt(1e18)) == -1 {
			continue
		}

		from, _ := types.Sender(c.signer, txn)
		cex, ok := c.exchanges[from]
		if ok {
			log.Info("Detected new CEX transfer", "To", *txn.To(), "Value", new(big.Int).Div(txn.Value(), big.NewInt(1e18)).String()+" ETH", "From", cex.Name)
			acc := Account{
				Address: *txn.To(),
				Value:   txn.Value(),
				From:    cex,
			}

			c.seenAccounts[*txn.To()] = acc

			continue
		}

		if _, ok := c.seenAccounts[from]; !ok || len(txn.Data()) < 4 {
			continue
		}

		methodID := [4]byte{}
		copy(methodID[:], txn.Data()[:4])
		fmt.Println(methodID)
		if methodID != abintr.SwapETHForExactTokensID && methodID != abintr.SwapExactETHForTokensID {
			continue
		}

		txData, err := abintr.Decode(txn)
		if err != nil {
			log.Debug("Unable to decode transaction", "err", err)
			continue
		}

		var tokens []repo.Token
		for _, tokenAddr := range txData.Path {
			has, err1 := c.db.HasToken(tx, tokenAddr)
			if err1 != nil {
				log.Error("Unexpected error, db.HasToken()", "err", err1)
				continue

			}
			if has {
				token, errPeek := c.db.PeekToken(tx, tokenAddr)
				if errPeek != nil {
					log.Error("Unable to db.PeekToken()", "err", errPeek)
					continue
				}
				tokens = append(tokens, token)
				continue
			}

			token, err1 := c.client.TokenAt(tokenAddr)
			if err1 != nil {
				log.Error("Unable to retrieve Token data", "addr", tokenAddr.String(), "err", err1)
				continue
			}
			tokens = append(tokens, token)
		}
		tokenOut := tokens[len(tokens)-1]
		tokenOut.TotalBought = new(big.Int).Add(tokenOut.TotalBought, txn.Value())
		tokenOut.TimesBought++
		tokens[len(tokens)-1] = tokenOut

		var pattern repo.Pattern
		hasPattern, err := c.db.HasPattern(tx, tokenOut.Address, cex.Name)
		if err != nil {
			log.Error("unable to check if pattern already persists in the database", "err", err)
			continue
		}
		if hasPattern {
			pattern, err = c.db.PeekPattern(tx, tokenOut.Address, cex.Name)
			if err != nil {
				log.Error("unable to peek pattern", "err", err)
				continue
			}
		} else {
			pattern = repo.Pattern{
				TokenAddr:    tokenOut.Address,
				ExchangeName: cex.Name,
				Value:        big.NewInt(0),
				TimesOccured: 0,
			}
		}
		pattern.Value = new(big.Int).Add(pattern.Value, txn.Value())
		pattern.TimesOccured++
		if err := c.db.PutPattern(tx, pattern); err != nil {
			log.Error("unable to put pattern", "err", err)
		}

		log.Info("Pattern", "token", pattern.TokenAddr, "Exchange", pattern.ExchangeName, "totalAmount", pattern.Value, "timesOccured", pattern.TimesOccured)

		for _, token := range tokens {
			if err = c.db.PutToken(tx, token); err != nil {
				log.Error("Unable to update Token information for", "addr", token.Address, "err", err)
				continue
			}
			log.Info("Successfully updated Token statistics for", "addr", token.Address, "symbol", token.Symbol, "totalBought", token.TimesBought, "timesBought", token.TimesBought)
		}

		if err := tx.Commit(); err != nil {
			log.Error("Unable to commit storage Transaction, rollbacking", "err", err)
		}

	}
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
			c.processTransactions(ctx, txs)
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
