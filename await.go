package await

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

// Worker is a function that performs work and returns no error when succeeds.
type Worker func(ctx context.Context) (interface{}, error)

// Result is what is returned from the api for any worker function call.
type Result struct {
	Value interface{}
	Err   error
}

// Error informs that a cancellation took place before the worker
// function returned successfully.
type Error struct {
	errWork error
	since   time.Duration
}

// Error implements the error interface and returns information about
// the timeout error.
func (err *Error) Error() string {
	if err.errWork != nil {
		return fmt.Sprintf("context cancelled after %v : %s", err.since, err.errWork)
	}
	return fmt.Sprintf("context cancelled after %v", err.since)
}

// Unwrap returns the context error, if any
func (err *Error) Unwrap() error {
	return err.errWork
}

// Func calls the worker function every retry interval until the worker
// function succeeds or the context times out.
func Func(ctx context.Context, retryInterval time.Duration, worker Worker) Result {
	var retry *time.Timer
	start := time.Now()

	for {
		value, err := worker(ctx)
		if err == nil {
			return Result{Value: value}
		}

		if ctx.Err() != nil {
			return Result{Err: &Error{errWork: err, since: time.Since(start)}}
		}

		if retry == nil {
			retry = time.NewTimer(retryInterval)
		}

		select {
		case <-ctx.Done():
			retry.Stop()
			return Result{Err: &Error{errWork: err, since: time.Since(start)}}
		case <-retry.C:
			retry.Reset(retryInterval)
		}
	}
}

type namedResult struct {
	name string
	Result
}

// workerMap calls the map of worker functions every retry interval until the worker
// function succeeds or the context times out. As worker functions complete, their
// results are signaled over the channel for processing.
func workerMap(ctx context.Context, retry time.Duration, workers map[string]Worker) <-chan namedResult {
	g := len(workers)
	results := make(chan namedResult, g)

	go func() {
		var wg sync.WaitGroup
		wg.Add(g)
		for name, worker := range workers {
			name, worker := name, worker
			go func() {
				defer wg.Done()
				result := Func(ctx, retry, worker)
				results <- namedResult{name: name, Result: result}
			}()
		}
		wg.Wait()
		close(results)
	}()

	return results
}

// All calls all the worker functions every retry interval until the worker
// functions succeeds or the context times out.
func All(ctx context.Context, retryTime time.Duration, workers map[string]Worker) map[string]Result {
	results := make(map[string]Result)
	for result := range workerMap(ctx, retryTime, workers) {
		results[result.name] = result.Result
	}

	return results
}

// First calls all the worker functions every retry interval until the worker
// functions succeeds or the context times out. Once the first worker function
// succeeds, this function will return that result.
func First(ctx context.Context, retryTime time.Duration, workers map[string]Worker) Result {
	start := time.Now()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	for result := range workerMap(ctx, retryTime, workers) {
		if result.Result.Err != nil {
			continue
		}
		return result.Result
	}

	return Result{Err: &Error{errWork: errors.New("all worker functions failed"), since: time.Since(start)}}
}
