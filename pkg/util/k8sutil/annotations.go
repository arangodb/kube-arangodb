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
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type OwnerRefObj interface {
	GetOwnerReferences() []meta.OwnerReference
}

func IsOwnerFromRef(owner, ref meta.OwnerReference) bool {
	return owner.UID == ref.UID
}

func IsOwner(ref meta.OwnerReference, object OwnerRefObj) bool {
	for _, ownerRef := range object.GetOwnerReferences() {
		if IsOwnerFromRef(ref, ownerRef) {
			return true
		}
	}

	return false
}

func IsChildResource(kind, name, namespace string, resource meta.Object) bool {
	if resource == nil {
		return false
	}

	if namespace != resource.GetNamespace() {
		return false
	}

	ownerRef := resource.GetOwnerReferences()

	if len(ownerRef) == 0 {
		return false
	}

	for _, owner := range ownerRef {
		if owner.Kind != kind {
			continue
		}

		if owner.Name != name {
			continue
		}

		return true
	}

	return false
}
