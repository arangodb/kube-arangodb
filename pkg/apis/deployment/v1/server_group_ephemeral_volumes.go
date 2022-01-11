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

package v1

import "k8s.io/apimachinery/pkg/api/resource"

// EphemeralVolumes keeps info about ephemeral volumes. Used only with `ephemeral-volumes` feature.
type EphemeralVolumes struct {
	// Apps define apps ephemeral volume in case if `ephemeral-volumes` feature is enabled.
	Apps *EphemeralVolume `json:"apps,omitempty"`
	// Temp define temp ephemeral volume in case if `ephemeral-volumes` feature is enabled.
	Temp *EphemeralVolume `json:"temp,omitempty"`
}

// GetAppsSize returns apps volume size with default value of nil.
func (e *EphemeralVolumes) GetAppsSize() *resource.Quantity {
	return e.getAppsSize(nil)
}

func (e *EphemeralVolumes) getAppsSize(d *resource.Quantity) *resource.Quantity {
	if e == nil {
		return d
	}

	return e.Apps.GetSize(d)
}

// GetTempSize returns temp volume size with default value of nil.
func (e *EphemeralVolumes) GetTempSize() *resource.Quantity {
	return e.getTempSize(nil)
}

func (e *EphemeralVolumes) getTempSize(d *resource.Quantity) *resource.Quantity {
	if e == nil {
		return d
	}

	return e.Temp.GetSize(d)
}

// EphemeralVolume keeps information about ephemeral volumes.
type EphemeralVolume struct {
	Size *resource.Quantity `json:"size"`
}

// GetSize returns size. If not defined, default is returned.
func (e *EphemeralVolume) GetSize(d *resource.Quantity) *resource.Quantity {
	if e == nil || e.Size == nil {
		return d
	}

	return e.Size
}
