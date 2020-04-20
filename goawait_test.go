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
		poll := func(ctx context.Context) (interface{}, error) {
			if calls >= 3 {
				return nil, nil
			}
			calls++
			return nil, errors.New("foo")
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
		poll := func(ctx context.Context) (interface{}, error) {
			cancel()
			return nil, pollError
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
		poll := func(ctx context.Context) (interface{}, error) {
			return nil, pollError
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
		poll := func(ctx context.Context) (interface{}, error) {
			if calls >= 3 {
				return nil, nil
			}
			calls++
			return nil, errors.New("foo")
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
		poll := func(ctx context.Context) (interface{}, error) {
			cancel()
			return nil, pollError
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
		poll1 := func(ctx context.Context) (interface{}, error) {
			return nil, nil
		}
		poll2 := func(ctx context.Context) (interface{}, error) {
			return nil, pollError
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

func TestPollFirst(t *testing.T) {
	t.Run("noerror", func(t *testing.T) {
		t.Log("PollFirst should the result we chose from three functions.")
		poll5 := func(ctx context.Context) (interface{}, error) {
			time.Sleep(5 * time.Millisecond)
			return "5 Milliseconds", nil
		}
		poll8 := func(ctx context.Context) (interface{}, error) {
			time.Sleep(8 * time.Millisecond)
			return "8 Milliseconds", nil
		}
		poll12 := func(ctx context.Context) (interface{}, error) {
			time.Sleep(12 * time.Millisecond)
			return "12 Milliseconds", nil
		}
		polls := map[string]goawait.PollFunc{"poll5": poll5, "poll8": poll8, "poll12": poll12}
		result := goawait.PollFirst(context.Background(), time.Nanosecond, polls)
		assert.Nil(t, result.Err)
		assert.Equal(t, result.Value.(string), "5 Milliseconds")
	})
}

func ExamplePoll() {
	poll := func(ctx context.Context) (interface{}, error) {
		return nil, errors.New("poll fail")
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()
	result := goawait.Poll(ctx, 2*time.Millisecond, poll)
	fmt.Println(result)
}

func ExamplePollAll() {
	poll1 := func(ctx context.Context) (interface{}, error) {
		return nil, errors.New("poll1 fail")
	}
	poll2 := func(ctx context.Context) (interface{}, error) {
		return nil, errors.New("poll1 fail")
	}
	polls := map[string]goawait.PollFunc{"poll1": poll1, "poll2": poll2}
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()
	results := goawait.PollAll(ctx, 2*time.Millisecond, polls)
	fmt.Println(results)
}
