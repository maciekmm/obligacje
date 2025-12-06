package periodicaldata

import (
	"errors"
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

func NewLoader[T any](interval time.Duration, load LoadFunc[T], errBehavior ErrBehavior) *Loader[T] {
	return &Loader[T]{
		interval:    interval,
		load:        load,
		ticker:      time.NewTicker(interval),
		errBehavior: errBehavior,
	}
}

func (l *Loader[T]) Start() {
	go func() {
		l.loadAndSet()
		for range l.ticker.C {
			l.loadAndSet()
		}
	}()
}

func (l *Loader[T]) loadAndSet() {
	new, err := l.load()
	if err != nil {
		if l.errBehavior == ErrBehaviorReturnError {
			slog.Error("failed to load data", "err", err)
			l.current.Store(value[T]{err: err})
		}
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
