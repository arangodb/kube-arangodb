//
// DISCLAIMER
//
// Copyright 2016-2023 ArangoDB GmbH, Cologne, Germany
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

import (
	policy "k8s.io/api/policy/v1"

	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/gvk"
)

type Inspector interface {
	gvk.GVK

	GetSimple(name string) (*policy.PodDisruptionBudget, bool)
	Iterate(action Action, filters ...Filter) error
	Read() ReadInterface
}

type Filter func(podDisruptionBudget *policy.PodDisruptionBudget) bool
type Action func(podDisruptionBudget *policy.PodDisruptionBudget) error
