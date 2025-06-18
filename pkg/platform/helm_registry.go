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

package platform

import (
	"github.com/spf13/cobra"

	platformApi "github.com/arangodb/kube-arangodb/pkg/apis/platform/v1alpha1"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/helm"
)

func fetchLocallyInstalledCharts(cmd *cobra.Command) (map[string]*platformApi.ArangoPlatformChart, error) {
	client, err := getKubernetesClient(cmd)
	if err != nil {
		return nil, err
	}

	namespace, err := flagNamespace.Get(cmd)
	if err != nil {
		return nil, err
	}

	return helm.GetLocalCharts(cmd.Context(), client, namespace)
}
