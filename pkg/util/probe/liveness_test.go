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
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLivenessLock(t *testing.T) {
	p := &LivenessProbe{}
	assert.True(t, p.waitUntilNotLocked(time.Millisecond))

	// Test single lock
	p.Lock()
	assert.False(t, p.waitUntilNotLocked(time.Millisecond))
	p.Unlock()
	assert.True(t, p.waitUntilNotLocked(time.Millisecond))

	// Test multiple locks
	p.Lock()
	assert.False(t, p.waitUntilNotLocked(time.Millisecond))
	p.Lock()
	assert.False(t, p.waitUntilNotLocked(time.Millisecond))
	p.Unlock()
	assert.False(t, p.waitUntilNotLocked(time.Millisecond))
	p.Unlock()
	assert.True(t, p.waitUntilNotLocked(time.Millisecond))

	// Test concurrent waits
	wg := sync.WaitGroup{}
	p.Lock()
	wg.Add(1)
	go func() {
		// Waiter 1
		defer wg.Done()
		assert.True(t, p.waitUntilNotLocked(time.Millisecond*200))
	}()
	wg.Add(1)
	go func() {
		// Waiter 2
		defer wg.Done()
		assert.True(t, p.waitUntilNotLocked(time.Millisecond*200))
	}()
	wg.Add(1)
	go func() {
		// Waiter 3
		defer wg.Done()
		assert.False(t, p.waitUntilNotLocked(time.Millisecond*5))
	}()
	wg.Add(1)
	go func() {
		// Unlocker
		defer wg.Done()
		time.Sleep(time.Millisecond * 50)
		p.Unlock()
	}()
	wg.Wait()
}
