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

package v2alpha1

import (
	"sort"
	"strings"
)

type DeploymentStatusAgencySize int

func (d *DeploymentStatusAgencySize) Equal(b *DeploymentStatusAgencySize) bool {
	if d == nil && b == nil {
		return true
	}

	if d == nil || b == nil {
		return true
	}

	return *d == *b
}

type DeploymentStatusAgencyIDs []string

func (d DeploymentStatusAgencyIDs) Sort() {
	sort.Slice(d, func(i, j int) bool {
		return strings.Compare(d[i], d[j]) > 0
	})
}

func (d DeploymentStatusAgencyIDs) Equal(b DeploymentStatusAgencyIDs) bool {
	if len(d) != len(b) {
		return false
	}

	for id := range d {
		if d[id] != b[id] {
			return false
		}
	}

	return true
}

type DeploymentStatusAgencyInfo struct {
	Size *DeploymentStatusAgencySize `json:"size,omitempty"`
	IDs  DeploymentStatusAgencyIDs   `json:"ids,omitempty"`
}

func (d *DeploymentStatusAgencyInfo) Equal(b *DeploymentStatusAgencyInfo) bool {
	if d == nil && b == nil {
		return true
	}

	if d == nil || b == nil {
		return true
	}

	return d.IDs.Equal(b.IDs) && d.Size.Equal(b.Size)
}
