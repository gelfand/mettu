package coordinator

import (
	"context"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gelfand/mettu/repo"
)

type Coordinator struct {
	Handler *http.Handler

	db *repo.DB
	wg sync.WaitGroup

	exchanges atomic.Value
	tokens    atomic.Value
	patterns  atomic.Value
	swaps     atomic.Value
	wallets   atomic.Value
}

func New() *Coordinator {
	return &Coordinator{}
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
