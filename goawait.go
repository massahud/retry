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

type namedResult struct {
	name string
	Result
}

func poll(ctx context.Context, retry time.Duration, name string, poll PollFunc) namedResult {
	var timer *time.Timer
	start := time.Now()

	for {
		value, err := poll(ctx)
		if err == nil {
			return namedResult{name: name, Result: Result{Value: value}}
		}

		if ctx.Err() != nil {
			return namedResult{name: name, Result: Result{Err: &Error{errPoll: err, since: time.Since(start)}}}
		}

		if timer == nil {
			timer = time.NewTimer(retry)
		}

		select {
		case <-ctx.Done():
			timer.Stop()
			return namedResult{name: name, Result: Result{Err: &Error{errPoll: err, since: time.Since(start)}}}
		case <-timer.C:
			timer.Reset(retry)
		}
	}
}

func pollMap(ctx context.Context, retry time.Duration, pollers map[string]PollFunc) <-chan namedResult {
	rc := make(chan namedResult, len(pollers))
	go func() {
		wg := sync.WaitGroup{}

		wg.Add(len(pollers))
		for n, f := range pollers {
			n, f := n, f
			go func() { rc <- poll(ctx, retry, n, f); wg.Done() }()
		}
		wg.Wait()
		close(rc)
	}()

	return rc
}

// Poll calls the poll function every retry interval until the poll
// function succeeds or the context times out.
func Func(ctx context.Context, retryInterval time.Duration, pf PollFunc) Result {
	res := poll(ctx, retryInterval, "", pf)
	return res.Result
}

// PollAll calls all the poll functions every retry interval until the poll
// functions succeeds or the context times out.
func All(ctx context.Context, retryTime time.Duration, polls map[string]PollFunc) map[string]Result {
	g := len(polls)
	var wg sync.WaitGroup
	wg.Add(g)

	results := map[string]Result{}

	// iterate over pollMap() results and add each to our results map.
	for result := range pollMap(ctx, retryTime, polls) {
		results[result.name] = result.Result
	}

	return results
}

// PollFirst calls all the poll functions every retry interval until the poll
// functions succeeds or the context times out. Once the first polling function
// the succeeds returns, this function will return that result.
func First(ctx context.Context, retryTime time.Duration, polls map[string]PollFunc) Result {
	start := time.Now()

	// iterate over results of pollMap() and return the first value
	// from a successful poll.
	for result := range pollMap(ctx, retryTime, polls) {
		if result.Result.Err != nil {
			continue
		}
		return result.Result
	}

	// at this point, no poller ended successfully so return a boilerplate error.
	return Result{Err: &Error{errPoll: errors.New("all poll functions failed"), since: time.Since(start)}}
}
