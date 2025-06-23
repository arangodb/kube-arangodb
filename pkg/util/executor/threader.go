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

package executor

import "sync"

func NewThreadManager(threads int) ThreadManager {
	r := make(chan ThreadID, threads)

	for id := 0; id < threads; id++ {
		r <- ThreadID(id)
	}

	return &threadManager{
		threads: r,
	}
}

type ThreadID int

type ThreadManager interface {
	Acquire() Thread
}

type threadManager struct {
	threads chan ThreadID
}

func (t *threadManager) Acquire() Thread {
	id := <-t.threads

	return &thread{
		parent: t,
		id:     id,
	}
}

type Thread interface {
	ID() ThreadID

	Release()
}

type thread struct {
	lock sync.Mutex

	parent *threadManager

	released bool

	id ThreadID
}

func (t *thread) ID() ThreadID {
	return t.id
}

func (t *thread) Release() {
	t.lock.Lock()
	defer t.lock.Unlock()

	if t.released {
		return
	}

	t.released = true

	t.parent.threads <- t.id
}

func (t *thread) Wait() {
	t.lock.Lock()
	defer t.lock.Unlock()

	t.parent.threads <- t.id

	t.id = <-t.parent.threads
}
