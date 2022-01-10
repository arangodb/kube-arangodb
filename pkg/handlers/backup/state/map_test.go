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

package state

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMap(t *testing.T) {
	// Arrange
	var state State = "Test"
	var target State = "Target"
	var missingState State = "Missing"

	states := Map{
		state:  []State{target},
		target: []State{},
	}

	// Act/Assert
	assert.EqualError(t, states.Transit(missingState, state), NotFound{state: missingState}.Error())
	assert.EqualError(t, states.Transit(state, missingState), NotFound{state: missingState}.Error())

	assert.NoError(t, states.Transit(state, target))

	assert.NoError(t, states.Transit(state, state))

	assert.EqualError(t, states.Transit(target, state), ChangeNotPossible{from: target, to: state}.Error())
}
