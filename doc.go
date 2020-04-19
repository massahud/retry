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

// Package goawait is a simple module for asynchronous waiting.
//
// Use goawait when you need to wait for asynchronous tasks to complete before continuing normal
// execution. It is very useful for waiting on integration and end to end tests.
//
// Polling
//
// GoAwait has polling functions that polls for something until it happens, or until the context
// is canceled.
//
//     goawait.UntilNoError(cancelCtx, 500 * time.Millisecond, connectToDatabase)
//
//     goawait.Untiltrue(cancelCtx, 500 * time.Millisecond, messageReceived)
//
// The polling functions are based on Bill Kennedy's retryTimeout concurrency example(https://github.com/ardanlabs/gotraining/blob/master/topics/go/concurrency/channels/example1/example1.go)
//
// package goawait also has a DSL, based on Awaitility's (https://github.com/awaitility/awaitility),
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
