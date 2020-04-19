// Copyright 2020 Geraldo Augusto Massahud Rodrigues dos Santos
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// GoAwait is a simple module for asynchronous waiting.
//
// Use goawait when you need to wait for asynchronous tasks to complete before continuing normal
// execution. It is very useful for waiting on integration and end to end tests.
//
// Polling
//
// GoAwait has polling functions that polls for something until it happens, or until the context
// is canceled.
//
//     goawait.UntilNoError(cancelCtx, 500 * time.Millisecond, connectToDatabase)
//
//     goawait.Untiltrue(cancelCtx, 500 * time.Millisecond, messageReceived)
//
// The polling functions are based on Bill Kennedy's retryTimeout concurrency example(https://github.com/ardanlabs/gotraining/blob/master/topics/go/concurrency/channels/example1/example1.go)
package goawait

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// TimeoutError informs that awaiting was cancelled before poll function returned desired
// output
type TimeoutError struct {
	ctx       context.Context
	start     time.Time
	end       time.Time
	lastError error
}

func (e *TimeoutError) Error() string {
	if e.ctx.Err() != nil {
		return fmt.Sprintf("context cancelled after %s: %s", e.end.Sub(e.start).String(), e.ctx.Err().Error())
	}
	return fmt.Sprintf("context cancelled after %s", e.end.Sub(e.start).String())
}

// Unwrap returns the context error, if any
func (e *TimeoutError) Unwrap() error {
	return e.ctx.Err()
}

// LastError returns the last error returned by the polling function, if any
func (e *TimeoutError) LastError() error {
	return e.lastError
}

// UntilNoError retries the poll function every "retryTime" until it returns nil or the context is done
//
// Returns TimeoutError if context is done before poll is true.
func UntilNoError(ctx context.Context, retryTime time.Duration, poll func(ctx context.Context) error) error {
	start := time.Now()

	select {
	case <-ctx.Done():
		return &TimeoutError{ctx: ctx, start: start, end: start}
	default:
		retry := time.NewTimer(retryTime)
		var err error
		for {
			err = poll(ctx)

			if err == nil {
				return nil
			}

			select {
			case <-ctx.Done():
				retry.Stop()
				return &TimeoutError{ctx: ctx, start: start, end: time.Now(), lastError: err}
			case <-retry.C:
				retry.Reset(retryTime)
			}
		}
	}
}

// UntilTrue retries the poll function every "retryTime" until it returns true or the context is done
//
// Returns TimeoutError if context is done before poll is true.
func UntilTrue(ctx context.Context, retryTime time.Duration, poll func(ctx context.Context) bool) error {
	boolWrap := func(ctx context.Context) error {
		if poll(ctx) {
			return nil
		}
		return errors.New("")
	}
	err := UntilNoError(ctx, retryTime, boolWrap)
	if err != nil {
		err.(*TimeoutError).lastError = nil
	}
	return err
}
