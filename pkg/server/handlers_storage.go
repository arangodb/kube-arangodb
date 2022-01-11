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

package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// LocalStorage is the API implemented by an ArangoLocalStorage.
type LocalStorage interface {
	Name() string
	LocalPaths() []string
	StateColor() StateColor
	StorageClass() string
	StorageClassIsDefault() bool
	Volumes() []Volume
}

// StorageOperator is the API implemented by the storage operator.
type StorageOperator interface {
	// GetLocalStorages returns basic information for all local storages managed by the operator
	GetLocalStorages() ([]LocalStorage, error)
	// GetLocalStorage returns detailed information for a local, managed by the operator, with given name
	GetLocalStorage(name string) (LocalStorage, error)
}

// LocalStorageInfo is the information returned per local storage.
type LocalStorageInfo struct {
	Name                  string     `json:"name"`
	LocalPaths            []string   `json:"local_paths"`
	StateColor            StateColor `json:"state_color"`
	StorageClass          string     `json:"storage_class"`
	StorageClassIsDefault bool       `json:"storage_class_is_default"`
}

// newLocalStorageInfo initializes a LocalStorageInfo for the given LocalStorage.
func newLocalStorageInfo(ls LocalStorage) LocalStorageInfo {
	return LocalStorageInfo{
		Name:                  ls.Name(),
		LocalPaths:            ls.LocalPaths(),
		StateColor:            ls.StateColor(),
		StorageClass:          ls.StorageClass(),
		StorageClassIsDefault: ls.StorageClassIsDefault(),
	}
}

// LocalStorageInfoDetails contains detailed info a local storage
type LocalStorageInfoDetails struct {
	LocalStorageInfo
	Volumes []VolumeInfo `json:"volumes"`
}

// newLocalStorageInfoDetails creates a LocalStorageInfoDetails for the given local storage
func newLocalStorageInfoDetails(ls LocalStorage) LocalStorageInfoDetails {
	vols := ls.Volumes()
	result := LocalStorageInfoDetails{
		LocalStorageInfo: newLocalStorageInfo(ls),
		Volumes:          make([]VolumeInfo, 0, len(vols)),
	}
	for _, v := range vols {
		result.Volumes = append(result.Volumes, newVolumeInfo(v))
	}
	return result
}

// Volume is the API implemented by a volume created in a ArangoLocalStorage.
type Volume interface {
	Name() string
	StateColor() StateColor
	NodeName() string
	Capacity() string
}

// VolumeInfo contained the information returned per volume that is created on behalf of a local storage.
type VolumeInfo struct {
	Name       string     `json:"name"`
	StateColor StateColor `json:"state_color"`
	NodeName   string     `json:"node_name"`
	Capacity   string     `json:"capacity"`
}

// newVolumeInfo creates a VolumeInfo for the given volume
func newVolumeInfo(v Volume) VolumeInfo {
	return VolumeInfo{
		Name:       v.Name(),
		StateColor: v.StateColor(),
		NodeName:   v.NodeName(),
		Capacity:   v.Capacity(),
	}
}

// Handle a GET /api/storage request
func (s *Server) handleGetLocalStorages(c *gin.Context) {
	if o := s.deps.Operators.StorageOperator(); o != nil {
		// Fetch local storages
		stgs, err := o.GetLocalStorages()
		if err != nil {
			sendError(c, err)
		} else {
			result := make([]LocalStorageInfo, len(stgs))
			for i, ls := range stgs {
				result[i] = newLocalStorageInfo(ls)
			}
			c.JSON(http.StatusOK, gin.H{
				"storages": result,
			})
		}
	}
}

// Handle a GET /api/storage/:name request
func (s *Server) handleGetLocalStorageDetails(c *gin.Context) {
	if o := s.deps.Operators.StorageOperator(); o != nil {
		// Fetch deployments
		ls, err := o.GetLocalStorage(c.Params.ByName("name"))
		if err != nil {
			sendError(c, err)
		} else {
			result := newLocalStorageInfoDetails(ls)
			c.JSON(http.StatusOK, result)
		}
	}
}
