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

func TestPollAll(t *testing.T) {
	t.Run("noerror", func(t *testing.T) {
		t.Log("PollAll should return because all poll functions complete successfully")
		retryInterval := time.Nanosecond
		var calls int
		poll := func(ctx context.Context) error {
			if calls >= 3 {
				return nil
			}
			calls++
			return errors.New("foo")
		}
		polls := map[string]goawait.PollFunc{"poll1": poll, "poll2": poll}
		err := goawait.PollAll(context.Background(), retryInterval, polls)
		assert.NoError(t, err)
	})

	t.Run("cancel", func(t *testing.T) {
		t.Log("PollAll should return errors because the cancel function was called")
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		pollError := fmt.Errorf("foo")
		poll := func(ctx context.Context) error {
			cancel()
			return pollError
		}
		polls := map[string]goawait.PollFunc{"poll1": poll, "poll2": poll}
		err := goawait.PollAll(ctx, time.Second, polls)
		if assert.Error(t, err) {
			assert.IsType(t, goawait.Errors{}, err)
			var errs goawait.Errors
			assert.True(t, errors.As(err, &errs))
			assert.Len(t, errs, 2)
			assert.Equal(t, pollError, errors.Unwrap(errs["poll1"]))
			assert.Equal(t, pollError, errors.Unwrap(errs["poll2"]))
		}
	})

	t.Run("timeout", func(t *testing.T) {
		t.Log("PollAll should return error because the timeout exceeded and not all functions completed")
		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
		defer cancel()
		pollError := fmt.Errorf("foo")
		poll1 := func(ctx context.Context) error {
			return nil
		}
		poll2 := func(ctx context.Context) error {
			return pollError
		}
		polls := map[string]goawait.PollFunc{"poll1": poll1, "poll2": poll2}
		err := goawait.PollAll(ctx, time.Second, polls)
		if assert.Error(t, err) {
			assert.IsType(t, goawait.Errors{}, err)
			var errs goawait.Errors
			assert.True(t, errors.As(err, &errs))
			assert.Len(t, errs, 1)
			assert.Nil(t, errs["poll1"])
			assert.Equal(t, pollError, errors.Unwrap(errs["poll2"]))
		}
	})
}

func ExamplePoll() {
	poll := func(ctx context.Context) error {
		return errors.New("poll fail")
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()
	err := goawait.Poll(ctx, 2*time.Millisecond, poll)
	fmt.Println(err)
}

func ExamplePollAll() {
	poll1 := func(ctx context.Context) error {
		return errors.New("poll1 fail")
	}
	poll2 := func(ctx context.Context) error {
		return errors.New("poll1 fail")
	}
	polls := map[string]goawait.PollFunc{"poll1": poll1, "poll2": poll2}
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()
	err := goawait.PollAll(ctx, 2*time.Millisecond, polls)
	fmt.Println(err)
}
