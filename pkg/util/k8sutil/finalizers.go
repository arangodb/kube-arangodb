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

package k8sutil

import (
	"context"
	"sort"

	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/persistentvolumeclaim"
	persistentvolumeclaimv1 "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/persistentvolumeclaim/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/pod"
	podv1 "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/pod/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/kerrors"
)

const (
	maxRemoveFinalizersAttempts = 50
)

// RemovePodFinalizers removes the given finalizers from the given pod.
func RemovePodFinalizers(ctx context.Context, cachedStatus pod.Inspector, c podv1.ModInterface, p *core.Pod,
	finalizers []string, ignoreNotFound bool) (int, error) {
	getFunc := func() (meta.Object, error) {
		ctxChild, cancel := globals.GetGlobalTimeouts().Kubernetes().WithTimeout(ctx)
		defer cancel()

		result, err := cachedStatus.Pod().V1().Read().Get(ctxChild, p.GetName(), meta.GetOptions{})
		if err != nil {
			return nil, errors.WithStack(err)
		}
		return result, nil
	}
	updateFunc := func(updated meta.Object) error {
		updatedPod := updated.(*core.Pod)
		ctxChild, cancel := globals.GetGlobalTimeouts().Kubernetes().WithTimeout(ctx)
		defer cancel()

		result, err := c.Update(ctxChild, updatedPod, meta.UpdateOptions{})
		if err != nil {
			return errors.WithStack(err)
		}
		*p = *result
		return nil
	}
	if count, err := RemoveFinalizers(finalizers, getFunc, updateFunc, ignoreNotFound); err != nil {
		return 0, errors.WithStack(err)
	} else {
		return count, nil
	}
}

// RemovePVCFinalizers removes the given finalizers from the given PVC.
func RemovePVCFinalizers(ctx context.Context, cachedStatus persistentvolumeclaim.Inspector, c persistentvolumeclaimv1.ModInterface,
	p *core.PersistentVolumeClaim, finalizers []string, ignoreNotFound bool) (int, error) {
	getFunc := func() (meta.Object, error) {
		ctxChild, cancel := globals.GetGlobalTimeouts().Kubernetes().WithTimeout(ctx)
		defer cancel()

		result, err := cachedStatus.PersistentVolumeClaim().V1().Read().Get(ctxChild, p.GetName(), meta.GetOptions{})
		if err != nil {
			return nil, errors.WithStack(err)
		}
		return result, nil
	}
	updateFunc := func(updated meta.Object) error {
		updatedPVC := updated.(*core.PersistentVolumeClaim)
		ctxChild, cancel := globals.GetGlobalTimeouts().Kubernetes().WithTimeout(ctx)
		defer cancel()

		result, err := c.Update(ctxChild, updatedPVC, meta.UpdateOptions{})
		if err != nil {
			return errors.WithStack(err)
		}
		*p = *result
		return nil
	}
	if count, err := RemoveFinalizers(finalizers, getFunc, updateFunc, ignoreNotFound); err != nil {
		return 0, errors.WithStack(err)
	} else {
		return count, nil
	}
}

// RemoveFinalizers is a helper used to remove finalizers from an object.
// The functions tries to get the object using the provided get function,
// then remove the given finalizers and update the update using the given update function.
// In case of an update conflict, the functions tries again.
func RemoveFinalizers(finalizers []string, getFunc func() (meta.Object, error), updateFunc func(meta.Object) error, ignoreNotFound bool) (int, error) {
	attempts := 0
	for {
		attempts++
		obj, err := getFunc()
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
			obj.SetFinalizers(newList)
			if err := updateFunc(obj); kerrors.IsConflict(err) {
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
