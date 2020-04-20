package goawait

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
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

// wrapResult is a helper function to wrap a PollResultFunc into a PollFunc.
// The result is put into result parameter.
func wrapResult(poll PollResultFunc, result *interface{}) PollFunc {
	return func(ctx context.Context) error {
		var err error
		*result, err = poll(ctx)
		if err != nil {
			return err
		}
		return nil
	}
}

// PollResult calls the poll function every retry interval until the poll
// function succeeds or the context times out.
// It returns the result in case of a success execution
func PollResult(ctx context.Context, retryInterval time.Duration, poll PollResultFunc) (interface{}, error) {
	var result interface{}

	if err := Poll(ctx, retryInterval, wrapResult(poll, &result)); err != nil {
		return nil, err
	}
	return result, nil
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

	start := time.Now()
	cancel := make(chan interface{})
	firstResult := make(chan FirstResult)
	defer close(firstResult)
	errs := make(Errors)
	var errsLock sync.Mutex
	wrapCancellation := func(name string, poll PollResultFunc) PollResultFunc {
		return func(ctx context.Context) (interface{}, error) {
			select {
			case <-cancel:
				return nil, nil
			default:
				return poll(ctx)
			}
		}
	}

	var results int32

	for name, poll := range polls {
		name, poll := name, poll
		go func() {
			result, err := PollResult(ctx, retryTime, wrapCancellation(name, poll))

			if err != nil {
				errsLock.Lock()
				select {
				case <-ctx.Done():
					// do not write error on timeout, fixes possible map race
				default:
					errs[name] = err.(*Error)
				}
				errsLock.Unlock()
				return
			}
			if atomic.AddInt32(&results, 1) == 1 {
				firstResult <- FirstResult{Name: name, Result: result}
			}
		}()
	}

	select {
	case <-ctx.Done():
		errsLock.Lock()
		for name := range polls {
			if _, ok := errs[name]; !ok {
				errs[name] = &Error{errPoll: errors.New("timed out"), since: time.Since(start)}
			}
		}
		errsLock.Unlock()
		return FirstResult{}, errs
	case result := <-firstResult:
		close(cancel)
		return result, nil
	}
}
