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

package patcher

import (
	"context"
	"slices"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arangodb/kube-arangodb/pkg/deployment/patch"
)

// EnsureFinalizersGone ensures that finalizers are gone
func EnsureFinalizersGone[T meta.Object](ctx context.Context, client Client[T], in T, finalizers ...string) (bool, error) {
	current := in.GetFinalizers()
	expected := make([]string, 0, len(current))

	for _, finalizer := range current {
		if !slices.Contains(finalizers, finalizer) {
			expected = append(expected, finalizer)
		}
	}

	if len(expected) == len(current) {
		return false, nil
	}

	_, changed, err := Patcher[T](ctx, client, in, meta.PatchOptions{}, Finalizers[T](expected))
	if err != nil {
		return false, err
	}

	return changed, nil
}

// EnsureFinalizersPresent ensures that finalizers are present
func EnsureFinalizersPresent[T meta.Object](ctx context.Context, client Client[T], in T, finalizers ...string) (bool, error) {
	if in.GetDeletionTimestamp() != nil {
		return false, nil
	}

	current := in.GetFinalizers()

	expected := make([]string, len(current), len(current)+len(finalizers))
	copy(expected, current)

	for _, finalizer := range finalizers {
		if !slices.Contains(expected, finalizer) {
			expected = append(expected, finalizer)
		}
	}

	if len(expected) == len(current) {
		return false, nil
	}

	_, changed, err := Patcher[T](ctx, client, in, meta.PatchOptions{}, Finalizers[T](expected))
	if err != nil {
		return false, err
	}

	return changed, nil
}

func Finalizers[T meta.Object](finalizers []string) Patch[T] {
	return func(in T) []patch.Item {
		if len(finalizers) == 0 {
			return []patch.Item{
				patch.ItemRemove(patch.NewPath("metadata", "finalizers")),
			}
		}

		return []patch.Item{
			patch.ItemReplace(patch.NewPath("metadata", "finalizers"), finalizers),
		}
	}
}
