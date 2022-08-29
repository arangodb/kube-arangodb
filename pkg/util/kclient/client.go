//
// DISCLAIMER
//
// Copyright 2016-2022 ArangoDB GmbH, Cologne, Germany
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

package kclient

import (
	"fmt"
	"os"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/arangodb/kube-arangodb/pkg/util"
)

const Kubeconfig util.EnvironmentVariable = "KUBECONFIG"

// newKubeConfig loads config from KUBECONFIG or as incluster
func newKubeConfig() (*rest.Config, error) {
	// If KUBECONFIG is defined use this variable
	if kubeconfig, ok := Kubeconfig.Lookup(); ok {
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}

	// Try to load incluster config
	if cfg, err := rest.InClusterConfig(); err == nil {
		return cfg, nil
	} else if err != rest.ErrNotInCluster {
		return nil, err
	}

	// At the end try to use default path
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	return clientcmd.BuildConfigFromFlags("", fmt.Sprintf("%s/.kube/config", home))
}
