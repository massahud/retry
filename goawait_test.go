package goawait_test

import (
	"context"
	"errors"
	"fmt"
	"github.com/massahud/goawait"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
	"time"
)

func TestUntilNoError(t *testing.T) {
	t.Run("should return when poll returns no error", func(t *testing.T) {
		nanoRetry := 1 * time.Nanosecond
		var calls int
		noErrorOnThirdCall := func(ctx context.Context) error {
			if calls >= 3 {
				return nil
			}
			calls++
			return errors.New("foo")
		}
		ctx := context.Background()
		err := goawait.UntilNoError(ctx, nanoRetry, noErrorOnThirdCall)
		assert.NoError(t, err)
		assert.Equal(t, 3, calls)
	})

	t.Run("should return TimeoutError when context is cancelled before poll is true", func(t *testing.T) {
		context.Background().Err()
		ctx, cancel := context.WithCancel(context.Background())
		pollError := fmt.Errorf("foo")
		err := goawait.UntilNoError(ctx, 1*time.Nanosecond,
			func(ctx context.Context) error {
				cancel()
				return pollError
			})
		if assert.Error(t, err) {
			assert.IsType(t, &goawait.TimeoutError{}, err)
			assert.Equal(t, ctx.Err(), errors.Unwrap(err))
			assert.Equal(t, pollError, err.(*goawait.TimeoutError).LastError())
		}
	})

	t.Run("should pass ctx to poll function", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "foo", "bar")
		var calls int

		err := goawait.UntilNoError(ctx, 10*time.Millisecond, func(receivedCtx context.Context) error {
			calls++
			assert.Equal(t, ctx, receivedCtx)
			return nil
		})

		if assert.NoError(t, err) {
			assert.Equal(t, 1, calls)
		}
	})

	t.Run("should not call poll function if ctx is already cancelled", func(t *testing.T) {
		context.Background().Err()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		var calls int
		err := goawait.UntilNoError(ctx, 1*time.Nanosecond,
			func(ctx context.Context) error {
				calls++
				return fmt.Errorf("foo")
			})
		if assert.Error(t, err) {
			assert.Zero(t, calls)
			assert.Nil(t, err.(*goawait.TimeoutError).LastError())
		}
	})
}

func TestUntilTrue(t *testing.T) {
	t.Run("should return when poll is true", func(t *testing.T) {
		var calls int
		trueOnThirdCall := func(ctx context.Context) bool {
			if calls >= 3 {
				return true
			}
			calls++
			return calls >= 3
		}
		err := goawait.UntilTrue(context.Background(), 1*time.Nanosecond, trueOnThirdCall)
		assert.NoError(t, err)
		assert.Equal(t, 3, calls)
	})

	t.Run("should return TimeoutError when context is cancelled before poll is true", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		err := goawait.UntilTrue(ctx, 1*time.Nanosecond,
			func(ctx context.Context) bool {
				cancel()
				return false
			})
		if assert.Error(t, err) {
			assert.IsType(t, &goawait.TimeoutError{}, err)
			assert.Equal(t, ctx.Err(), errors.Unwrap(err))
			assert.Nil(t, err.(*goawait.TimeoutError).LastError())
		}
	})

	t.Run("should pass ctx to poll function", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "foo", "bar")
		var calls int

		err := goawait.UntilTrue(ctx, 10*time.Millisecond,
			func(receivedCtx context.Context) bool {
				calls++
				assert.Equal(t, ctx, receivedCtx)
				return true
			})

		if assert.NoError(t, err) {
			assert.Equal(t, 1, calls)
		}
	})

	t.Run("should not call poll function if ctx is already cancelled", func(t *testing.T) {
		context.Background().Err()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		var calls int
		err := goawait.UntilTrue(ctx, 1*time.Nanosecond,
			func(ctx context.Context) bool {
				calls++
				return false
			})
		if assert.Error(t, err) {
			assert.Zero(t, calls)
			assert.Nil(t, err.(*goawait.TimeoutError).LastError())
		}
	})
}

// helper function for examples
func connectToDatabase(ctx context.Context) error {
	return nil
}

// helper function for examples
func pageHasItem(ctx context.Context) bool {
	return true
}

func ExampleUntilNoError() {

	// func connectToDatabase(ctx context.Context) error { ... }

	err := goawait.UntilNoError(context.Background(), 10*time.Millisecond, connectToDatabase)

	if err != nil {
		log.Fatalf("database not ready: %s", err.Error())
	}

	fmt.Println("Database ready")

	// Output: Database ready
}

func ExampleUntilTrue() {

	// func pageHasItem(ctx context.Context) bool { ... }

	err := goawait.UntilTrue(context.Background(), 10*time.Millisecond, pageHasItem)

	if err != nil {
		log.Fatalf("page does not have item")
	}

	fmt.Println( "page has item, continuing...")

	// Output: page has item, continuing...
}
