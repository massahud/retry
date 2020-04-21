package goawait

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

// Result is what is returned from the api for any poll function call.
type Result struct {
	Value interface{}
	Err   error
}

// PollFunc is a polling function that returns no error when succeeds
type PollFunc func(ctx context.Context) (interface{}, error)

// Error informs that a cancellation took place before the poll
// function returned successfully.
type Error struct {
	errPoll error
	since   time.Duration
}

// Error implements the error interface and returns information about
// the timeout error.
func (err *Error) Error() string {
	if err.errPoll != nil {
		return fmt.Sprintf("context cancelled after %v : %s", err.since, err.errPoll)
	}
	return fmt.Sprintf("context cancelled after %v", err.since)
}

// Unwrap returns the context error, if any
func (err *Error) Unwrap() error {
	return err.errPoll
}

// Poll calls the poll function every retry interval until the poll
// function succeeds or the context times out.
func Poll(ctx context.Context, retryInterval time.Duration, poll PollFunc) Result {
	var retry *time.Timer
	start := time.Now()

	for {
		value, err := poll(ctx)
		if err == nil {
			return Result{Value: value}
		}

		if ctx.Err() != nil {
			return Result{Err: &Error{errPoll: err, since: time.Since(start)}}
		}

		if retry == nil {
			retry = time.NewTimer(retryInterval)
		}

		select {
		case <-ctx.Done():
			retry.Stop()
			return Result{Err: &Error{errPoll: err, since: time.Since(start)}}
		case <-retry.C:
			retry.Reset(retryInterval)
		}
	}
}

// PollAll calls all the poll functions every retry interval until the poll
// functions succeeds or the context times out.
func PollAll(ctx context.Context, retryTime time.Duration, polls map[string]PollFunc) map[string]Result {
	g := len(polls)
	var wg sync.WaitGroup
	wg.Add(g)

	mutex := sync.Mutex{}
	results := make(map[string]Result)
	for name, poll := range polls {
		name, poll := name, poll
		go func() {
			defer wg.Done()

			mutex.Lock()
			results[name] = Poll(ctx, retryTime, poll)
			mutex.Unlock()
		}()
	}

	wg.Wait()

	return results
}

// PollFirst calls all the poll functions every retry interval until the poll
// functions succeeds or the context times out. Once the first polling function
// the succeeds returns, this function will return that result.
func PollFirst(ctx context.Context, retryTime time.Duration, polls map[string]PollFunc) Result {
	start := time.Now()
	g := len(polls)
	results := make(chan Result, g)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	for _, poll := range polls {
		poll := poll
		go func() {
			results <- Poll(ctx, retryTime, poll)
		}()
	}

	for i := 0; i < g; i++ {
		select {
		case <-ctx.Done():
			return Result{Err: &Error{since: time.Since(start)}}
		case result := <-results:
			if result.Err == nil {
				return result
			}
		}
	}

	return Result{Err: &Error{errPoll: errors.New("all poll functions failed"), since: time.Since(start)}}
}
