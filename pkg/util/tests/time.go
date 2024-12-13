//
// DISCLAIMER
//
// Copyright 2016-2024 ArangoDB GmbH, Cologne, Germany
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

package tests

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

func DurationBetween() func(t *testing.T, expected time.Duration, skew float64) {
	start := time.Now()
	return func(t *testing.T, expected time.Duration, skew float64) {
		current := time.Since(start)
		min := time.Duration(float64(expected) * (1 - skew))
		max := time.Duration(float64(expected) * (1 + skew))

		if current > max || current < min {
			require.Failf(t, "Skew is too big", "Expected %d, got %d", expected, current)
		}
	}
}

func Interrupt() error {
	return interrupt{}
}

type interrupt struct {
}

func (i interrupt) Error() string {
	return "interrupt"
}

func NewTimeout(in Timeout) Timeout {
	return in
}

type Timeout func() error

func (t Timeout) WithTimeout(timeout, interval time.Duration) error {
	timeoutT := time.NewTimer(timeout)
	defer timeoutT.Stop()

	intervalT := time.NewTicker(interval)
	defer intervalT.Stop()

	for {
		select {
		case <-timeoutT.C:
			return errors.Errorf("Timeouted!")
		case <-intervalT.C:
			if err := t(); err != nil {
				var interrupt interrupt
				if errors.As(err, &interrupt) {
					return nil
				}

				return err
			}
		}
	}
}

func (t Timeout) WithContextTimeout(ctx context.Context, timeout, interval time.Duration) error {
	timeoutT := time.NewTimer(timeout)
	defer timeoutT.Stop()

	intervalT := time.NewTicker(interval)
	defer intervalT.Stop()

	for {
		select {
		case <-ctx.Done():
			return errors.Errorf("ContextCancelled!")
		case <-timeoutT.C:
			return errors.Errorf("Timeouted!")
		case <-intervalT.C:
			if err := t(); err != nil {
				var interrupt interrupt
				if errors.As(err, &interrupt) {
					return nil
				}

				return err
			}
		}
	}
}

func (t Timeout) WithTimeoutT(z *testing.T, timeout, interval time.Duration) {
	require.NoError(z, t.WithTimeout(timeout, interval))
}

func (t Timeout) WithContextTimeoutT(z *testing.T, ctx context.Context, timeout, interval time.Duration) {
	require.NoError(z, t.WithContextTimeout(ctx, timeout, interval))
}
