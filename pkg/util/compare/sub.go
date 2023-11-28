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

package compare

type SubElementExtractor[T1, T2 interface{}] func(in *T1) *T2

func SubElementsP2[T1, T2, P1, P2 interface{}](extractor SubElementExtractor[T1, T2], gens ...GenP2[T2, P1, P2]) GenP2[T1, P1, P2] {
	return func(p1 P1, p2 P2, spec, status *T1) Func {
		specF := extractor(spec)
		statusF := extractor(status)

		funcs := make([]Func, len(gens))

		for id := range gens {
			funcs[id] = gens[id](p1, p2, specF, statusF)
		}

		return Merge(funcs...)
	}
}

func ArrayExtractorP2[T, P1, P2 interface{}](gens ...GenP2[T, P1, P2]) GenP2[[]T, P1, P2] {
	return func(p1 P1, p2 P2, spec, status *[]T) Func {
		if spec == nil || status == nil {
			return SkippedRotation.Func()
		}

		specA := *spec
		statusA := *status

		if len(specA) != len(statusA) {
			return SkippedRotation.Func()
		}

		funcs := make([]Func, 0, len(specA)*len(gens))
		// Iterate over ids
		for id := range specA {
			for _, gen := range gens {
				funcs = append(funcs, gen(p1, p2, &specA[id], &statusA[id]))
			}
		}

		return Merge(funcs...)
	}
}
