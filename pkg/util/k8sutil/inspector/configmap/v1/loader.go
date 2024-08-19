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

package v1

import (
	core "k8s.io/api/core/v1"

	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/gvk"
)

// Inspector for configmaps
type Inspector interface {
	gvk.GVK

	ListSimple() []*core.ConfigMap
	GetSimple(name string) (*core.ConfigMap, bool)
	Iterate(action Action, filters ...Filter) error
	Read() ReadInterface
}

type Filter func(pod *core.ConfigMap) bool
type Action func(pod *core.ConfigMap) error
