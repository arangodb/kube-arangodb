//
// DISCLAIMER
//
// Copyright 2024 ArangoDB GmbH, Cologne, Germany
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

package helpers

import (
	"context"
	"fmt"
	"reflect"
)

type DecisionObject[T Object] struct {
	Checksum string
	Object   T
}

func EmptyDecision[T Object](ctx context.Context, current, expected DecisionObject[T]) (Action, error) {
	return ActionOK, nil
}

type Decision[T Object] func(ctx context.Context, current, expected DecisionObject[T]) (Action, error)

func (d Decision[T]) Call(ctx context.Context, current, expected DecisionObject[T]) (Action, error) {
	if d == nil {
		return EmptyDecision[T](ctx, current, expected)
	}

	return d(ctx, current, expected)
}

func (d Decision[T]) With(other ...Decision[T]) Decision[T] {
	return func(ctx context.Context, current, expected DecisionObject[T]) (Action, error) {
		action, err := d.Call(ctx, current, expected)
		if err != nil {
			return 0, err
		}

		for _, o := range other {
			if action == ActionReplace {
				return ActionReplace, nil
			}

			otherAction, err := o.Call(ctx, current, expected)
			if err != nil {
				return 0, err
			}

			action = action.Or(otherAction)
		}

		return action, nil
	}
}

func ReplaceChecksum[T Object](ctx context.Context, current, expected DecisionObject[T]) (Action, error) {
	if current.Checksum != expected.Checksum {
		return ActionReplace, nil
	}

	return ActionOK, nil
}

func UpdateChecksum[T Object](ctx context.Context, current, expected DecisionObject[T]) (Action, error) {
	if current.Checksum != expected.Checksum {
		return ActionUpdate, nil
	}

	return ActionOK, nil
}

func UpdateOwnerReference[T Object](ctx context.Context, current, expected DecisionObject[T]) (Action, error) {
	if !reflect.DeepEqual(current.Object.GetOwnerReferences(), expected.Object.GetOwnerReferences()) {
		return ActionUpdate, nil
	}

	return ActionOK, nil
}

type CompareStack[T any] func(a, b T) map[string]string

type SimpleCompareStack[T any] func(a, b T) (string, bool)

func FromSimpleCompareStack[T any](name string, c SimpleCompareStack[T]) CompareStack[T] {
	return func(a, b T) map[string]string {
		if message, changed := c(a, b); changed {
			return map[string]string{
				name: message,
			}
		}

		return nil
	}
}

func Compare[T any](a, b T, ins ...CompareStack[T]) map[string]string {
	r := make(map[string]string)

	for _, in := range ins {
		for k, v := range in(a, b) {
			r[k] = v
		}
	}

	return r
}

type SubCompareExtract[T, X any] func(in T) X

func SubCompare[T, X any](prefix string, extract SubCompareExtract[T, X], ins ...CompareStack[X]) CompareStack[T] {
	return func(a, b T) map[string]string {
		res := Compare[X](extract(a), extract(b), ins...)
		r := make(map[string]string, len(res))
		for k := range res {
			r[fmt.Sprintf("%s%s", prefix, k)] = res[k]
		}
		return res
	}
}

func NewImmutableFields[T Object](in ImmutableFields[T]) ImmutableFields[T] {
	return in
}

type ImmutableFields[T Object] func(a, b T, changes map[string]string) Action

func (i ImmutableFields[T]) Evaluate(ins ...CompareStack[T]) Decision[T] {
	return func(ctx context.Context, current, expected DecisionObject[T]) (Action, error) {
		r := make(map[string]string)

		for _, in := range ins {
			for k, v := range in(current.Object, expected.Object) {
				r[k] = v
			}
		}

		return i(current.Object, expected.Object, r), nil
	}
}
