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
// See the License for the Awaitific language governing permissions and
// limitations under the License.

package goawait_test

import (
	"context"
	"errors"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/massahud/goawait"
	"github.com/stretchr/testify/assert"
)

func TestAwait_UntilNoError(t *testing.T) {
	t.Run("should retry until the poll function does not return error", func(t *testing.T) {
		var calls int
		err := goawait.AtMost(1 * time.Second).RetryingEvery(1 * time.Nanosecond).UntilNoError(func(_ context.Context) error {
			calls++
			if calls == 3 {
				return nil
			}
			return errors.New("foo")
		})
		if assert.NoError(t, err) {
			assert.Equal(t, 3, calls)
		}
	})

	t.Run("should return a TimeoutError if max time is reached", func(t *testing.T) {
		var calls int
		err := goawait.AtMost(1 * time.Millisecond).
			RetryingEvery(1 * time.Nanosecond).
			UntilNoError(func(_ context.Context) error {
				calls++
				return fmt.Errorf("foo %d", calls)
			})
		if assert.Error(t, err) {
			assert.IsType(t, &goawait.TimeoutError{}, err)
			assert.Equal(t, fmt.Errorf("foo %d", calls), err.(*goawait.TimeoutError).LastError())
			assert.Greater(t, calls, 0)
		}
	})

	t.Run("should pass the await context to the poll function", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "foo", "bar")
		var calls int
		err := goawait.WithContext(ctx).AtMost(1 * time.Millisecond).RetryingEvery(1 * time.Nanosecond).UntilNoError(func(receivedCtx context.Context) error {
			calls++
			assert.Equal(t, ctx, receivedCtx)
			return nil
		})
		if assert.NoError(t, err) {
			assert.EqualValues(t, 1, calls)
		}
	})

	t.Run("should return TimeoutError when context is cancelled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		err := goawait.WithContext(ctx).
			UntilNoError(func(ctx context.Context) error {
				cancel()
				return errors.New("foo")
			})

		if assert.Error(t, err) {
			assert.IsType(t, &goawait.TimeoutError{}, err)
			assert.Equal(t, "context canceled", errors.Unwrap(err).Error())
		}
	})

	t.Run("should not call function when context is cancelled before running", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		var calls int
		err := goawait.WithContext(ctx).
			UntilNoError(func(ctx context.Context) error {
				calls++
				return errors.New("foo")
			})

		if assert.Error(t, err) {
			assert.Zero(t, calls)
		}
	})
}

func TestAwait_UntilTrue(t *testing.T) {
	t.Run("should retry until the poll function returns true", func(t *testing.T) {
		var calls int
		err := goawait.AtMost(1 * time.Second).RetryingEvery(1 * time.Nanosecond).UntilTrue(func(ctx context.Context) bool {
			calls++
			if calls == 3 {
				return true
			}
			return false
		})
		if assert.NoError(t, err) {
			assert.Equal(t, 3, calls)
		}
	})

	t.Run("should return a TimeoutError if max time is reached", func(t *testing.T) {
		var calls int
		err := goawait.AtMost(1 * time.Millisecond).RetryingEvery(1 * time.Nanosecond).UntilTrue(func(_ context.Context) bool {
			calls++
			return false
		})
		if assert.Error(t, err) {
			assert.IsType(t, &goawait.TimeoutError{}, err)
			assert.Nil(t, err.(*goawait.TimeoutError).LastError())
			assert.Greater(t, calls, 0)
		}
	})

	t.Run("should pass the await context to the poll function", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "foo", "bar")
		var calls int
		err := goawait.WithContext(ctx).AtMost(1 * time.Millisecond).RetryingEvery(1 * time.Nanosecond).UntilTrue(func(receivedCtx context.Context) bool {
			calls++
			assert.Equal(t, ctx, receivedCtx)
			return true
		})
		if assert.NoError(t, err) {
			assert.EqualValues(t, 1, calls)
		}
	})

	t.Run("should return TimeoutError when context is cancelled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		err := goawait.WithContext(ctx).
			UntilTrue(func(ctx context.Context) bool {
				cancel()
				return false
			})

		if assert.Error(t, err) {
			assert.IsType(t, &goawait.TimeoutError{}, err)
			assert.Equal(t, "context canceled", errors.Unwrap(err).Error())
		}
	})

	t.Run("should not call function when context is cancelled before running", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		var calls int
		err := goawait.WithContext(ctx).
			UntilTrue(func(ctx context.Context) bool {
				calls++
				return false
			})

		if assert.Error(t, err) {
			assert.Zero(t, calls)
		}
	})
}

func ExampleAwait_UntilNoError() {

	// func connectToDatabase(ctx context.Context) error { ... }

	err := goawait.AtMost(1 * time.Second).
		RetryingEvery(5 * time.Millisecond).
		UntilNoError(connectToDatabase)

	if err != nil {
		log.Fatalf("database not ready: %s", err.Error())
	}

	fmt.Println("Database ready")

	// Output: Database ready
}

func ExampleAwait_UntilTrue() {

	// func pageHasItem(ctx context.Context) bool { ... }

	err := goawait.AtMost(1 * time.Second).
		RetryingEvery(5 * time.Millisecond).
		UntilTrue(pageHasItem)

	if err != nil {
		log.Fatalf("page does not have item")
	}

	fmt.Println( "page has item, continuing...")

	// Output: page has item, continuing...
}
