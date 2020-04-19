GoAwait
=======

GoAwait is a simple module for asynchronous waiting.

Use goawait when you need to wait for asynchronous tasks to complete before continuing normal 
execution. It is very useful for waiting on integration and end to end tests.  

To use it just create a spec and call one of it's Until methods

```go
  func receivedMessage() bool { ... }

  goawait.AtMost(10 * time.Second).
      RetryingEvery(200 * time.Millisecond).
      Until(receivedMessage)
```

GoAwait is based on [Java's Awaitility](https://github.com/awaitility/awaitility)'s DSL.

The polling functions were based on [Bill Kennedy's **retryTimeout** concurrency example](https://github.com/ardanlabs/gotraining/blob/0728ec842fbde65115e1a0a255b62b4a93d4c6a8/topics/go/concurrency/channels/example1/example1.go#L290)
