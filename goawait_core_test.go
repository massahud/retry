package goawait_test

import (
	"context"
	"github.com/massahud/goawait"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestUntil(t *testing.T) {
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
		err := goawait.Until(ctx, nanoRetry, trueOnThirdCall)
		assert.NoError(t, err)
		assert.EqualValues(t, 3, calls)
	})
}
