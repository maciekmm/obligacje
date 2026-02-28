package periodical

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestLoader_InitializesOnStart(t *testing.T) {
	loader, err := NewLoader(10*time.Minute, func() (int, error) {
		return 42, nil
	}, ErrBehaviorKeepOld)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	defer loader.Stop()

	val, err := loader.Current()
	if err != nil {
		t.Errorf("expected no error after start, got: %v", err)
	}
	if val != 42 {
		t.Errorf("expected 42, got: %d", val)
	}
}

func TestLoader_InitialLoadFailsReturnsError(t *testing.T) {
	_, err := NewLoader(10*time.Minute, func() (string, error) {
		return "", errors.New("load error")
	}, ErrBehaviorKeepOld)
	if err == nil {
		t.Fatal("expected error when initial load fails")
	}
}

func TestLoader_ConcurrentReads(t *testing.T) {
	var loadCount atomic.Int32
	loadCount.Store(1)
	loader, err := NewLoader(10*time.Millisecond, func() (int, error) {
		return int(loadCount.Add(1)), nil
	}, ErrBehaviorKeepOld)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer loader.Stop()

	// Spawn multiple concurrent readers
	const numReaders = 100
	var wg sync.WaitGroup
	errs := make(chan error, numReaders)

	for i := 0; i < numReaders; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				val, err := loader.Current()
				if err != nil {
					errs <- err
					return
				}
				if val < 1 {
					errs <- errors.New("got value less than 1")
					return
				}
			}
		}()
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		t.Errorf("concurrent read failed: %v", err)
	}
}

func TestLoader_ConcurrentReadsWhileUpdating(t *testing.T) {
	var counter atomic.Int64
	loader, err := NewLoader(5*time.Millisecond, func() (int64, error) {
		return counter.Add(1), nil
	}, ErrBehaviorKeepOld)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer loader.Stop()

	const numReaders = 50
	const readIterations = 200
	var wg sync.WaitGroup
	readErrors := make(chan error, numReaders)

	// start concurrent readers while loader is updating in background
	for i := 0; i < numReaders; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var lastVal int64
			for j := 0; j < readIterations; j++ {
				val, err := loader.Current()
				if err != nil {
					readErrors <- err
					return
				}
				// Values should be monotonically increasing (or same)
				if val < lastVal {
					readErrors <- errors.New("value decreased unexpectedly")
					return
				}
				lastVal = val
				time.Sleep(time.Microsecond)
			}
		}()
	}

	wg.Wait()
	close(readErrors)

	for err := range readErrors {
		t.Errorf("concurrent read during update failed: %v", err)
	}
}

func TestLoader_ErrBehaviorKeepOld(t *testing.T) {
	var callCount atomic.Int32
	loader, err := NewLoader(10*time.Millisecond, func() (string, error) {
		n := callCount.Add(1)
		if n == 1 {
			return "initial", nil
		}
		return "", errors.New("load error")
	}, ErrBehaviorKeepOld)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer loader.Stop()

	val, err := loader.Current()
	if err != nil {
		t.Fatalf("expected no error after first load, got: %v", err)
	}
	if val != "initial" {
		t.Errorf("expected 'initial', got: %s", val)
	}

	// Wait for error load (should keep old value)
	time.Sleep(15 * time.Millisecond)

	val, err = loader.Current()
	if err != nil {
		t.Errorf("expected old value to be kept on error, got error: %v", err)
	}
	if val != "initial" {
		t.Errorf("expected 'initial' to be kept, got: %s", val)
	}
}

func TestLoader_ErrBehaviorReturnError(t *testing.T) {
	var callCount atomic.Int32
	loader, err := NewLoader(40*time.Millisecond, func() (string, error) {
		n := callCount.Add(1)
		if n == 1 {
			return "initial", nil
		}
		return "", errors.New("load error")
	}, ErrBehaviorReturnError)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer loader.Stop()

	// Wait for the background periodic load to fail and trigger ErrBehaviorReturnError
	time.Sleep(60 * time.Millisecond)

	_, err = loader.Current()
	if err == nil {
		t.Fatal("expected error to be returned by Current() after background load failure")
	}
	if err.Error() != "load error" {
		t.Errorf("expected 'load error', got: %v", err)
	}
}
