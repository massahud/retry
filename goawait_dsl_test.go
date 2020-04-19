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
		await := goawait.NewAwait(time.Second, time.Nanosecond)
		poll := func(_ context.Context) error {
			calls++
			if calls == 3 {
				return nil
			}
			return errors.New("foo")
		}
		err := await.UntilNoError(poll)
		if assert.NoError(t, err) {
			assert.Equal(t, 3, calls)
		}
	})

	t.Run("should return a TimeoutError if max time is reached", func(t *testing.T) {
		var calls int
		await := goawait.NewAwait(time.Millisecond, time.Nanosecond)
		poll := func(_ context.Context) error {
			calls++
			return fmt.Errorf("foo %d", calls)
		}
		err := await.UntilNoError(poll)
		if assert.Error(t, err) {
			assert.IsType(t, &goawait.TimeoutError{}, err)
			assert.Equal(t, fmt.Errorf("foo %d", calls), err.(*goawait.TimeoutError).ErrPoll)
			assert.Greater(t, calls, 0)
		}
	})

	t.Run("should pass the await context to the poll function", func(t *testing.T) {
		var calls int
		ctx := context.WithValue(context.Background(), "foo", "bar")
		ctx, cancel := context.WithTimeout(ctx, time.Millisecond)
		defer cancel()
		await := goawait.NewAwaitContext(ctx, time.Nanosecond)
		poll := func(receivedCtx context.Context) error {
			calls++
			assert.Equal(t, ctx, receivedCtx)
			return nil
		}
		err := await.UntilNoError(poll)
		if assert.NoError(t, err) {
			assert.EqualValues(t, 1, calls)
		}
	})

	t.Run("should return TimeoutError when context is cancelled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		await := goawait.NewAwaitContext(ctx, time.Nanosecond)
		poll := func(ctx context.Context) error {
			cancel()
			return errors.New("foo")
		}
		err := await.UntilNoError(poll)
		if assert.Error(t, err) {
			assert.IsType(t, &goawait.TimeoutError{}, err)
			assert.Equal(t, "context canceled", errors.Unwrap(err).Error())
		}
	})

	t.Run("should not call function when context is cancelled before running", func(t *testing.T) {
		var calls int
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		await := goawait.NewAwaitContext(ctx, time.Nanosecond)
		poll := func(ctx context.Context) error {
			calls++
			return errors.New("foo")
		}
		err := await.UntilNoError(poll)
		if assert.Error(t, err) {
			assert.Zero(t, calls)
		}
	})
}

func TestAwait_UntilTrue(t *testing.T) {
	t.Run("should retry until the poll function returns true", func(t *testing.T) {
		var calls int
		await := goawait.NewAwait(time.Second, time.Nanosecond)
		poll := func(ctx context.Context) bool {
			calls++
			if calls == 3 {
				return true
			}
			return false
		}
		err := await.UntilTrue(poll)
		if assert.NoError(t, err) {
			assert.Equal(t, 3, calls)
		}
	})

	t.Run("should return a TimeoutError if max time is reached", func(t *testing.T) {
		var calls int
		await := goawait.NewAwait(time.Millisecond, time.Nanosecond)
		poll := func(context.Context) bool {
			calls++
			return false
		}
		err := await.UntilTrue(poll)
		if assert.Error(t, err) {
			assert.IsType(t, &goawait.TimeoutError{}, err)
			assert.Equal(t, errors.New("poll failed"), err.(*goawait.TimeoutError).ErrPoll)
			assert.Greater(t, calls, 0)
		}
	})

	t.Run("should pass the await context to the poll function", func(t *testing.T) {
		var calls int
		ctx := context.WithValue(context.Background(), "foo", "bar")
		ctx, cancel := context.WithTimeout(ctx, time.Millisecond)
		defer cancel()
		await := goawait.NewAwaitContext(ctx, time.Nanosecond)
		poll := func(ctx context.Context) bool {
			calls++
			assert.Equal(t, ctx, ctx)
			return true
		}
		err := await.UntilTrue(poll)
		if assert.NoError(t, err) {
			assert.EqualValues(t, 1, calls)
		}
	})

	t.Run("should return TimeoutError when context is cancelled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		await := goawait.NewAwaitContext(ctx, time.Nanosecond)
		poll := func(ctx context.Context) bool {
			cancel()
			return false
		}
		err := await.UntilTrue(poll)
		if assert.Error(t, err) {
			assert.IsType(t, &goawait.TimeoutError{}, err)
			assert.Equal(t, "context canceled", errors.Unwrap(err).Error())
		}
	})

	t.Run("should not call function when context is cancelled before running", func(t *testing.T) {
		var calls int
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		await := goawait.NewAwaitContext(ctx, time.Nanosecond)
		poll := func(ctx context.Context) bool {
			calls++
			return false
		}
		err := await.UntilTrue(poll)
		if assert.Error(t, err) {
			assert.Zero(t, calls)
		}
	})
}

func ExampleAwait_UntilNoError() {

	// func connectToDatabase(ctx context.Context) error { ... }

	await := goawait.NewAwait(time.Second, time.Millisecond)
	err := await.UntilNoError(connectToDatabase)
	if err != nil {
		log.Fatalf("database not ready: %s", err.Error())
	}
	fmt.Println("Database ready")

	// Output: Database ready
}

func ExampleAwait_UntilTrue() {

	// func pageHasItem(ctx context.Context) bool { ... }

	await := goawait.NewAwait(time.Second, time.Millisecond)
	err := await.UntilTrue(pageHasItem)
	if err != nil {
		log.Fatalf("page does not have item")
	}
	fmt.Println("page has item, continuing...")

	// Output: page has item, continuing...
}
