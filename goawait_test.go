package goawait_test

import (
	"context"
	"errors"
	"fmt"
	"github.com/massahud/goawait"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestUntilNoError(t *testing.T) {
	t.Run("should return when poll returns no error", func(t *testing.T) {
		nanoRetry := 1 * time.Nanosecond
		var calls int32
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
		assert.EqualValues(t, 3, calls)
	})

	t.Run("should return TimeoutError when context is cancelled before poll is true", func(t *testing.T) {
		context.Background().Err()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		pollError := fmt.Errorf("foo")
		err := goawait.UntilNoError(ctx, 1*time.Nanosecond, func(ctx context.Context) error { return pollError })
		if assert.Error(t, err) {
			assert.IsType(t, &goawait.TimeoutError{}, err)
			assert.Equal(t, ctx.Err(), errors.Unwrap(err))
			assert.Equal(t, pollError, err.(*goawait.TimeoutError).LastError())
		}
	})
}

func TestUntilTrue(t *testing.T) {
	t.Run("should return when poll is true", func(t *testing.T) {
		nanoRetry := 1 * time.Nanosecond
		var calls int32
		trueOnThirdCall := func(ctx context.Context) bool {
			if calls >= 3 {
				return true
			}
			calls++
			return calls >= 3
		}
		ctx := context.Background()
		err := goawait.UntilTrue(ctx, nanoRetry, trueOnThirdCall)
		assert.NoError(t, err)
		assert.EqualValues(t, 3, calls)
	})

	t.Run("should return TimeoutError when context is cancelled before poll is true", func(t *testing.T) {
		context.Background().Err()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		err := goawait.UntilTrue(ctx, 1*time.Nanosecond, func(ctx context.Context) bool { return false })
		if assert.Error(t, err) {
			assert.IsType(t, &goawait.TimeoutError{}, err)
			assert.Equal(t, ctx.Err(), errors.Unwrap(err))
			assert.Nil(t, err.(*goawait.TimeoutError).LastError())
		}
	})
}
