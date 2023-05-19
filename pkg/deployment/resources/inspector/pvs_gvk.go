//
// DISCLAIMER
//
// Copyright 2023 ArangoDB GmbH, Cologne, Germany
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

package inspector

import (
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/constants"
)

func (p *persistentVolumesInspectorV1) GroupVersionKind() schema.GroupVersionKind {
	return constants.PersistentVolumeGKv1()
}

func (p *persistentVolumesInspectorV1) GroupVersionResource() schema.GroupVersionResource {
	return constants.PersistentVolumeGRv1()
}

func (p *persistentVolumesInspector) GroupKind() schema.GroupKind {
	return constants.PersistentVolumeGK()
}

func (p *persistentVolumesInspector) GroupResource() schema.GroupResource {
	return constants.PersistentVolumeGR()
}
