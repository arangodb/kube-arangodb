//
// DISCLAIMER
//
// Copyright 2023 ArangoDB GmbH, Cologne, Germany
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

package timer

import (
	"sync"
	"time"
)

func NewDelayer() Delayer {
	return &delayer{}
}

type Delayer interface {
	Delay(delay time.Duration)

	Wait() time.Duration

	Copy() Delayer
}

type delayer struct {
	lock sync.Mutex

	last, next time.Time
}

func (d *delayer) Copy() Delayer {
	d.lock.Lock()
	defer d.lock.Unlock()

	return &delayer{
		last: d.last,
		next: d.next,
	}
}

func (d *delayer) Wait() time.Duration {
	d.lock.Lock()
	defer d.lock.Unlock()

	since := time.Until(d.next)

	if since <= time.Millisecond {
		return 0
	}

	time.Sleep(since)

	return since
}

func (d *delayer) Delay(delay time.Duration) {
	d.lock.Lock()
	defer d.lock.Unlock()

	d.last = time.Now()
	d.next = d.last.Add(delay)
}
