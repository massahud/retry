![Go](https://github.com/massahud/retry/workflows/Go/badge.svg)

Retry
=======

Package retry is a simple module for retrying a function on a defined interval
until the function succeeds or timeouts. There is support for retrying a
group of functions at different concurrency levels.

Example with worker function that returns an error:
```go
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()
	worker := func(ctx context.Context) (interface{}, error) {
		return nil, errors.New("worker error")
	}
	if result != retry.Func(ctx, 500*time.Microsecond, worker); result.Err != nil {
		return result.Err
	}
```

Example with multiple worker functions running in parallel:
```go
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()
	worker1 := func(ctx context.Context) (interface{}, error) {
		return "worker1 return", nil
	}
	worker2 := func(ctx context.Context) (interface{}, error) {
		return nil, fmt.Error("worker2 error")
	}
	workers := map[string]retry.Worker{"worker2": worker1, "worker2": worker2}
	results := retry.All(ctx, 200*time.Microsecond, workers, 0)
	for name, result := range results {
		fmt.Println("Name:", name, "result:", result)
	}
```

Example with multiple worker functions running in parallel waiting for first to return:
```go
	faster := func(ctx context.Context) (interface{}, error) {
		time.Sleep(time.Microsecond)
		return "I'm fast", nil
	}
	slower := func(ctx context.Context) (interface{}, error) {
		time.Sleep(time.Millisecond)
		return "I'm slow", nil
	}
	workers := map[string]retry.Worker{"faster": faster, "slower": slower}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	result := retry.First(ctx, time.Millisecond, workers, 0)
	if result.Err != nil {
		log.Fatal(result.Err)
	}
	fmt.Prinln("first result:", result.Value)
```

[GoDoc](https://pkg.go.dev/github.com/massahud/retry?tab=doc)
