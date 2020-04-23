package await_test

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync/atomic"
	"testing"
	"time"

	"github.com/massahud/await"
	"github.com/stretchr/testify/assert"
)

func TestFunc(t *testing.T) {
	t.Run("noerror", func(t *testing.T) {
		t.Log("Func should return because the worker function completes successfully.")
		retryInterval := time.Nanosecond
		worker := func(ctx context.Context) (interface{}, error) {
			time.Sleep(time.Millisecond)
			return nil, nil
		}
		result := await.Func(context.Background(), retryInterval, worker)
		assert.NoError(t, result.Err)
	})

	t.Run("cancel", func(t *testing.T) {
		t.Log("Func should return error because the context cancel function is called.")
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		err := fmt.Errorf("foo")
		var counter int32
		worker := func(ctx context.Context) (interface{}, error) {
			if atomic.AddInt32(&counter, 1) > 10 {
				cancel()
			}
			return nil, err
		}
		result := await.Func(ctx, time.Millisecond, worker)
		if assert.Error(t, result.Err) {
			assert.IsType(t, &await.Error{}, result.Err)
			assert.Equal(t, err, errors.Unwrap(result.Err))
		}
	})

	t.Run("failfast", func(t *testing.T) {
		t.Log("Func should return error if context is closed before execution.")
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		err := fmt.Errorf("foo")
		worker := func(ctx context.Context) (interface{}, error) {
			cancel()
			return nil, err
		}
		result := await.Func(ctx, time.Second, worker)
		if assert.Error(t, result.Err) {
			assert.IsType(t, &await.Error{}, result.Err)
			assert.Equal(t, nil, errors.Unwrap(result.Err))
		}
	})

	t.Run("timeout", func(t *testing.T) {
		t.Log("Func should return error because the context timeout exceeded.")
		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
		defer cancel()
		err := fmt.Errorf("foo")
		worker := func(ctx context.Context) (interface{}, error) {
			return nil, err
		}
		result := await.Func(ctx, time.Second, worker)
		if assert.Error(t, result.Err) {
			assert.IsType(t, &await.Error{}, result.Err)
			assert.Equal(t, err, errors.Unwrap(result.Err))
		}
	})
}

func TestAll(t *testing.T) {
	t.Run("noerror", func(t *testing.T) {
		t.Log("All should return because all worker functions complete successfully.")
		retryInterval := time.Nanosecond
		worker := func(ctx context.Context) (interface{}, error) {
			time.Sleep(time.Millisecond)
			return nil, nil
		}
		workers := map[string]await.Worker{"worker1": worker, "worker2": worker}
		results := await.All(context.Background(), retryInterval, workers, 0)
		for _, result := range results {
			assert.NoError(t, result.Err)
		}
	})

	t.Run("cancel", func(t *testing.T) {
		t.Log("All should return errors because the context cancel function was called.")
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		errWork := fmt.Errorf("foo")
		var counter int32
		worker := func(ctx context.Context) (interface{}, error) {
			if atomic.AddInt32(&counter, 1) > 10 {
				cancel()
			}
			return nil, errWork
		}
		workers := map[string]await.Worker{"worker1": worker, "worker2": worker}
		results := await.All(ctx, time.Millisecond, workers, 0)
		assert.Len(t, results, 2)
		for _, result := range results {
			if assert.Error(t, result.Err) {
				assert.IsType(t, &await.Error{}, result.Err)
				var err *await.Error
				assert.True(t, errors.As(result.Err, &err))
			}
		}
		assert.Equal(t, errWork, errors.Unwrap(results["worker1"].Err))
		assert.Equal(t, errWork, errors.Unwrap(results["worker2"].Err))
	})

	t.Run("timeout", func(t *testing.T) {
		t.Log("All should return error because the context timeout exceeded and not all worker functions completed.")
		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
		defer cancel()
		errWork := fmt.Errorf("foo")
		worker1 := func(ctx context.Context) (interface{}, error) {
			return nil, nil
		}
		worker2 := func(ctx context.Context) (interface{}, error) {
			return nil, errWork
		}
		workers := map[string]await.Worker{"worker1": worker1, "worker2": worker2}
		results := await.All(ctx, time.Millisecond, workers, 0)
		assert.Len(t, results, 2)
		assert.NoError(t, results["worker1"].Err)
		if assert.Error(t, results["worker2"].Err) {
			assert.IsType(t, &await.Error{}, results["worker2"].Err)
			var err *await.Error
			assert.True(t, errors.As(results["worker2"].Err, &err))
			assert.Equal(t, errWork, errors.Unwrap(results["worker2"].Err))
		}
	})
}

