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

package storage

import (
	"time"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// inspectPVs queries all PersistentVolume's and triggers a cleanup for
// released volumes.
// Returns the number of available PV's.
func (ls *LocalStorage) inspectPVs() (int, error) {
	log := ls.deps.Log
	list, err := ls.deps.KubeCli.CoreV1().PersistentVolumes().List(metav1.ListOptions{})
	if err != nil {
		return 0, errors.WithStack(err)
	}
	spec := ls.apiObject.Spec
	availableVolumes := 0
	cleanupBeforeTimestamp := time.Now().Add(time.Hour * -24)
	for _, pv := range list.Items {
		if pv.Spec.StorageClassName != spec.StorageClass.Name {
			// Not our storage class
			continue
		}
		switch pv.Status.Phase {
		case v1.VolumeAvailable:
			// Is this an old volume?
			if pv.GetObjectMeta().GetCreationTimestamp().Time.Before(cleanupBeforeTimestamp) {
				// Let's clean it up
				if ls.isOwnerOf(&pv) {
					// Cleanup this volume
					log.Debug().Str("name", pv.GetName()).Msg("Added PersistentVolume to cleaner")
					ls.pvCleaner.Add(pv)
				} else {
					log.Debug().Str("name", pv.GetName()).Msg("PersistentVolume is not owned by us")
					availableVolumes++
				}
			} else {
				availableVolumes++
			}
		case v1.VolumeReleased:
			if ls.isOwnerOf(&pv) {
				// Cleanup this volume
				log.Debug().Str("name", pv.GetName()).Msg("Added PersistentVolume to cleaner")
				ls.pvCleaner.Add(pv)
			} else {
				log.Debug().Str("name", pv.GetName()).Msg("PersistentVolume is not owned by us")
			}
		}
	}
	return availableVolumes, nil
}
