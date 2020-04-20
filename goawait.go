package goawait

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

// PollFunc is a polling function that returns no error when succeeds
type PollFunc func(ctx context.Context) error

// PollResultFunc is a polling function that returns some result and no error when succeeds
type PollResultFunc func(ctx context.Context) (interface{}, error)

// FirstResult holds the first received result on PollFirstResult function
type FirstResult struct {
	Name   string
	Result interface{}
}

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

// Errors provides errors for a set of poll functions that are
// executed together as a group.
type Errors map[string]*Error

// Error implements the error interface
func (errs Errors) Error() string {
	var b strings.Builder
	for name, err := range errs {
		b.WriteString(fmt.Sprintf("%s: %v\n", name, err))
	}
	return b.String()
}

// Poll calls the poll function every retry interval until the poll
// function succeeds or the context times out.
func Poll(ctx context.Context, retryInterval time.Duration, poll PollFunc) error {
	var retry *time.Timer
	start := time.Now()

	for {
		err := poll(ctx)
		if err == nil {
			return nil
		}

		if ctx.Err() != nil {
			return &Error{errPoll: err, since: time.Since(start)}
		}

		if retry == nil {
			retry = time.NewTimer(retryInterval)
		}

		select {
		case <-ctx.Done():
			retry.Stop()
			return &Error{errPoll: err, since: time.Since(start)}
		case <-retry.C:
			retry.Reset(retryInterval)
		}
	}
}

// PollAll calls all the poll functions every retry interval until the poll
// functions succeeds or the context times out.
func PollAll(ctx context.Context, retryTime time.Duration, polls map[string]PollFunc) error {
	g := len(polls)
	var wg sync.WaitGroup
	wg.Add(g)

	var mapLock sync.Mutex
	errs := make(Errors)
	for name, poll := range polls {
		name, poll := name, poll
		go func() {
			defer wg.Done()
			if err := Poll(ctx, retryTime, poll); err != nil {
				mapLock.Lock()
				errs[name] = err.(*Error)
				mapLock.Unlock()
			}
		}()
	}

	wg.Wait()

	if len(errs) > 0 {
		return errs
	}
	return nil
}

// PollFirstResult polls all functions until at least one succeeds or the context times out.
func PollFirstResult(ctx context.Context, retryTime time.Duration, polls map[string]PollResultFunc) (FirstResult, error) {

	cancel := make(chan interface{})

	var firstResult FirstResult
	var onlyOnce sync.Once
	wrapCancellation := func(name string, poll PollResultFunc) PollFunc {
		return func(ctx context.Context) error {
			select {
			case <-cancel:
				return nil
			default:
				result, err := poll(ctx)
				if err != nil {
					return err
				}
				onlyOnce.Do(func() {
					close(cancel)
					firstResult.Name = name
					firstResult.Result = result
				})
				return nil
			}
		}
	}

	wrappedPolls := make(map[string]PollFunc)

	for name, poll := range polls {
		wrappedPolls[name] = wrapCancellation(name, poll)
	}

	err := PollAll(ctx, retryTime, wrappedPolls)

	return firstResult, err
}
