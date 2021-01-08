//
// DISCLAIMER
//
// Copyright 2020 ArangoDB GmbH, Cologne, Germany
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
// Author Ewout Prangsma
//

package k8sutil

import (
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/rs/zerolog"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	maxRemoveFinalizersAttempts = 50
)

// RemovePodFinalizers removes the given finalizers from the given pod.
func RemovePodFinalizers(log zerolog.Logger, kubecli kubernetes.Interface, p *v1.Pod, finalizers []string, ignoreNotFound bool) error {
	pods := kubecli.CoreV1().Pods(p.GetNamespace())
	getFunc := func() (metav1.Object, error) {
		result, err := pods.Get(p.GetName(), metav1.GetOptions{})
		if err != nil {
			return nil, errors.WithStack(err)
		}
		return result, nil
	}
	updateFunc := func(updated metav1.Object) error {
		updatedPod := updated.(*v1.Pod)
		result, err := pods.Update(updatedPod)
		if err != nil {
			return errors.WithStack(err)
		}
		*p = *result
		return nil
	}
	if err := RemoveFinalizers(log, finalizers, getFunc, updateFunc, ignoreNotFound); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// RemovePVCFinalizers removes the given finalizers from the given PVC.
func RemovePVCFinalizers(log zerolog.Logger, kubecli kubernetes.Interface, p *v1.PersistentVolumeClaim, finalizers []string, ignoreNotFound bool) error {
	pvcs := kubecli.CoreV1().PersistentVolumeClaims(p.GetNamespace())
	getFunc := func() (metav1.Object, error) {
		result, err := pvcs.Get(p.GetName(), metav1.GetOptions{})
		if err != nil {
			return nil, errors.WithStack(err)
		}
		return result, nil
	}
	updateFunc := func(updated metav1.Object) error {
		updatedPVC := updated.(*v1.PersistentVolumeClaim)
		result, err := pvcs.Update(updatedPVC)
		if err != nil {
			return errors.WithStack(err)
		}
		*p = *result
		return nil
	}
	if err := RemoveFinalizers(log, finalizers, getFunc, updateFunc, ignoreNotFound); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// RemoveFinalizers is a helper used to remove finalizers from an object.
// The functions tries to get the object using the provided get function,
// then remove the given finalizers and update the update using the given update function.
// In case of an update conflict, the functions tries again.
func RemoveFinalizers(log zerolog.Logger, finalizers []string, getFunc func() (metav1.Object, error), updateFunc func(metav1.Object) error, ignoreNotFound bool) error {
	attempts := 0
	for {
		attempts++
		obj, err := getFunc()
		if err != nil {
			if IsNotFound(err) && ignoreNotFound {
				// Object no longer found and we're allowed to ignore that.
				return nil
			}
			log.Warn().Err(err).Msg("Failed to get resource")
			return errors.WithStack(err)
		}
		original := obj.GetFinalizers()
		if len(original) == 0 {
			// We're done
			return nil
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
		if len(newList) < len(original) {
			obj.SetFinalizers(newList)
			if err := updateFunc(obj); IsConflict(err) {
				if attempts > maxRemoveFinalizersAttempts {
					log.Warn().Err(err).Msg("Failed to update resource with fewer finalizers after many attempts")
					return errors.WithStack(err)
				} else {
					// Try again
					continue
				}
			} else if IsNotFound(err) && ignoreNotFound {
				// Object no longer found and we're allowed to ignore that.
				return nil
			} else if err != nil {
				log.Warn().Err(err).Msg("Failed to update resource with fewer finalizers")
				return errors.WithStack(err)
			}
		} else {
			log.Debug().Msg("No finalizers needed removal. Resource unchanged")
		}
		return nil
	}
}
