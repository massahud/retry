package goawait

import (
	"context"
	"errors"
	"fmt"
	"time"
)

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

// Poll calls the poll function every retry interval until the poll
// function succeeds or the context times out.
func Poll(ctx context.Context, retryInterval time.Duration, poll func(ctx context.Context) error) error {
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
func PollBool(ctx context.Context, retryTime time.Duration, poll func(ctx context.Context) bool) error {
	pollErr := errors.New("poll failed")
	f := func(ctx context.Context) error {
		if poll(ctx) {
			return nil
		}
		return pollErr
	}

	return Poll(ctx, retryTime, f)
}
