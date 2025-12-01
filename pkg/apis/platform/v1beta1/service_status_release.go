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

package v1beta1

import (
	helmRelease "helm.sh/helm/v3/pkg/release"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arangodb/kube-arangodb/pkg/util"
)

type ArangoPlatformServiceStatusRelease struct {
	Name    string                                 `json:"name,omitempty"`
	Version int                                    `json:"version"`
	Hash    string                                 `json:"hash,omitempty"`
	Info    ArangoPlatformServiceStatusReleaseInfo `json:"info"`
}

func (a *ArangoPlatformServiceStatusRelease) Compare(o *ArangoPlatformServiceStatusRelease) bool {
	if a == nil && o == nil {
		return true
	}
	if a == nil || o == nil {
		return false
	}

	return a.Name == o.Name &&
		a.Version == o.Version &&
		a.Hash == o.Hash &&
		a.Info.Compare(&o.Info)
}

type ArangoPlatformServiceStatusReleaseInfo struct {
	FirstDeployed *meta.Time         `json:"first_deployed,omitempty"`
	LastDeployed  *meta.Time         `json:"last_deployed,omitempty"`
	Status        helmRelease.Status `json:"status,omitempty"`
}

func (a *ArangoPlatformServiceStatusReleaseInfo) Compare(o *ArangoPlatformServiceStatusReleaseInfo) bool {
	if a == nil && o == nil {
		return true
	}
	if a == nil || o == nil {
		return false
	}

	return util.TimeCompareEqualPointer(a.FirstDeployed, o.FirstDeployed) &&
		util.TimeCompareEqualPointer(a.LastDeployed, o.LastDeployed) &&
		a.Status == o.Status
}
