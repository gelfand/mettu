package core

import (
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gelfand/mettu/repo"
)

type Coordinator struct {
	db     *repo.DB
	client *ethclient.Client

	headersCh chan *types.Header
	txChan    chan *types.Transaction
}

func (c *Coordinator) Run() {
}

func subscribeNewHeaders()