func TestFirst(t *testing.T) {
	t.Run("noerror", func(t *testing.T) {
		t.Log("First should return the result we chose from three worker functions.")
		worker5 := func(ctx context.Context) (interface{}, error) {
			time.Sleep(5 * time.Millisecond)
			return "5 Milliseconds", nil
		}
		worker8 := func(ctx context.Context) (interface{}, error) {
			time.Sleep(8 * time.Millisecond)
			return "8 Milliseconds", nil
		}
		worker12 := func(ctx context.Context) (interface{}, error) {
			time.Sleep(12 * time.Millisecond)
			return "12 Milliseconds", nil
		}
		workers := map[string]await.Worker{"worker5": worker5, "worker8": worker8, "worker12": worker12}
		result := await.First(context.Background(), time.Millisecond, workers, 0)
		if assert.NoError(t, result.Err) {
			assert.Equal(t, result.Value.(string), "5 Milliseconds")
		}
	})

	t.Run("waitsuccess", func(t *testing.T) {
		t.Log("First should return the result from the first successful worker function.")
		worker5 := func(ctx context.Context) (interface{}, error) {
			time.Sleep(5 * time.Millisecond)
			return nil, errors.New("some error")
		}
		worker8 := func(ctx context.Context) (interface{}, error) {
			time.Sleep(8 * time.Millisecond)
			return "8 Milliseconds", nil
		}
		worker12 := func(ctx context.Context) (interface{}, error) {
			time.Sleep(12 * time.Millisecond)
			return "12 Milliseconds", nil
		}
		workers := map[string]await.Worker{"worker5": worker5, "worker8": worker8, "worker12": worker12}
		result := await.First(context.Background(), time.Millisecond, workers, 0)
		if assert.NoError(t, result.Err) {
			assert.Equal(t, result.Value.(string), "8 Milliseconds")
		}
	})

	t.Run("waitnotallsuccess", func(t *testing.T) {
		t.Log("First should return the result from the first successful worker function.")
		ch := make(chan string, 2)
		worker5 := func(ctx context.Context) (interface{}, error) {
			time.Sleep(5 * time.Millisecond)
			return "5 Milliseconds", nil
		}
		worker8 := func(ctx context.Context) (interface{}, error) {
			<-ctx.Done()
			time.Sleep(8 * time.Millisecond)
			ch <- "8 Milliseconds cancelled"
			return "8 Milliseconds", nil
		}
		worker12 := func(ctx context.Context) (interface{}, error) {
			<-ctx.Done()
			time.Sleep(12 * time.Millisecond)
			ch <- "12 Milliseconds cancelled"
			return "12 Milliseconds", nil
		}
		workers := map[string]await.Worker{"worker5": worker5, "worker8": worker8, "worker12": worker12}
		result := await.First(context.Background(), time.Millisecond, workers, 0)
		if assert.NoError(t, result.Err) {
			assert.Equal(t, result.Value.(string), "5 Milliseconds")
			assert.Equal(t, <-ch, "8 Milliseconds cancelled")
			assert.Equal(t, <-ch, "12 Milliseconds cancelled")
		}
	})

	t.Run("timeout", func(t *testing.T) {
		t.Log("PollFirst should return the result from the first successful function.")
		worker1 := func(ctx context.Context) (interface{}, error) {
			return nil, fmt.Errorf("error message")
		}
		worker2 := func(ctx context.Context) (interface{}, error) {
			return nil, fmt.Errorf("error message")
		}
		workers := map[string]await.Worker{"worker1": worker1, "worker2": worker2}
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Millisecond)
		defer cancel()
		result := await.First(ctx, time.Millisecond, workers, 0)
		if assert.Error(t, result.Err) {
			assert.Regexp(t, "context cancelled after .+", result.Err.Error())
		}
	})
}

