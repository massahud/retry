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

func TestPoll(t *testing.T) {
	t.Run("noerror", func(t *testing.T) {
		t.Log("Poll should return because the poll function completes successfully")
		retryInterval := time.Nanosecond
		var calls int
		poll := func(ctx context.Context) error {
			if calls >= 3 {
				return nil
			}
			calls++
			return errors.New("foo")
		}
		err := goawait.Poll(context.Background(), retryInterval, poll)
		assert.NoError(t, err)
		assert.Equal(t, 3, calls)
	})

	t.Run("cancel", func(t *testing.T) {
		t.Log("Poll should return error because the cancel function is called")
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		pollError := fmt.Errorf("foo")
		poll := func(ctx context.Context) error {
			cancel()
			return pollError
		}
		err := goawait.Poll(ctx, time.Second, poll)
		if assert.Error(t, err) {
			assert.IsType(t, &goawait.Error{}, err)
			assert.Equal(t, pollError, errors.Unwrap(err))
		}
	})

	t.Run("timeout", func(t *testing.T) {
		t.Log("Poll should return error because the timeout exceeded")
		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
		defer cancel()
		pollError := fmt.Errorf("foo")
		poll := func(ctx context.Context) error {
			return pollError
		}
		err := goawait.Poll(ctx, time.Second, poll)
		if assert.Error(t, err) {
			assert.IsType(t, &goawait.Error{}, err)
			assert.Equal(t, pollError, errors.Unwrap(err))
		}
	})
}

func TestPollBool(t *testing.T) {
	t.Run("noerror", func(t *testing.T) {
		t.Log("Poll should return because the poll function completes successfully")
		retryInterval := time.Nanosecond
		var calls int
		poll := func(ctx context.Context) bool {
			if calls >= 3 {
				return true
			}
			calls++
			return false
		}
		err := goawait.PollBool(context.Background(), retryInterval, poll)
		assert.NoError(t, err)
		assert.Equal(t, 3, calls)
	})

	t.Run("cancel", func(t *testing.T) {
		t.Log("Poll should return error because the cancel function is called")
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		poll := func(ctx context.Context) bool {
			cancel()
			return false
		}
		err := goawait.PollBool(ctx, time.Second, poll)
		if assert.Error(t, err) {
			assert.IsType(t, &goawait.Error{}, err)
		}
	})

	t.Run("timeout", func(t *testing.T) {
		t.Log("Poll should return error because the timeout exceeded")
		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
		defer cancel()
		poll := func(ctx context.Context) bool {
			return false
		}
		err := goawait.PollBool(ctx, time.Second, poll)
		if assert.Error(t, err) {
			assert.IsType(t, &goawait.Error{}, err)
		}
	})
}

func ExamplePoll() {
	poll := func(ctx context.Context) error {
		return nil
	}
	err := goawait.Poll(context.Background(), 10*time.Millisecond, poll)
	if err != nil {
		log.Fatalf("database not ready: %s", err)
	}
	fmt.Println("Database ready")

	// Output:
	// Database ready
}

func ExamplePollBool() {
	poll := func(ctx context.Context) bool {
		return true
	}
	err := goawait.PollBool(context.Background(), 10*time.Millisecond, poll)
	if err != nil {
		log.Fatalf("page does not have item")
	}
	fmt.Println("page has item, continuing...")

	// Output:
	// page has item, continuing...
}
