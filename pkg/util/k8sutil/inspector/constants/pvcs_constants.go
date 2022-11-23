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

package constants

import (
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// PersistentVolumeClaim
const (
	PersistentVolumeClaimGroup     = core.GroupName
	PersistentVolumeClaimResource  = "persistentvolumeclaims"
	PersistentVolumeClaimKind      = "PersistentVolumeClaim"
	PersistentVolumeClaimVersionV1 = "v1"
)

func PersistentVolumeClaimGK() schema.GroupKind {
	return schema.GroupKind{
		Group: PersistentVolumeClaimGroup,
		Kind:  PersistentVolumeClaimKind,
	}
}

func PersistentVolumeClaimGKv1() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   PersistentVolumeClaimGroup,
		Kind:    PersistentVolumeClaimKind,
		Version: PersistentVolumeClaimVersionV1,
	}
}

func PersistentVolumeClaimGR() schema.GroupResource {
	return schema.GroupResource{
		Group:    PersistentVolumeClaimGroup,
		Resource: PersistentVolumeClaimResource,
	}
}

func PersistentVolumeClaimGRv1() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    PersistentVolumeClaimGroup,
		Resource: PersistentVolumeClaimResource,
		Version:  PersistentVolumeClaimVersionV1,
	}
}
