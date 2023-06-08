//
// DISCLAIMER
//
// Copyright 2016-2023 ArangoDB GmbH, Cologne, Germany
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

package state

import (
	"sync"
	"testing"
)

var (
	currentID int
	idLock    sync.Mutex
)

func id() int {
	idLock.Lock()
	defer idLock.Unlock()

	z := currentID
	currentID++
	return z
}

type Generator func(t *testing.T, s *State)

func GenerateState(t *testing.T, generators ...Generator) State {
	var s State

	for _, g := range generators {
		g(t, &s)
	}

	return s
}
