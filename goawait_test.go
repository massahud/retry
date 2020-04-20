package goawait_test

import (
	"context"
	"errors"
	"fmt"
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
		poll := func(ctx context.Context) goawait.Result {
			if calls >= 3 {
				return goawait.Result{}
			}
			calls++
			return goawait.Result{Err: errors.New("foo")}
		}
		result := goawait.Poll(context.Background(), retryInterval, poll)
		assert.NoError(t, result.Err)
		assert.Equal(t, 3, calls)
	})

	t.Run("cancel", func(t *testing.T) {
		t.Log("Poll should return error because the cancel function is called")
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		pollError := fmt.Errorf("foo")
		poll := func(ctx context.Context) goawait.Result {
			cancel()
			return goawait.Result{Err: pollError}
		}
		result := goawait.Poll(ctx, time.Second, poll)
		if assert.Error(t, result.Err) {
			assert.IsType(t, &goawait.Error{}, result.Err)
			assert.Equal(t, pollError, errors.Unwrap(result.Err))
		}
	})

	t.Run("timeout", func(t *testing.T) {
		t.Log("Poll should return error because the timeout exceeded")
		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
		defer cancel()
		pollError := fmt.Errorf("foo")
		poll := func(ctx context.Context) goawait.Result {
			return goawait.Result{Err: pollError}
		}
		result := goawait.Poll(ctx, time.Second, poll)
		if assert.Error(t, result.Err) {
			assert.IsType(t, &goawait.Error{}, result.Err)
			assert.Equal(t, pollError, errors.Unwrap(result.Err))
		}
	})
}

func TestPollAll(t *testing.T) {
	t.Run("noerror", func(t *testing.T) {
		t.Log("PollAll should return because all poll functions complete successfully")
		retryInterval := time.Nanosecond
		var calls int
		poll := func(ctx context.Context) goawait.Result {
			if calls >= 3 {
				return goawait.Result{}
			}
			calls++
			return goawait.Result{Err: errors.New("foo")}
		}
		polls := map[string]goawait.PollFunc{"poll1": poll, "poll2": poll}
		results := goawait.PollAll(context.Background(), retryInterval, polls)
		for _, result := range results {
			assert.NoError(t, result.Err)
		}
	})

	t.Run("cancel", func(t *testing.T) {
		t.Log("PollAll should return errors because the cancel function was called")
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		pollError := fmt.Errorf("foo")
		poll := func(ctx context.Context) goawait.Result {
			cancel()
			return goawait.Result{Err: pollError}
		}
		polls := map[string]goawait.PollFunc{"poll1": poll, "poll2": poll}
		results := goawait.PollAll(ctx, time.Second, polls)
		assert.Len(t, results, 2)
		for _, result := range results {
			if assert.Error(t, result.Err) {
				assert.IsType(t, &goawait.Error{}, result.Err)
				var err *goawait.Error
				assert.True(t, errors.As(result.Err, &err))
				assert.Equal(t, pollError, errors.Unwrap(result.Err))
			}
		}
	})

	t.Run("timeout", func(t *testing.T) {
		t.Log("PollAll should return error because the timeout exceeded and not all functions completed")
		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
		defer cancel()
		pollError := fmt.Errorf("foo")
		poll1 := func(ctx context.Context) goawait.Result {
			return goawait.Result{}
		}
		poll2 := func(ctx context.Context) goawait.Result {
			return goawait.Result{Err: pollError}
		}
		polls := map[string]goawait.PollFunc{"poll1": poll1, "poll2": poll2}
		results := goawait.PollAll(ctx, time.Second, polls)
		assert.Len(t, results, 2)
		assert.Nil(t, results["poll1"].Err)
		if assert.Error(t, results["poll2"].Err) {
			assert.IsType(t, &goawait.Error{}, results["poll2"].Err)
			var err *goawait.Error
			assert.True(t, errors.As(results["poll2"].Err, &err))
			assert.Equal(t, pollError, errors.Unwrap(results["poll2"].Err))
		}
	})
}

func ExamplePoll() {
	poll := func(ctx context.Context) goawait.Result {
		return goawait.Result{Err: errors.New("poll fail")}
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()
	result := goawait.Poll(ctx, 2*time.Millisecond, poll)
	fmt.Println(result)
}

func ExamplePollAll() {
	poll1 := func(ctx context.Context) goawait.Result {
		return goawait.Result{Err: errors.New("poll1 fail")}
	}
	poll2 := func(ctx context.Context) goawait.Result {
		return goawait.Result{Err: errors.New("poll1 fail")}
	}
	polls := map[string]goawait.PollFunc{"poll1": poll1, "poll2": poll2}
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()
	results := goawait.PollAll(ctx, 2*time.Millisecond, polls)
	fmt.Println(results)
}
