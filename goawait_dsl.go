package goawait

import (
	"context"
	"time"
)

// DefaultAwait can be used instead of creating your own Await value.
var DefaultAwait = Await{
	maxWait:       0,
	retryInterval: 100 * time.Millisecond,
}

// Await is GoAwait's DSL type, create one with goawait.AtMost or
// goawait.WithContext functions
type Await struct {
	ctx           context.Context
	maxWait       time.Duration
	retryInterval time.Duration
}

// NewAwait creates a new Await with a specified timeout.
func NewAwait(maxWait time.Duration, retryInterval time.Duration) Await {
	return Await{maxWait: maxWait, retryInterval: retryInterval}
}

// NewAwaitContext creates a new Await with a specified timeout provided
// by the context value.
func NewAwaitContext(ctx context.Context, retryInterval time.Duration) Await {
	return Await{ctx: ctx, retryInterval: retryInterval}
}

// UntilNoError calls the poll function every retry interval until the poll
// function succeeds or the max wait time is reached.
func (await Await) UntilNoError(poll func(ctx context.Context) error) error {
	ctx, cancel := createContext(await)
	if cancel != nil {
		defer cancel()
	}

	return UntilNoError(ctx, await.retryInterval, poll)
}

// UntilTrue calls the poll function every retry interval until the poll
// function succeeds or the context times out.
func (await Await) UntilTrue(poll func(ctx context.Context) bool) error {
	ctx, cancel := createContext(await)
	if cancel != nil {
		defer cancel()
	}

	return UntilTrue(ctx, await.retryInterval, poll)
}

func createContext(await Await) (context.Context, context.CancelFunc) {
	if await.ctx != nil {
		return await.ctx, nil
	}
	if await.maxWait <= 0 {
		return context.Background(), func() {}
	}
	return context.WithTimeout(context.Background(), await.maxWait)
}
