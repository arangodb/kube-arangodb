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

package v1beta1

import (
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	kresources "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/resources"
)

type ProfileSelectors struct {
	// Label keeps information about label selector
	// +doc/type: meta.LabelSelector
	Label *meta.LabelSelector `json:"label,omitempty"`
}

func (p *ProfileSelectors) Validate() error {
	if p == nil {
		return nil
	}

	return nil
}

func (p *ProfileSelectors) Select(labels map[string]string) bool {
	if p == nil || p.Label == nil {
		return false
	}

	return kresources.SelectLabels(p.Label, labels)
}
