package periodical

import (
	"errors"
	"fmt"
	"log/slog"
	"sync/atomic"
	"time"
)

type ErrBehavior int

const (
	ErrBehaviorKeepOld ErrBehavior = iota
	ErrBehaviorReturnError
)

var (
	ErrNotInitialized = errors.New("not initialized yet")
)

type LoadFunc[T any] func() (T, error)

type value[T any] struct {
	data T
	err  error
}

type Loader[T any] struct {
	current atomic.Value

	interval    time.Duration
	load        LoadFunc[T]
	ticker      *time.Ticker
	errBehavior ErrBehavior
}

// NewLoader creates a new Loader that periodically calls load to refresh data.
// The initial load is performed synchronously. If it fails, an error is returned
// and the Loader is not started.
func NewLoader[T any](interval time.Duration, load LoadFunc[T], errBehavior ErrBehavior) (*Loader[T], error) {
	loader := &Loader[T]{
		interval:    interval,
		load:        load,
		errBehavior: errBehavior,
	}

	data, err := load()
	if err != nil {
		return nil, fmt.Errorf("initial load failed: %w", err)
	}
	loader.current.Store(value[T]{data: data})

	loader.ticker = time.NewTicker(interval)
	go func() {
		for range loader.ticker.C {
			loader.loadAndSet()
		}
	}()

	return loader, nil
}

func (l *Loader[T]) loadAndSet() {
	new, err := l.load()
	if err != nil {
		if l.errBehavior == ErrBehaviorReturnError {
			slog.Error("failed to load new data", "err", err)
			l.current.Store(value[T]{err: err})
			return
		}
		slog.Error("failed to load new data, keeping old", "err", err)
		return
	}

	l.current.Store(value[T]{data: new})
}

func (l *Loader[T]) Stop() {
	l.ticker.Stop()
}

func (l *Loader[T]) Current() (T, error) {
	v, ok := l.current.Load().(value[T])
	if !ok {
		return *new(T), ErrNotInitialized
	}

	return v.data, v.err
}
