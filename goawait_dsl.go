// Copyright 2020 Geraldo Augusto Massahud Rodrigues dos Santos
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// DSL
//
// GoAwait also has a DSL, based on Awaitility's (https://github.com/awaitility/awaitility),
// to use it just start with AtMost or WithContext functions:
//
//     goawait.AtMost(10 * time.Second).
//         UntilTrue(receivedMessage)
//
//     goawait.WithContext(cancelContext).
//         UntilNoError(connectToServer)
//
//     goawait.WithContext(cancelContext).
//         AtMost(1 * time.Second).
//         RetryingEvery(10 * time.Millisecond).
//         UntilNoError(connectToServer)
//
// DSL constructors have a default retry time of 100ms
package goawait

import (
	"context"
	"time"
)

const DefaultRetryTime = 100 * time.Millisecond

// Await is GoAwait's DSL type, create one with goawait.AtMost or goawait.WithContext functions
type Await struct {
	ctx       context.Context
	maxWait   time.Duration
	retryTime time.Duration
}

// AtMost creates a new Await with a specified timeout
func AtMost(maxWait time.Duration) Await {
	return Await{ctx: context.Background(), maxWait: maxWait, retryTime: defaultRetryTime}
}

// WithContext creates a new Await with a context that can be used for cancelation
func WithContext(ctx context.Context) Await {
	return Await{ctx: ctx, maxWait: -1, retryTime: defaultRetryTime}
}

// AtMost configures the maximum await time of the Await
func (await Await) AtMost(maxWait time.Duration) Await {
	await.maxWait = maxWait
	return await
}

// RetryingEvery configures the Await retryTime
func (await Await) RetryingEvery(retryTime time.Duration) Await {
	await.retryTime = retryTime
	return await
}

// UntilTrue executes the polling function until the poll function returns true, or a timeout occurs
//
// Returns a TimeoutError on timeout or when the context is done.
func (await Await) UntilTrue(poll func(ctx context.Context) bool) error {
	timeoutCtx, cancel := createTimeoutContext(await)
	defer cancel()
	// poll must receive the await context, not timeoutCtx
	wrappedPoll := func(_ context.Context) bool {
		return poll(await.ctx)
	}
	return UntilTrue(timeoutCtx, await.retryTime, wrappedPoll)
}

// UntilNoError executes the polling function until it does not return an error.
//
// Returns a TimeoutError on timeout or when the context is done.
func (await Await) UntilNoError(poll func(ctx context.Context) error) error {
	timeoutCtx, cancel := createTimeoutContext(await)
	defer cancel()
	// poll must receive the await context, not timeoutCtx
	wrappedPoll := func(_ context.Context) error {
		return poll(await.ctx)
	}
	return UntilNoError(timeoutCtx, await.retryTime, wrappedPoll)
}

func createTimeoutContext(await Await) (context.Context, context.CancelFunc) {
	if await.maxWait < 0 {
		return await.ctx, func() {}
	}
	return context.WithTimeout(context.Background(), await.maxWait)
}