func TestAllWithPooling(t *testing.T) {
	t.Run("noerror", func(t *testing.T) {
		t.Log("All should return because all worker functions complete successfully.")
		retryInterval := time.Nanosecond
		worker := func(ctx context.Context) (interface{}, error) {
			time.Sleep(time.Millisecond)
			return nil, nil
		}
		workers := map[string]await.Worker{"worker1": worker, "worker2": worker}
		results := await.All(context.Background(), retryInterval, workers, 16)
		for _, result := range results {
			assert.NoError(t, result.Err)
		}
	})

	t.Run("cancel", func(t *testing.T) {
		t.Log("All should return errors because the context cancel function was called.")
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		errWork := fmt.Errorf("foo")
		var count int32
		worker := func(ctx context.Context) (interface{}, error) {
			if atomic.AddInt32(&count, 1) > 10 {
				cancel()
			}
			return nil, errWork
		}
		workers := map[string]await.Worker{"worker1": worker, "worker2": worker}
		results := await.All(ctx, time.Millisecond, workers, 16)
		assert.Len(t, results, 2)
		for _, result := range results {
			if assert.Error(t, result.Err) {
				assert.IsType(t, &await.Error{}, result.Err)
				var err *await.Error
				assert.True(t, errors.As(result.Err, &err))
			}
		}
		assert.Equal(t, errWork, errors.Unwrap(results["worker1"].Err))
		assert.Equal(t, errWork, errors.Unwrap(results["worker2"].Err))
	})

	t.Run("timeout", func(t *testing.T) {
		t.Log("All should return error because the context timeout exceeded and not all worker functions completed.")
		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
		defer cancel()
		errWork := fmt.Errorf("foo")
		worker1 := func(ctx context.Context) (interface{}, error) {
			return nil, nil
		}
		worker2 := func(ctx context.Context) (interface{}, error) {
			return nil, errWork
		}
		workers := map[string]await.Worker{"worker1": worker1, "worker2": worker2}
		results := await.All(ctx, time.Millisecond, workers, 16)
		assert.Len(t, results, 2)
		assert.NoError(t, results["worker1"].Err)
		if assert.Error(t, results["worker2"].Err) {
			assert.IsType(t, &await.Error{}, results["worker2"].Err)
			var err *await.Error
			assert.True(t, errors.As(results["worker2"].Err, &err))
			assert.Equal(t, errWork, errors.Unwrap(results["worker2"].Err))
		}
	})
}

func ExampleFunc() {
	t := time.NewTimer(time.Millisecond)
	worker := func(ctx context.Context) (interface{}, error) {
		select {
		case <-t.C:
			return "timer finished", nil
		default:
			return nil, errors.New("poll fail")
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Millisecond)
	defer cancel()
	result := await.Func(ctx, 200*time.Microsecond, worker)
	fmt.Println("Result:", result.Value)

	// Output:
	// Result: timer finished
}

func ExampleAll() {
	worker1 := func(ctx context.Context) (interface{}, error) {
		return nil, errors.New("error message")
	}
	worker2 := func(ctx context.Context) (interface{}, error) {
		return "ok", nil
	}
	workers := map[string]await.Worker{"worker1": worker1, "worker2": worker2}
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()
	results := await.All(ctx, 2*time.Millisecond, workers, 0)
	for name, result := range results {
		switch {
		case result.Err != nil:
			fmt.Println(name, "timed out")
		default:
			fmt.Println(name, "returned value:", result.Value)
		}
	}

	// Unordered output:
	// worker1 timed out
	// worker2 returned value: ok
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
	workers := map[string]await.Worker{"faster": faster, "slower": slower}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Millisecond)
	defer cancel()
	result := await.First(ctx, retryInterval, workers, 0)
	if result.Err != nil {
		log.Fatal(result.Err)
	}
	fmt.Println("First returned value:", result.Value)

	// Output:
	// First returned value: I'm fast
}
