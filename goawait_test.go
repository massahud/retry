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

package goawait

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func ExampleSpec_Until() {

	var message string

	t := time.NewTimer(30 * time.Millisecond)
	go func() {
		<-t.C
		message = "Hello, async World"
	}()

	receivedMessage := func() bool {
		if message == "" {
			return false
		}
		fmt.Printf("Received message: %s", message)
		return true
	}

	AtMost(1 * time.Second).
		WithRetryInterval(5 * time.Millisecond).
		Until(receivedMessage)

	// Output: Received message: Hello, async World
}

func ExampleSpec_UntilNoError() {

	var message string

	t := time.NewTimer(30 * time.Millisecond)
	go func() {
		<-t.C
		message = "Hello, async World"
	}()

	getMessage := func() error {
		if message == "" {
			return fmt.Errorf("404, no message")
		}
		fmt.Printf("Got message: %s", message)
		return nil
	}

	AtMost(1 * time.Second).
		WithRetryInterval(5 * time.Millisecond).
		UntilNoError(getMessage)

	// Output: Got message: Hello, async World
}

func TestAtMost(t *testing.T) {
	t.Run("should create a spec with specified timeout and default retry time", func(t *testing.T) {
		spec := AtMost(12 * time.Second)
		assert.Equal(t, Spec{maxWait: 12 * time.Second, retryTime: defaultRetryTime}, spec)
	})
}

func TestSpec_WithRetryInterval(t *testing.T) {
	t.Run("should set the spec retry interval", func(t *testing.T) {
		spec := AtMost(12 * time.Second).WithRetryInterval(150 * time.Millisecond)
		assert.Equal(t, Spec{maxWait: 12 * time.Second, retryTime: 150 * time.Millisecond}, spec)
	})
}

func TestSpec_Until(t *testing.T) {
	t.Run("should retry until the poll function returns true", func(t *testing.T) {
		var called int
		err := AtMost(1 * time.Second).WithRetryInterval(1 * time.Nanosecond).Until(func() bool {
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
		await := AtMost(1 * time.Millisecond).WithRetryInterval(1 * time.Nanosecond)
		err := await.Until(func() bool {
			called++
			return false
		})
		if assert.Error(t, err) {
			assert.IsType(t, &TimeoutError{}, err)
			assert.Equal(t, await, err.(*TimeoutError).Spec)
			assert.Nil(t, err.(*TimeoutError).LastError)
			assert.Greater(t, called, 0)
		}
	})
}

func TestSpec_UntilNoError(t *testing.T) {
	t.Run("should retry until the poll function does not return error", func(t *testing.T) {
		var called int
		err := AtMost(1 * time.Second).WithRetryInterval(1 * time.Nanosecond).UntilNoError(func() error {
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
		await := AtMost(1 * time.Millisecond).WithRetryInterval(1 * time.Nanosecond)
		err := await.UntilNoError(func() error {
			called++
			return fmt.Errorf("foo %d", called)
		})
		if assert.Error(t, err) {
			assert.IsType(t, &TimeoutError{}, err)
			assert.Equal(t, await, err.(*TimeoutError).Spec)
			assert.Equal(t, fmt.Errorf("foo %d", called), err.(*TimeoutError).LastError)
			assert.Greater(t, called, 0)
		}
	})
}

func TestTimeoutError_Error(t *testing.T) {
	t.Run("should return the timeout message", func(t *testing.T) {
		err := TimeoutError{Spec: AtMost(13 * time.Millisecond), LastError: fmt.Errorf("some error")}
		assert.EqualError(t, &err, "timed out after 13ms")
	})
}

func TestTimeoutError_Unwrap(t *testing.T) {
	t.Run("should unwrap the last error", func(t *testing.T) {
		timeoutError := &TimeoutError{Spec: AtMost(13 * time.Millisecond), LastError: fmt.Errorf("some error")}
		assert.EqualError(t, timeoutError.Unwrap(), "some error")
	})
}
