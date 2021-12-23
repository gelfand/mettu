package crawler

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gelfand/mettu/repo"
	"github.com/go-chi/chi/v5"
)

type Coordinator struct {
	r chi.Router

	db *repo.DB
	wg sync.WaitGroup

	exchanges atomic.Value
	tokens    atomic.Value
	patterns  atomic.Value
	swaps     atomic.Value
	wallets   atomic.Value
}

func (c *Coordinator) ServeHTTP() {
}

func New(db *repo.DB) *Coordinator {
	c := &Coordinator{
		r:         chi.NewRouter(),
		db:        db,
		wg:        sync.WaitGroup{},
		exchanges: atomic.Value{},
		tokens:    atomic.Value{},
		patterns:  atomic.Value{},
		swaps:     atomic.Value{},
		wallets:   atomic.Value{},
	}
	return c
}

func (c *Coordinator) Run(ctx context.Context) error {
	cacheCycle := time.NewTicker(1 * time.Minute)

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-cacheCycle.C:
			c.wg.Add(1)
			c.doCacheCycle(ctx)
			c.wg.Wait()

		}
	}
}

func (c *Coordinator) doCacheCycle(ctx context.Context) {
	defer c.wg.Done()
	c.wg.Add(5)
}
