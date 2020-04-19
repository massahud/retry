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
	start   time.Time
	end     time.Time
}

// Error implements the error interface and returns information about
// the timeout error.
func (e *TimeoutError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("context cancelled after %s: %s", e.end.Sub(e.start), e.ErrPoll)
	}
	return fmt.Sprintf("context cancelled after %s", e.end.Sub(e.start))
}

// Unwrap returns the context error, if any
func (e *TimeoutError) Unwrap() error {
	return e.Err
}

// UntilNoError calls the poll function every retry interval until the poll
// function succeeds or the context times out.
func UntilNoError(ctx context.Context, retryInterval time.Duration, poll func(ctx context.Context) error) error {
	start := time.Now()

	if ctx.Err() != nil {
		return &TimeoutError{Err: ctx.Err(), ErrPoll: nil, start: start, end: time.Now()}
	}

	var retry *time.Timer

	for {
		err := poll(ctx)
		if err == nil {
			return nil
		}

		if ctx.Err() != nil {
			return &TimeoutError{Err: ctx.Err(), ErrPoll: err, start: start, end: time.Now()}
		}

		if retry == nil {
			retry = time.NewTimer(retryInterval)
		}

		select {
		case <-ctx.Done():
			retry.Stop()
			return &TimeoutError{Err: ctx.Err(), ErrPoll: err, start: start, end: time.Now()}
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
