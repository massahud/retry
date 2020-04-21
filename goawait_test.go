package await_test

import (
	"context"
	"errors"
	"fmt"
	"log"
	"testing"
	"time"

	await "github.com/massahud/goawait"
	"github.com/stretchr/testify/assert"
)

func TestPoll(t *testing.T) {
	t.Run("noerror", func(t *testing.T) {
		t.Log("Poll should return because the poll function completes successfully")
		retryInterval := time.Nanosecond
		poll := func(ctx context.Context) (interface{}, error) {
			time.Sleep(time.Millisecond)
			return nil, nil
		}
		result := await.Func(context.Background(), retryInterval, poll)
		assert.NoError(t, result.Err)
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
		result := await.Func(ctx, time.Second, poll)
		if assert.Error(t, result.Err) {
			assert.IsType(t, &await.Error{}, result.Err)
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
		result := await.Func(ctx, time.Second, poll)
		if assert.Error(t, result.Err) {
			assert.IsType(t, &await.Error{}, result.Err)
			assert.Equal(t, pollError, errors.Unwrap(result.Err))
		}
	})
}

func TestPollAll(t *testing.T) {
	t.Run("noerror", func(t *testing.T) {
		t.Log("PollAll should return because all poll functions complete successfully")
		retryInterval := time.Nanosecond
		poll := func(ctx context.Context) (interface{}, error) {
			time.Sleep(time.Millisecond)
			return nil, nil
		}
		polls := map[string]await.PollFunc{"poll1": poll, "poll2": poll}
		results := await.All(context.Background(), retryInterval, polls)
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
		polls := map[string]await.PollFunc{"poll1": poll, "poll2": poll}
		results := await.All(ctx, time.Second, polls)
		assert.Len(t, results, 2)
		for _, result := range results {
			if assert.Error(t, result.Err) {
				assert.IsType(t, &await.Error{}, result.Err)
				var err *await.Error
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
		polls := map[string]await.PollFunc{"poll1": poll1, "poll2": poll2}
		results := await.All(ctx, time.Second, polls)
		assert.Len(t, results, 2)
		assert.NoError(t, results["poll1"].Err)
		if assert.Error(t, results["poll2"].Err) {
			assert.IsType(t, &await.Error{}, results["poll2"].Err)
			var err *await.Error
			assert.True(t, errors.As(results["poll2"].Err, &err))
			assert.Equal(t, pollError, errors.Unwrap(results["poll2"].Err))
		}
	})
}

func TestPollFirst(t *testing.T) {
	t.Run("noerror", func(t *testing.T) {
		t.Log("PollFirst should return the result we chose from three functions.")
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
		polls := map[string]await.PollFunc{"poll5": poll5, "poll8": poll8, "poll12": poll12}
		result := await.First(context.Background(), time.Millisecond, polls)
		if assert.NoError(t, result.Err) {
			assert.Equal(t, result.Value.(string), "5 Milliseconds")
		}
	})

	t.Run("waitsuccess", func(t *testing.T) {
		t.Log("PollFirst should return the result from the first successful function.")
		poll5 := func(ctx context.Context) (interface{}, error) {
			time.Sleep(5 * time.Millisecond)
			return nil, errors.New("some error")
		}
		poll8 := func(ctx context.Context) (interface{}, error) {
			time.Sleep(8 * time.Millisecond)
			return "8 Milliseconds", nil
		}
		poll12 := func(ctx context.Context) (interface{}, error) {
			time.Sleep(12 * time.Millisecond)
			return "12 Milliseconds", nil
		}
		polls := map[string]await.PollFunc{"poll5": poll5, "poll8": poll8, "poll12": poll12}
		result := await.First(context.Background(), time.Millisecond, polls)
		if assert.NoError(t, result.Err) {
			assert.Equal(t, result.Value.(string), "8 Milliseconds")
		}
	})

	t.Run("cancel", func(t *testing.T) {
		t.Log("PollFirst should return the result from the first successful function.")
		ch := make(chan string, 2)
		poll5 := func(ctx context.Context) (interface{}, error) {
			time.Sleep(5 * time.Millisecond)
			return "5 Milliseconds", nil
		}
		poll8 := func(ctx context.Context) (interface{}, error) {
			<-ctx.Done()
			time.Sleep(8 * time.Millisecond)
			ch <- "8 Milliseconds cancelled"
			return "8 Milliseconds", nil
		}
		poll12 := func(ctx context.Context) (interface{}, error) {
			<-ctx.Done()
			time.Sleep(12 * time.Millisecond)
			ch <- "12 Milliseconds cancelled"
			return "12 Milliseconds", nil
		}
		polls := map[string]await.PollFunc{"poll5": poll5, "poll8": poll8, "poll12": poll12}
		result := await.First(context.Background(), time.Millisecond, polls)
		if assert.NoError(t, result.Err) {
			assert.Equal(t, result.Value.(string), "5 Milliseconds")
			assert.Equal(t, <-ch, "8 Milliseconds cancelled")
			assert.Equal(t, <-ch, "12 Milliseconds cancelled")
		}
	})

	t.Run("timeout", func(t *testing.T) {
		t.Log("PollFirst should return the result from the first successful function.")

		poll1 := func(ctx context.Context) (interface{}, error) {
			return nil, fmt.Errorf("error message")
		}
		poll2 := func(ctx context.Context) (interface{}, error) {
			return nil, fmt.Errorf("error message")
		}
		polls := map[string]await.PollFunc{"poll1": poll1, "poll2": poll2}
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Millisecond)
		defer cancel()
		result := await.First(ctx, time.Millisecond, polls)
		if assert.Error(t, result.Err) {
			assert.Regexp(t, "context cancelled after .+", result.Err.Error())
		}
	})
}

func ExampleFunc() {
	t := time.NewTimer(time.Millisecond)
	poll := func(ctx context.Context) (interface{}, error) {
		select {
		case <-t.C:
			return "timer finished", nil
		default:
			return nil, errors.New("poll fail")
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Millisecond)
	defer cancel()
	result := await.Func(ctx, 200*time.Microsecond, poll)
	fmt.Println("Result:", result.Value)

	// Output:
	// Result: timer finished
}

func ExampleAll() {
	poll1 := func(ctx context.Context) (interface{}, error) {
		return nil, errors.New("error message")
	}
	poll2 := func(ctx context.Context) (interface{}, error) {
		return "ok", nil
	}
	polls := map[string]await.PollFunc{"poll1": poll1, "poll2": poll2}
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()
	results := await.All(ctx, 2*time.Millisecond, polls)
	for name, result := range results {
		switch {
		case result.Err != nil:
			fmt.Println(name, "timed out")
		default:
			fmt.Println(name, "returned value:", result.Value)
		}
	}

	// Unordered output:
	// poll1 timed out
	// poll2 returned value: ok
}

func ExampleFirst() {
	retryInterval := time.Nanosecond
	faster := func(ctx context.Context) (interface{}, error) {
		time.Sleep(time.Microsecond)
		return "I'm fast", nil
	}
	slower := func(ctx context.Context) (interface{}, error) {
		time.Sleep(time.Millisecond)
		return "I'm slow", nil
	}
	polls := map[string]await.PollFunc{"faster": faster, "slower": slower}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Millisecond)
	defer cancel()
	result := await.First(ctx, retryInterval, polls)
	if result.Err != nil {
		log.Fatal(result.Err)
	}
	fmt.Println("First returned value:", result.Value)

	// Output:
	// First returned value: I'm fast
}
