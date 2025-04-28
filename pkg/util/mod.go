//
// DISCLAIMER
//
// Copyright 2024-2025 ArangoDB GmbH, Cologne, Germany
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

package util

func emptyMod[T any](_ *T) {}

type Mod[T any] func(in *T)

func (m Mod[T]) Optional() Mod[T] {
	if m == nil {
		return emptyMod[T]
	}

	return m
}

func ApplyMods[T any](in *T, mods ...Mod[T]) {
	for _, mod := range mods {
		if mod != nil {
			mod(in)
		}
	}
}

func emptyModE[T any](_ *T) error {
	return nil
}

type ModE[T any] func(in *T) error

func (m ModE[T]) Optional() ModE[T] {
	if m == nil {
		return emptyModE[T]
	}

	return m
}

func ApplyModsE[T any](in *T, mods ...ModE[T]) error {
	for _, mod := range mods {
		if mod != nil {
			if err := mod(in); err != nil {
				return err
			}
		}
	}

	return nil
}

func emptyModEP1[T, P1 any](_ *T, _ P1) error {
	return nil
}

type ModEP1[T, P1 any] func(in *T, p1 P1) error

func (m ModEP1[T, P1]) Optional() ModEP1[T, P1] {
	if m == nil {
		return emptyModEP1[T, P1]
	}

	return m
}

func ApplyModsEP1[T, P1 any](in *T, p1 P1, mods ...ModEP1[T, P1]) error {
	for _, mod := range mods {
		if mod != nil {
			if err := mod(in, p1); err != nil {
				return err
			}
		}
	}

	return nil
}

func emptyModR[T any](z T) T { return z }

type ModR[T any] func(in T) T

func (m ModR[T]) Optional() ModR[T] {
	if m == nil {
		return emptyModR[T]
	}

	return m
}

func ApplyModsR[T any](in T, mods ...ModR[T]) T {
	for _, mod := range mods {
		if mod != nil {
			mod(in)
		}
	}
	return in
}
