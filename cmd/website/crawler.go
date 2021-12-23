package main

import (
	"sync/atomic"

	"github.com/gelfand/mettu/internal/ethclient"
	"github.com/gelfand/mettu/repo"
)

type crawler struct {
	db     *repo.DB
	client *ethclient.Client

	swaps atomic.Value
}
