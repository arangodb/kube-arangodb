//
// DISCLAIMER
//
// Copyright 2018 ArangoDB GmbH, Cologne, Germany
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
// Author Ewout Prangsma
//

package main

import (
	"sync"
	"time"

	"github.com/rs/zerolog"

	"github.com/arangodb/kube-arangodb/tests/duration/test"
)

const (
	recentFailureTimeout        = time.Hour       // Disregard failures old than this timeout
	requiredRecentFailureSpread = time.Minute * 5 // How far apart the first and last recent failure must be
	requiredRecentFailures      = 30              // At least so many recent failures are needed to fail the test
)

type testListener struct {
	mutex          sync.Mutex
	Log            zerolog.Logger
	FailedCallback func()
	recentFailures []time.Time
	failed         bool
}

var _ test.TestListener = &testListener{}

// ReportFailure logs the given failure and keeps track of recent failure timestamps.
func (l *testListener) ReportFailure(f test.Failure) {
	l.Log.Error().Msg(f.Message)

	// Remove all old recent failures
	l.mutex.Lock()
	defer l.mutex.Unlock()
	for {
		if len(l.recentFailures) == 0 {
			break
		}
		isOld := l.recentFailures[0].Add(recentFailureTimeout).Before(time.Now())
		if isOld {
			// Remove first entry
			l.recentFailures = l.recentFailures[1:]
		} else {
			// First failure is not old, keep the list as is
			break
		}
	}
	l.recentFailures = append(l.recentFailures, time.Now())

	// Detect failed state
	if len(l.recentFailures) > requiredRecentFailures {
		spread := l.recentFailures[len(l.recentFailures)-1].Sub(l.recentFailures[0])
		if spread > requiredRecentFailureSpread {
			l.failed = true
			if l.FailedCallback != nil {
				l.FailedCallback()
			}
		}
	}
}

// IsFailed returns true when the number of recent failures
// has gone above the set maximum, false otherwise.
func (l *testListener) IsFailed() bool {
	return l.failed
}
