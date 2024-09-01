//
// DISCLAIMER
//
// Copyright 2024 ArangoDB GmbH, Cologne, Germany
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

package resources

import (
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1beta1/interfaces"
	"github.com/arangodb/kube-arangodb/pkg/util"
	kresources "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/resources"
)

var _ interfaces.Pod[Metadata] = &Metadata{}

type Metadata struct {
	// Map of string keys and values that can be used to organize and categorize
	// (scope and select) objects. May match selectors of replication controllers
	// and services.
	// +doc/link: Kubernetes docs|https://kubernetes.io/docs/concepts/overview/working-with-objects/labels
	Labels map[string]string `json:"labels,omitempty"`

	// Annotations is an unstructured key value map stored with a resource that may be
	// set by external tools to store and retrieve arbitrary metadata. They are not
	// queryable and should be preserved when modifying objects.
	// +doc/link: Kubernetes docs|https://kubernetes.io/docs/concepts/overview/working-with-objects/annotations
	Annotations map[string]string `json:"annotations,omitempty"`

	// List of objects depended by this object. If ALL objects in the list have
	// been deleted, this object will be garbage collected. If this object is managed by a controller,
	// then an entry in this list will point to this controller, with the controller field set to true.
	// There cannot be more than one managing controller.
	// +doc/type: meta.OwnerReference
	OwnerReferences []meta.OwnerReference `json:"ownerReferences,omitempty"`
}

func (m *Metadata) Apply(template *core.PodTemplateSpec) error {
	if m == nil {
		return nil
	}

	z := m.DeepCopy()

	template.Labels = util.MergeMaps(true, template.Labels, z.Labels)
	template.Annotations = util.MergeMaps(true, template.Annotations, z.Annotations)
	template.OwnerReferences = append(template.OwnerReferences, z.OwnerReferences...)

	return nil
}

func (m *Metadata) With(other *Metadata) *Metadata {
	if m == nil && other == nil {
		return nil
	}

	if m == nil {
		return other.DeepCopy()
	}

	if other == nil {
		return m.DeepCopy()
	}

	return &Metadata{
		Labels:          util.MergeMaps(true, m.Labels, other.Labels),
		Annotations:     util.MergeMaps(true, m.Annotations, other.Annotations),
		OwnerReferences: kresources.MergeOwnerReferences(m.OwnerReferences, other.OwnerReferences),
	}
}

func (m *Metadata) Validate() error {
	return nil
}
