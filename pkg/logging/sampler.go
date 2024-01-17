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
	"fmt"
	goSync "sync"
	"time"

	"github.com/arangodb/kube-arangodb/pkg/util/sync"
)

// Sampler defines an interface to a log sampler.
type Sampler interface {
	// Sample returns true if the event should be part of the sample, false if
	// the event should be dropped.
	Sample(level Level, msg string, args ...any) bool
}

func NewLogEventSampler(period time.Duration) Sampler {
	return &logEventSampler{
		period: period,
	}
}

type logEventSampler struct {
	// period defines the burst period
	period time.Duration

	// lruCache is the thread-safe map of message hash -> timestamp when event happened last time
	lruCache sync.Map[string, int64]

	gcLock  goSync.Mutex
	gcCycle uint64
}

func (s *logEventSampler) Sample(level Level, format string, args ...any) bool {
	if s.period <= 0 {
		// sampling disabled
		return true
	}

	s.gc()

	hash := s.hash(level, format, args...)
	return s.store(hash)
}

func (s *logEventSampler) hash(level Level, format string, args ...any) string {
	msg := fmt.Sprintf(format, args...)
	return fmt.Sprintf("%s_%s", level.String(), msg)
}

// store returns true if event hash was not stored, or it was stored more than s.period ago
func (s *logEventSampler) store(hash string) bool {
	storedAt, _ := s.lruCache.LoadOrStore(hash, 0)
	now := time.Now().UnixNano()
	if now > storedAt {
		newStoredAt := now + s.period.Nanoseconds()
		return s.lruCache.CompareAndSwap(hash, storedAt, newStoredAt)
	}
	return false
}

func (s *logEventSampler) gc() {
	s.gcLock.Lock()
	defer s.gcLock.Unlock()

	s.gcCycle++
	if s.gcCycle > 10e5 {
		// run cache cleanup every 10e5 cycles
		s.gcClean()
		s.gcCycle = 0
	}
}

func (s *logEventSampler) gcClean() {
	now := time.Now().UnixNano()
	s.lruCache.Range(func(key string, value int64) bool {
		gcPeriod := (time.Minute * 15).Nanoseconds()
		if now-value > gcPeriod {
			s.lruCache.Delete(key)
		}
		return true
	})
}
