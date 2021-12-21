package main

import (
	"fmt"
	"sync/atomic"

	"github.com/gelfand/log"
)

// handleErr wraps error with given messages, logs it and atomicly stores as cached http response.
func handleErr(v *atomic.Value, msg string, err error) {
	log.Error(msg, "err", err)

	err = fmt.Errorf(msg+": %w", err)

	v.Store([]byte(err.Error()))
}
