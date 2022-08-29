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

package storage

import (
	"context"
	"sort"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/storage/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/server"
)

// Name returns the name of the local storage resource
func (ls *LocalStorage) Name() string {
	return ls.apiObject.Name
}

// LocalPaths returns the local paths (on nodes) of the local storage resource
func (ls *LocalStorage) LocalPaths() []string {
	return ls.apiObject.Spec.LocalPath
}

// StateColor returns a color describing the state of the local storage resource
func (ls *LocalStorage) StateColor() server.StateColor {
	switch ls.status.State {
	case api.LocalStorageStateRunning:
		return server.StateGreen
	case api.LocalStorageStateFailed:
		return server.StateRed
	default:
		return server.StateYellow
	}
}

// StorageClass returns the name of the StorageClass specified in the local storage resource
func (ls *LocalStorage) StorageClass() string {
	return ls.apiObject.Spec.StorageClass.Name
}

// StorageClassIsDefault returns true if the StorageClass used by this local storage resource is supposed to be default
func (ls *LocalStorage) StorageClassIsDefault() bool {
	return ls.apiObject.Spec.StorageClass.IsDefault
}

// Volumes returns all volumes created by the local storage resource
func (ls *LocalStorage) Volumes() []server.Volume {
	list, err := ls.deps.Client.Kubernetes().CoreV1().PersistentVolumes().List(context.Background(), meta.ListOptions{})
	if err != nil {
		ls.log.Err(err).Error("Failed to list persistent volumes")
		return nil
	}
	result := make([]server.Volume, 0, len(list.Items))
	for _, pv := range list.Items {
		if ls.isOwnerOf(&pv) {
			result = append(result, serverVolume(pv))
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name() < result[j].Name()
	})
	return result
}
