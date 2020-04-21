![Go](https://github.com/massahud/goawait/workflows/Go/badge.svg)

GoAwait
=======

Package goawait is a simple module for asynchronous waiting. Use goawait when
you need to wait for asynchronous tasks to complete before continuing normal
execution.

GoAwait has functions that take a polling function and execute
that function until it succeeds or the specified timeout is exceeded.

Example with polling function that returns an error:
```go
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()
	poll := func(ctx context.Context) (interface{}, error) {
		return nil, errors.New("error message")
	}
	if result != await.Func(ctx, 500*time.Microsecond, poll); result.Err != nil {
		return result.Err
	}
```

Example simultaneously polling multiple functions:
```go
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()
	poll1 := func(ctx context.Context) (interface{}, error) {
		return "ok", nil
	}
	poll2 := func(ctx context.Context) (interface{}, error) {
		return nil, fmt.Error("error message")
	}
	polls := map[string]goawait.PollFunc{"poll1": poll1, "poll2": poll2}
	results := await.All(ctx, 200*time.Microsecond, polls)
	for name, result := range results {
		fmt.Println("Name:", name, "result:", result)
	}
```

Example polling until the first function returns:
```go
	faster := func(ctx context.Context) (interface{}, error) {
		time.Sleep(time.Microsecond)
		return "I'm fast", nil
	}
	slower := func(ctx context.Context) (interface{}, error) {
		time.Sleep(time.Millisecond)
		return "I'm slow", nil
	}
	polls := map[string]goawait.PollFunc{"faster": faster, "slower": slower}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	result := await.First(ctx, time.Millisecond, polls)
	if result.Err != nil {
		log.Fatal(result.Err)
	}
	fmt.Prinln("First result:", result.Value)
```

[GoDoc](https://pkg.go.dev/github.com/massahud/goawait?tab=doc)
