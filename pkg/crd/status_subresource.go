//
// DISCLAIMER
//
// Copyright 2026 ArangoDB GmbH, Cologne, Germany
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

package crd

import (
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

// mirrorStorageStatusSubresource makes the desired CRD spec carry the exact same status subresource
// as the existing (live) CRD on the storage version. It is used when the enable-arango-deployment-status
// feature is disabled: the operator must not touch the status subresource on the storage (v1) version
// — it neither adds it (when the live CRD lacks it) nor removes it (when the chart already ships it) —
// while still applying every other part of the desired spec. desired is mutated in place.
func mirrorStorageStatusSubresource(desired, existing *apiextensions.CustomResourceDefinitionSpec) {
	if desired == nil || existing == nil {
		return
	}

	for di := range desired.Versions {
		dv := &desired.Versions[di]
		if !dv.Storage {
			continue
		}

		live, ok := statusSubresourceForVersion(existing, dv.Name)
		if !ok {
			// The live CRD has no matching version; leave the desired version untouched.
			continue
		}

		setStatusSubresource(dv, live)
	}
}

// statusSubresourceForVersion returns the status subresource of the named version in spec, and whether
// that version exists in spec.
func statusSubresourceForVersion(spec *apiextensions.CustomResourceDefinitionSpec, version string) (*apiextensions.CustomResourceSubresourceStatus, bool) {
	for i := range spec.Versions {
		if spec.Versions[i].Name != version {
			continue
		}
		if spec.Versions[i].Subresources == nil {
			return nil, true
		}
		return spec.Versions[i].Subresources.Status, true
	}
	return nil, false
}

// setStatusSubresource sets (or clears) the status subresource on a single CRD version.
func setStatusSubresource(v *apiextensions.CustomResourceDefinitionVersion, status *apiextensions.CustomResourceSubresourceStatus) {
	if status == nil {
		if v.Subresources != nil {
			v.Subresources.Status = nil
			if v.Subresources.Scale == nil {
				v.Subresources = nil
			}
		}
		return
	}

	if v.Subresources == nil {
		v.Subresources = &apiextensions.CustomResourceSubresources{}
	}
	v.Subresources.Status = status.DeepCopy()
}
