GoAwait
=======

GoAwait is a simple module for asynchronous waiting.

Use goawait when you need to wait for asynchronous tasks to complete before continuing normal execution. 
It is very useful for waiting on integration and end to end tests.

## Documentation

[GoDoc](https://pkg.go.dev/github.com/massahud/goawait?tab=doc)

### Polling

GoAwait has polling functions that polls for something until it happens, or until the context is canceled.

```go
goawait.UntilNoError(cancelCtx, 500 * time.Millisecond, connectToDatabase)

goawait.Untiltrue(cancelCtx, 500 * time.Millisecond, messageReceived)
```

The polling functions are based on [Bill Kennedy's **retryTimeout** concurrency example](https://github.com/ardanlabs/gotraining/blob/master/topics/go/concurrency/channels/example1/example1.go)

### DSL

GoAwait also has a DSL, based on [Awaitility](https://github.com/awaitility/awaitility), to use it 
just start with AtMost or WithContext functions:

```go
goawait.AtMost(10 * time.Second).
    UntilTrue(receivedMessage)

goawait.WithContext(cancelContext).
    UntilNoError(connectToServer)

goawait.WithContext(cancelContext).
    AtMost(1 * time.Second).
    RetryingEvery(10 * time.Millisecond).
    UntilNoError(connectToServer)
```
DSL constructors have a default retry time of 100ms
 