package await

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

// Func calls the poll function every retry interval until the poll
// function succeeds or the context times out.
func Func(ctx context.Context, retryInterval time.Duration, pollFn PollFunc) Result {
	return poll(ctx, retryInterval, pollFn)
}

// All calls all the poll functions every retry interval until the poll
// functions succeeds or the context times out.
func All(ctx context.Context, retryTime time.Duration, pollFns map[string]PollFunc) map[string]Result {
	results := make(map[string]Result)
	for result := range pollMap(ctx, retryTime, pollFns) {
		results[result.name] = result.Result
	}

	return results
}

// First calls all the poll functions every retry interval until the poll
// functions succeeds or the context times out. Once the first polling function
// succeeds, this function will return that result.
func First(ctx context.Context, retryTime time.Duration, pollFns map[string]PollFunc) Result {
	start := time.Now()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	for result := range pollMap(ctx, retryTime, pollFns) {
		if result.Result.Err != nil {
			continue
		}
		return result.Result
	}

	return Result{Err: &Error{errPoll: errors.New("all poll functions failed"), since: time.Since(start)}}
}

// poll calls the poll function every retry interval until the poll
// function succeeds or the context times out.
func poll(ctx context.Context, retryInterval time.Duration, pollFn PollFunc) Result {
	var retry *time.Timer
	start := time.Now()

	for {
		value, err := pollFn(ctx)
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

type namedResult struct {
	name string
	Result
}

// pollMap calls the map of poll functions every retry interval until the poll
// function succeeds or the context times out. As poll functions complete, their
// results are signaled over the channel for processing.
func pollMap(ctx context.Context, retry time.Duration, pollFns map[string]PollFunc) <-chan namedResult {
	g := len(pollFns)
	results := make(chan namedResult, g)

	go func() {
		var wg sync.WaitGroup
		wg.Add(g)
		for name, pollFn := range pollFns {
			name, pollFn := name, pollFn
			go func() {
				defer wg.Done()
				result := poll(ctx, retry, pollFn)
				results <- namedResult{name: name, Result: result}
			}()
		}
		wg.Wait()
		close(results)
	}()

	return results
}
