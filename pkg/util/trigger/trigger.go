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

package trigger

import "sync"

// Trigger is a synchronization utility used to wait (in a select statement)
// until someone triggers it.
type Trigger struct {
	mu              sync.Mutex
	done            chan struct{}
	pendingTriggers int
}

// Done returns the channel to use in a select case.
// This channel is closed when someone calls Trigger.
func (t *Trigger) Done() <-chan struct{} {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.done == nil {
		t.done = make(chan struct{})
	}
	if t.pendingTriggers > 0 {
		t.pendingTriggers = 0
		d := t.done
		close(t.done)
		t.done = nil
		return d
	}
	return t.done
}

// Trigger closes any Done channel.
func (t *Trigger) Trigger() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.pendingTriggers++
	if t.done != nil {
		close(t.done)
		t.done = nil
	}
}
