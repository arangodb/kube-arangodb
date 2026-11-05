//
// DISCLAIMER
//
// Copyright 2025-2026 ArangoDB GmbH, Cologne, Germany
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
	utilConstants "github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/helm"
)

// IsPlatformManaged reports whether the Helm release is managed by the ArangoDB platform operator. It
// keys solely on the `managed` label; the `type` label (platform vs service) only categorizes the
// release and must not gate ownership, so both platform- and service-typed releases are recognized.
func IsPlatformManaged(r *helm.Release) bool {
	if r == nil {
		return false
	}

	return Label(r, utilConstants.HelmLabelArangoDBManaged) == "true"
}

// Label returns the value of the given release label, or "" if the release or label is absent.
func Label(r *helm.Release, key string) string {
	if r == nil {
		return ""
	}
	return r.Labels[key]
}

// Chart returns the ArangoPlatformChart name a release was installed from, from the `chart` label.
func Chart(r *helm.Release) string {
	return Label(r, utilConstants.HelmLabelArangoDBChart)
}

// DeploymentName returns the owning ArangoDeployment name recorded on a release.
func DeploymentName(r *helm.Release) string {
	return Label(r, utilConstants.LabelArangoDBDeploymentName)
}

// Type returns the platform type (platform or service) recorded on a release.
func Type(r *helm.Release) utilConstants.HelmType {
	return utilConstants.HelmType(Label(r, utilConstants.HelmLabelArangoDBType))
}

// WithDeploymentName adds the deployment-name release label used by the SchedulerV2 integration to
// discover releases. Kept 1:1 with SchedulerV2 so workflow-installed releases match its selector.
func WithDeploymentName(deployment string) util.ModR[map[string]string] {
	return func(m map[string]string) map[string]string {
		m[utilConstants.LabelArangoDBDeploymentName] = deployment
		return m
	}
}

// WithType overrides the platform type label (platform or service) on a release.
func WithType(t utilConstants.HelmType) util.ModR[map[string]string] {
	return func(m map[string]string) map[string]string {
		m[utilConstants.HelmLabelArangoDBType] = t.String()
		return m
	}
}

func GetLabels(deployment, chart string, mods ...util.ModR[map[string]string]) map[string]string {
	m := map[string]string{
		utilConstants.HelmLabelArangoDBManaged:    "true",
		utilConstants.HelmLabelArangoDBDeployment: deployment,
		utilConstants.HelmLabelArangoDBChart:      chart,
		utilConstants.HelmLabelArangoDBType:       utilConstants.HelmTypePlatform.String(),
	}

	for _, mod := range mods {
		m = mod(m)
	}

	return m
}
