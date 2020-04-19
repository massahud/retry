/*
Package goawait is a simple module for asynchronous waiting. Use goawait when
you need to wait for asynchronous tasks to complete before continuing normal
execution.

GoAwait has two polling functions that take a polling function and execute
that function until it succeeds or the specified timeout is exceeded.

Example with Polling function that returns an error:

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()
	poll := func(ctx context.Context) error {
			return errors.New("error message")
	}
	if err != goawait.Poll(ctx, 500 * time.Millisecond, poll); err != nil {
		return err
	}

Example with Polling function that returns a boolean:

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()
	poll := func(ctx context.Context) bool {
			return false
	}
	if err != goawait.PollBool(ctx, 500 * time.Millisecond, poll); err != nil {
		return err
	}
*/
package goawait
