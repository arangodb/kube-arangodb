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

package util

import (
	"time"

	"github.com/arangodb/kube-arangodb/pkg/util/timer"
)

// Interval is a specialization of Duration so we can add some
// helper functions to that.
type Interval time.Duration

func (i Interval) String() string {
	return time.Duration(i).String()
}

// ReduceTo returns an interval that is equal to min(x, i).
func (i Interval) ReduceTo(x Interval) Interval {
	if i < x {
		return i
	}
	return x
}

// IncreaseTo returns an interval that is equal to max(x, i).
func (i Interval) IncreaseTo(x Interval) Interval {
	if i > x {
		return i
	}
	return x
}

// Backoff returns an interval that is equal to min(i*factor, maxInt).
func (i Interval) Backoff(factor float64, maxInt Interval) Interval {
	i = Interval(float64(i) * factor)
	if i < maxInt {
		return i
	}
	return maxInt
}

// After waits for the interval to elapse and then sends the current time
// on the returned channel.
func (i Interval) After() <-chan time.Time {
	return timer.After(time.Duration(i))
}
