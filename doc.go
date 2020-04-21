/*
Package goawait is a simple module for asynchronous waiting. Use goawait when
you need to wait for asynchronous tasks to complete before continuing normal
execution.

GoAwait has functions that take a polling function and execute
that function until it succeeds or the specified timeout is exceeded.

Example with polling function that returns an error:
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()
	poll := func(ctx context.Context) (interface{}, error) {
			return nil, errors.New("error message")
	}
	if result != goawait.Poll(ctx, 500*time.Microsecond, poll); result.Err != nil {
		return result.Err
	}

Example simultaneously polling multiple functions:
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()
	poll1 := func(ctx context.Context) (interface{}, error) {
			return nil, nil
	}
	poll2 := func(ctx context.Context) (interface{}, error) {
			return nil, errors.New("error message")
	}
	polls := map[string]goawait.PollFunc{"poll1": poll1, "poll2": poll2}
	results := goawait.PollAll(ctx, 500*time.Microsecond, polls)
	for name, result := range results {
		fmt.Println("Name:", name, "result:", result)
	}
*/
package await
