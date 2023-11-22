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

package errors

type WithErrorArrayP2[IN, P1, P2 any] func(p1 P1, p2 P2, in IN) error

func ExecuteWithErrorArrayP2[IN, P1, P2 any](caller WithErrorArrayP2[IN, P1, P2], p1 P1, p2 P2, elements ...IN) error {
	errors := make([]error, len(elements))

	for id := range elements {
		errors[id] = caller(p1, p2, elements[id])
	}

	return Errors(errors...)
}
