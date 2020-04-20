package goawait

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

// PollFunc is a polling function that returns no error when succeeds
type PollFunc func(ctx context.Context) error

// PollBoolFunc is a polling function that returns true when succeeds
type PollBoolFunc func(ctx context.Context) bool

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
		return fmt.Sprintf("context cancelled after %v : poll error %s", err.since, err.errPoll)
	}
	return fmt.Sprintf("context cancelled after %v", err.since)
}

// Unwrap returns the context error, if any
func (err *Error) Unwrap() error {
	return err.errPoll
}

// Errors informs that a cancellation took place before a list of poll
// functions returned successfully.
type Errors []*Error

// Error implements the error interface
func (err Errors) Error() string {
	return "context cancelled before the polling succeeded"
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

// PollBool calls the poll function every retry interval until the poll
// function succeeds or the context times out.
func PollBool(ctx context.Context, retryTime time.Duration, poll PollBoolFunc) error {
	pollErr := errors.New("poll failed")
	f := func(ctx context.Context) error {
		if poll(ctx) {
			return nil
		}
		return pollErr
	}

	return Poll(ctx, retryTime, f)
}

// PollAll polls all functions until they all succeed or the context times out.
func PollAll(ctx context.Context, retryTime time.Duration, polls []PollFunc) error {

	errs := make(Errors, len(polls))

	var wg sync.WaitGroup
	for i := range polls {
		wg.Add(1)
		go func(pos int) {
			e := Poll(ctx, retryTime, polls[pos])
			if e != nil {
				if err, ok := e.(*Error); ok {
					errs[pos] = err
				}
			}
			wg.Done()
		}(i)
	}
	wg.Wait()

	for i := range errs {
		if errs[i] != nil {
			return errs
		}
	}
	return nil
}
