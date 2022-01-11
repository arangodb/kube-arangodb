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

// Map of the States and possible transitions
type Map map[State][]State

// Exists checks if State is defined in transitions
func (m Map) Exists(state State) error {
	if _, ok := m[state]; ok {
		return nil
	}

	return NotFound{state: state}
}

// Transit checks if change from one State to another is possible with current defined transitions
func (m Map) Transit(from, to State) error {
	if err := m.Exists(from); err != nil {
		return err
	}

	if err := m.Exists(to); err != nil {
		return err
	}

	if from == to {
		return nil
	}

	for _, targetState := range m[from] {
		if targetState == to {
			return nil
		}
	}

	return ChangeNotPossible{
		from: from,
		to:   to,
	}
}
