//
// DISCLAIMER
//
// Copyright 2024 ArangoDB GmbH, Cologne, Germany
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

package logging

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestLogEventSampler(t *testing.T) {
	const SamplePeriod = time.Millisecond * 500
	s := NewLogEventSampler(SamplePeriod)
	require.True(t, s.Sample(Debug, "new-msg"))

	events := []struct {
		l   Level
		msg string
	}{
		{Debug, "test"},
		{Info, "test"},
		{Debug, "test-debug"},
	}

	trigger := func(t *testing.T, invocation int) {
		for _, ev := range events {
			sampled := s.Sample(ev.l, ev.msg)
			if invocation == 0 {
				require.True(t, sampled, invocation)
			} else {
				require.False(t, sampled, invocation)
			}
			time.Sleep(time.Millisecond)
		}
	}

	t.Run("sequentially", func(t *testing.T) {
		for i := 0; i < 50; i++ {
			trigger(t, i)
		}

		time.Sleep(time.Second)

		for i := 0; i < 50; i++ {
			trigger(t, i)
		}

		time.Sleep(time.Second)
	})

	t.Run("parallel", func(t *testing.T) {
		var wg sync.WaitGroup
		// event index -> how many samples
		results := map[int]*atomic.Int32{
			0: {},
			1: {},
			2: {},
		}
		expectedSamples := int32(2)
		iters := int32(SamplePeriod/time.Millisecond)*(expectedSamples-1) + 100
		const sleep = time.Millisecond
		for i, ev := range events {
			wg.Add(1)

			go func(i int, l Level, msg string) {
				defer wg.Done()

				for j := int32(0); j < iters; j++ {
					sampled := s.Sample(l, msg)
					if sampled {
						results[i].Add(1)
					}

					time.Sleep(sleep)
				}
			}(i, ev.l, ev.msg)
		}
		wg.Wait()

		require.Equal(t, expectedSamples, results[0].Load())
		require.Equal(t, expectedSamples, results[1].Load())
		require.Equal(t, expectedSamples, results[2].Load())
	})
}
