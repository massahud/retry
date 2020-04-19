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

func ExampleAwait_UntilTrue() {

	var message string

	t := time.NewTimer(30 * time.Millisecond)
	go func() {
		<-t.C
		message = "Hello, async World"
	}()

	receivedMessage := func(ctx context.Context) bool {
		if message == "" {
			return false
		}
		fmt.Printf("Received message: %s", message)
		return true
	}

	err := goawait.AtMost(1 * time.Second).
		RetryingEvery(5 * time.Millisecond).
		UntilTrue(receivedMessage)

	if err != nil {
		log.Fatal(err.Error())
	}

	// Output: Received message: Hello, async World
}

func ExampleAwait_UntilNoError() {

	var message string

	t := time.NewTimer(30 * time.Millisecond)
	go func() {
		<-t.C
		message = "Hello, async World"
	}()

	getMessage := func(ctx context.Context) error {
		if message == "" {
			return fmt.Errorf("404, no message")
		}
		fmt.Printf("Got message: %s", message)
		return nil
	}

	err := goawait.AtMost(1 * time.Second).
		RetryingEvery(5 * time.Millisecond).
		UntilNoError(getMessage)

	if err != nil {
		log.Fatal(err.Error())
	}

	// Output: Got message: Hello, async World
}

func TestAwait_UntilTrue(t *testing.T) {
	t.Run("should retry until the poll function returns true", func(t *testing.T) {
		var called int
		err := goawait.AtMost(1 * time.Second).RetryingEvery(1 * time.Nanosecond).UntilTrue(func(ctx context.Context) bool {
			called++
			if called == 3 {
				return true
			}
			return false
		})
		if assert.NoError(t, err) {
			assert.Equal(t, 3, called)
		}
	})

	t.Run("should return a TimeoutError if max time is reached", func(t *testing.T) {
		var called int
		err := goawait.AtMost(1 * time.Millisecond).RetryingEvery(1 * time.Nanosecond).UntilTrue(func(_ context.Context) bool {
			called++
			return false
		})
		if assert.Error(t, err) {
			assert.IsType(t, &goawait.TimeoutError{}, err)
			assert.Nil(t, err.(*goawait.TimeoutError).LastError())
			assert.Greater(t, called, 0)
		}
	})
}

func TestAwait_UntilNoError(t *testing.T) {
	t.Run("should retry until the poll function does not return error", func(t *testing.T) {
		var called int
		err := goawait.AtMost(1 * time.Second).RetryingEvery(1 * time.Nanosecond).UntilNoError(func(_ context.Context) error {
			called++
			if called == 3 {
				return nil
			}
			return errors.New("foo")
		})
		if assert.NoError(t, err) {
			assert.Equal(t, 3, called)
		}
	})

	t.Run("should return a TimeoutError if max time is reached", func(t *testing.T) {
		var called int
		err := goawait.AtMost(1 * time.Millisecond).
			RetryingEvery(1 * time.Nanosecond).
			UntilNoError(func(_ context.Context) error {
				called++
				return fmt.Errorf("foo %d", called)
			})
		if assert.Error(t, err) {
			assert.IsType(t, &goawait.TimeoutError{}, err)
			assert.Equal(t, fmt.Errorf("foo %d", called), err.(*goawait.TimeoutError).LastError())
			assert.Greater(t, called, 0)
		}
	})
}
