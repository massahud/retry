/*
Package await is a simple module for asynchronous waiting. Use await when
you need to wait for asynchronous tasks to complete before continuing normal
execution.

Example with worker function that returns an error:
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()
	worker := func(ctx context.Context) (interface{}, error) {
		return nil, errors.New("worker error")
	}
	if result != await.Func(ctx, 500*time.Microsecond, worker); result.Err != nil {
		return result.Err
	}

Example with multiple worker functions running in parallel:
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()
	worker1 := func(ctx context.Context) (interface{}, error) {
		return "worker1 return", nil
	}
	worker2 := func(ctx context.Context) (interface{}, error) {
		return nil, fmt.Error("worker2 error")
	}
	workers := map[string]goawait.Worker{"worker2": worker1, "worker2": worker2}
	results := await.All(ctx, 200*time.Microsecond, workers)
	for name, result := range results {
		fmt.Println("Name:", name, "result:", result)
	}

Example with multiple worker functions running in parallel waiting for first to return:
	faster := func(ctx context.Context) (interface{}, error) {
		time.Sleep(time.Microsecond)
		return "I'm fast", nil
	}
	slower := func(ctx context.Context) (interface{}, error) {
		time.Sleep(time.Millisecond)
		return "I'm slow", nil
	}
	workers := map[string]goawait.Worker{"faster": faster, "slower": slower}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	result := await.First(ctx, time.Millisecond, workers)
	if result.Err != nil {
		log.Fatal(result.Err)
	}
	fmt.Prinln("first result:", result.Value)
*/
package await
