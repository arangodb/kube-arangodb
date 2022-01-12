//
// DISCLAIMER
//
// Copyright 2016-2022 ArangoDB GmbH, Cologne, Germany
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
//
// Copyright holder is ArangoDB GmbH, Cologne, Germany
//

package globals

import (
	"context"
	"time"
)

type TimeoutRunFunc func(ctxChild context.Context) error

type Timeout interface {
	Set(duration time.Duration)
	Get() time.Duration

	WithTimeout(ctx context.Context) (context.Context, context.CancelFunc)

	Run(run TimeoutRunFunc) error
	RunWithTimeout(ctx context.Context, run TimeoutRunFunc) error
}

func NewTimeout(duration time.Duration) Timeout {
	return &timeout{
		duration: duration,
	}
}

type timeout struct {
	duration time.Duration
}

func (t *timeout) Set(duration time.Duration) {
	t.duration = duration
}

func (t *timeout) Get() time.Duration {
	return t.duration
}

func (t *timeout) WithTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	if t.duration == 0 {
		return ctx, func() {}
	}

	return context.WithTimeout(ctx, t.duration)
}

func (t *timeout) Run(run TimeoutRunFunc) error {
	return t.RunWithTimeout(context.Background(), run)
}

func (t *timeout) RunWithTimeout(ctx context.Context, run TimeoutRunFunc) error {
	newCtx, c := t.WithTimeout(ctx)
	defer c()
	return run(newCtx)
}
