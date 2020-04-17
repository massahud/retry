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
// To use it just create a spec and call one of it's Until methods
//
//     func receivedMessage() bool { ... }
//
//     goawait.AtMost(10 * time.Second).
//         WithRetryTimeout(200 * time.Millisecond).
//         Until(receivedMessage)
//
// GoAwait is based on Java's Awaitility's DSL: https://github.com/awaitility/awaitility
// The polling functions were based on Bill Kennedy's **retryTimeout** concurrency example at
// https://github.com/ardanlabs/gotraining/blob/0728ec842fbde65115e1a0a255b62b4a93d4c6a8/topics/go/concurrency/channels/example1/example1.go#L290
package goawait

import (
	"fmt"
	"time"
)

// DefaultRetryTime: 1 seconds
var defaultRetryTime = 1 * time.Second

// Spec is the GoAwait specification
type Spec struct {
	maxWait   time.Duration
	retryTime time.Duration
}

// Timeout error
type TimeoutError struct {
	Spec      Spec
	LastError error
}

// Timeout error string
func (e *TimeoutError) Error() string {
	return fmt.Sprintf("timed out after %s", e.Spec.maxWait)
}

// Unwraps the last error received before timeout
func (e *TimeoutError) Unwrap() error {
	return e.LastError
}

// AtMost creates a new Spec with a specified timeout and default retry time of 1 second
func AtMost(maxWait time.Duration) Spec {
	return Spec{maxWait: maxWait, retryTime: defaultRetryTime}
}

// WithRetryInterval configures the Spec retryTime
func (spec Spec) WithRetryInterval(retryTime time.Duration) Spec {
	spec.retryTime = retryTime
	return spec
}

// Until executes the polling function until the poll function returns true, or a timeout occurs
// It returns a TimeoutError on timeout.
func (spec Spec) Until(poll func() bool) error {
	timeout := time.NewTimer(spec.maxWait)
	retry := time.NewTimer(spec.retryTime)

	for {
		if poll() {
			return nil
		}

		retry.Reset(spec.retryTime)

		select {
		case <-timeout.C:
			retry.Stop()
			return &TimeoutError{
				Spec:      spec,
				LastError: nil,
			}
		case <-retry.C:
		}
	}
}

// UntilNoError executes the polling function until it does not return an error.
// It returns a TimeoutError on timeout.
func (spec Spec) UntilNoError(poll func() error) error {
	timeout := time.NewTimer(spec.maxWait)
	retry := time.NewTimer(spec.retryTime)

	var err error
	for {
		if err = poll(); err == nil {
			return nil
		}

		retry.Reset(spec.retryTime)

		select {
		case <-timeout.C:
			retry.Stop()
			return &TimeoutError{
				Spec:      spec,
				LastError: err,
			}
		case <-retry.C:
		}
	}
}