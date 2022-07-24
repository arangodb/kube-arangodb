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

package probe

import (
	"net/http"
	"sync"
	"time"

	"github.com/arangodb/kube-arangodb/pkg/util/timer"
)

const (
	livenessHandlerTimeout = time.Second * 5
)

// LivenessProbe wraps a liveness probe handler.
type LivenessProbe struct {
	lock     int32
	mutex    sync.Mutex
	waitChan chan struct{}
}

// Lock the probe, preventing the LivenessHandler from responding to requests.
func (p *LivenessProbe) Lock() {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.lock++
}

// Unlock the probe, allowing the LivenessHandler to respond to requests.
func (p *LivenessProbe) Unlock() {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.lock--

	if p.lock == 0 && p.waitChan != nil {
		w := p.waitChan
		p.waitChan = nil
		close(w)
	}
}

// waitUntilNotLocked blocks until the probe is no longer locked
// or a timeout occurs.
// Returns true if the probe is unlocked, false on timeout.
func (p *LivenessProbe) waitUntilNotLocked(timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for {
		var w chan struct{}
		p.mutex.Lock()
		locked := p.lock != 0
		if locked {
			if p.waitChan == nil {
				p.waitChan = make(chan struct{})
			}
			w = p.waitChan
		}
		p.mutex.Unlock()
		if !locked {
			// All good
			return true
		}
		// We're locked, wait until w is closed
		select {
		case <-w:
			// continue
		case <-timer.After(time.Until(deadline)):
			// Timeout
			return false
		}
	}
}

// LivenessHandler writes back the HTTP status code 200 if the operator is ready, and 500 otherwise.
func (p *LivenessProbe) LivenessHandler(w http.ResponseWriter, r *http.Request) {
	if p.waitUntilNotLocked(livenessHandlerTimeout) {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
}
