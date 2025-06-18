//
// DISCLAIMER
//
// Copyright 2025 ArangoDB GmbH, Cologne, Germany
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

package labels

import (
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/helm"
)

func IsPlatformManaged(r *helm.Release) bool {
	if r == nil {
		return false
	}

	if managed, ok := r.Labels[constants.HelmLabelArangoDBManaged]; !ok || managed != "true" {
		return false
	}

	if managed, ok := r.Labels[constants.HelmLabelArangoDBType]; !ok || managed != "platform" {
		return false
	}

	return true
}

func GetLabels(deployment, chart string, mods ...util.ModR[map[string]string]) map[string]string {
	m := map[string]string{
		constants.HelmLabelArangoDBManaged:    "true",
		constants.HelmLabelArangoDBDeployment: deployment,
		constants.HelmLabelArangoDBChart:      chart,
		constants.HelmLabelArangoDBType:       "platform",
	}

	for _, mod := range mods {
		m = mod(m)
	}

	return m
}
