//
// DISCLAIMER
//
// Copyright 2025 ArangoDB GmbH, Cologne, Germany
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

package util

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/pkg/errors"
)

type TimeoutFunc[T any] func() (T, error)

type TimeoutContFunc[T any] func(in T) error

func (t TimeoutFunc[T]) Run(ctx context.Context, timeout, interval time.Duration) (T, error) {
	timeoutT := time.NewTimer(timeout)
	defer timeoutT.Stop()

	intervalT := time.NewTicker(interval)
	defer intervalT.Stop()

	for {
		res, err := t()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return res, nil
			}

			return Default[T](), err
		}

		select {
		case <-timeoutT.C:
			return Default[T](), os.ErrDeadlineExceeded
		case <-ctx.Done():
			return Default[T](), os.ErrDeadlineExceeded
		case <-intervalT.C:
			continue
		}
	}
}

func (t TimeoutFunc[T]) With(f TimeoutContFunc[T]) TimeoutFunc[T] {
	return func() (T, error) {
		o, err := t()
		if err != nil {
			if errors.Is(err, io.EOF) {
				if err := f(o); err != nil {
					if errors.Is(err, io.EOF) {
						return o, io.EOF
					}

					return Default[T](), err
				}
			}

			return Default[T](), err
		}

		return Default[T](), nil
	}
}

func NewTimeoutFunc[T any](in TimeoutFunc[T]) TimeoutFunc[T] {
	return in
}
