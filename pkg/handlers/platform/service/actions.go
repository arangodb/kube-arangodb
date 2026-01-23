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

package service

import (
	"helm.sh/helm/v3/pkg/action"

	platformApi "github.com/arangodb/kube-arangodb/pkg/apis/platform/v1beta1"
)

func withUpgradeActionOverrides(spec *platformApi.ArangoPlatformServiceSpecUpgrade) func(in *action.Upgrade) {
	return func(in *action.Upgrade) {
		in.Timeout = spec.GetTimeout()
		in.MaxHistory = spec.GetMaxHistory()
	}
}

func withInstallActionOverrides(spec *platformApi.ArangoPlatformServiceSpecInstall) func(in *action.Install) {
	return func(in *action.Install) {
		in.Timeout = spec.GetTimeout()
	}
}
