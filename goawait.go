package goawait

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// TimeoutError informs that a cancellation took place before the poll
// function returned successfully.
type TimeoutError struct {
	Err     error
	ErrPoll error
	since   time.Duration
}

// Error implements the error interface and returns information about
// the timeout error.
func (te *TimeoutError) Error() string {
	if te.Err != nil {
		return fmt.Sprintf("context cancelled after %v: %s", te.since, te.ErrPoll)
	}
	return fmt.Sprintf("context cancelled after %v", te.since)
}

// Unwrap returns the context error, if any
func (te *TimeoutError) Unwrap() error {
	return te.Err
}

// UntilNoError calls the poll function every retry interval until the poll
// function succeeds or the context times out.
func UntilNoError(ctx context.Context, retryInterval time.Duration, poll func(ctx context.Context) error) error {
	start := time.Now()

	if ctx.Err() != nil {
		return &TimeoutError{Err: ctx.Err(), ErrPoll: nil, since: time.Since(start)}
	}

	var retry *time.Timer

	for {
		err := poll(ctx)
		if err == nil {
			return nil
		}

		if ctx.Err() != nil {
			return &TimeoutError{Err: ctx.Err(), ErrPoll: err, since: time.Since(start)}
		}

		if retry == nil {
			retry = time.NewTimer(retryInterval)
		}

		select {
		case <-ctx.Done():
			retry.Stop()
			return &TimeoutError{Err: ctx.Err(), ErrPoll: err, since: time.Since(start)}
		case <-retry.C:
			retry.Reset(retryInterval)
		}
	}
}

// UntilTrue calls the poll function every retry interval until the poll
// function succeeds or the context times out.
func UntilTrue(ctx context.Context, retryTime time.Duration, poll func(ctx context.Context) bool) error {
	f := func(ctx context.Context) error {
		if poll(ctx) {
			return nil
		}
		return errors.New("poll failed")
	}

	return UntilNoError(ctx, retryTime, f)
}
