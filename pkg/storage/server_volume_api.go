//
// DISCLAIMER
//
// Copyright 2018 ArangoDB GmbH, Cologne, Germany
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
	"github.com/arangodb/kube-arangodb/pkg/server"
	"k8s.io/api/core/v1"
)

type serverVolume v1.PersistentVolume

// Name returns the name of the volume
func (v serverVolume) Name() string {
	return v.ObjectMeta.GetName()
}

// StateColor returns a color representing the state of the volume
func (v serverVolume) StateColor() server.StateColor {
	switch v.Status.Phase {
	default:
		return server.StateYellow
	case v1.VolumeBound:
		return server.StateGreen
	case v1.VolumeFailed:
		return server.StateRed
	}
}

// NodeName returns the name of the node the volume is created on volume
func (v serverVolume) NodeName() string {
	return v.GetAnnotations()[nodeNameAnnotation]
}

// Capacity returns the capacity of the volume in human readable form
func (v serverVolume) Capacity() string {
	c, found := v.Spec.Capacity[v1.ResourceStorage]
	if found {
		return c.String()
	}
	return "?"
}
