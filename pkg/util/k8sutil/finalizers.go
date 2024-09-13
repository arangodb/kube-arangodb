//
// DISCLAIMER
//
// Copyright 2016-2024 ArangoDB GmbH, Cologne, Germany
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

package k8sutil

import (
	"context"
	"sort"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/generic"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/kerrors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/patcher"
)

const (
	maxRemoveFinalizersAttempts = 50
)

// RemoveSelectedFinalizers removes the given finalizers from the given pod.
func RemoveSelectedFinalizers[T meta.Object](ctx context.Context, getter generic.GetInterface[T], patcher generic.PatchInterface[T], p T,
	finalizers []string, ignoreNotFound bool) (int, error) {
	if count, err := RemoveFinalizers(ctx, getter, patcher, p.GetName(), finalizers, ignoreNotFound); err != nil {
		return 0, errors.WithStack(err)
	} else {
		return count, nil
	}
}

type RemoveFinalizersClient[T meta.Object] interface {
	generic.GetInterface[T]
	generic.PatchInterface[T]
}

// RemoveFinalizers is a helper used to remove finalizers from an object.
// The functions tries to get the object using the provided get function,
// then remove the given finalizers and update the update using the given update function.
// In case of an update conflict, the functions tries again.
func RemoveFinalizers[T meta.Object](ctx context.Context, getter generic.GetInterface[T], p generic.PatchInterface[T], name string, finalizers []string, ignoreNotFound bool) (int, error) {
	attempts := 0
	for {
		attempts++
		obj, err := getter.Get(ctx, name, meta.GetOptions{})
		if err != nil {
			if kerrors.IsNotFound(err) && ignoreNotFound {
				// Object no longer found and we're allowed to ignore that.
				return 0, nil
			}
			return 0, errors.WithStack(err)
		}
		original := obj.GetFinalizers()
		if len(original) == 0 {
			// We're done
			return 0, nil
		}
		newList := make([]string, 0, len(original))
		shouldRemove := func(f string) bool {
			for _, x := range finalizers {
				if x == f {
					return true
				}
			}
			return false
		}
		for _, f := range original {
			if !shouldRemove(f) {
				newList = append(newList, f)
			}
		}
		if z := len(original) - len(newList); z > 0 {
			if _, _, err := patcher.Patcher[T](ctx, p, obj, meta.PatchOptions{}, patcher.Finalizers[T](newList)); kerrors.IsConflict(err) {
				if attempts > maxRemoveFinalizersAttempts {
					return 0, errors.WithStack(err)
				} else {
					// Try again
					continue
				}
			} else if kerrors.IsNotFound(err) && ignoreNotFound {
				// Object no longer found and we're allowed to ignore that.
				return 0, nil
			} else if err != nil {
				return 0, errors.WithStack(err)
			}
			return z, nil
		} else {
			return 0, nil
		}
	}
}

func EnsureFinalizers(in meta.Object, exists []string, missing []string) bool {
	present := make(map[string]bool, len(in.GetFinalizers()))

	for _, k := range in.GetFinalizers() {
		present[k] = true
	}

	changed := false

	for _, k := range exists {
		if _, ok := present[k]; !ok {
			present[k] = true
			changed = true
		}
	}

	for _, k := range missing {
		if _, ok := present[k]; ok {
			delete(present, k)
			changed = true
		}
	}

	if !changed {
		return false
	}

	q := make([]string, 0, len(present))

	for k := range present {
		q = append(q, k)
	}

	sort.Strings(q)

	in.SetFinalizers(q)
	return true
}
