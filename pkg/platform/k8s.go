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
	"fmt"

	"github.com/spf13/cobra"

	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
	"github.com/arangodb/kube-arangodb/pkg/util/kconfig"
)

func getKubernetesClient(cmd *cobra.Command) (kclient.Client, error) {
	f := kclient.GetUnattachedFactory()
	v, err := flagKubeconfig.Get(cmd)
	if err != nil {
		return nil, err
	}
	if v == "" {
		f.SetKubeConfigGetter(kclient.NewStaticConfigGetter(kconfig.NewConfig))
	} else {
		f.SetKubeConfigGetter(kclient.NewStaticConfigGetter(kconfig.NewFileConfig(v)))
	}

	if err := f.Refresh(); err != nil {
		return nil, err
	}

	if c, ok := f.Client(); !ok {
		return nil, fmt.Errorf("unable to find Kubernetes client")
	} else {
		return c, nil
	}
}
