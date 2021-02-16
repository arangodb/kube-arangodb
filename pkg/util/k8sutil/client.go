//
// DISCLAIMER
//
// Copyright 2020 ArangoDB GmbH, Cologne, Germany
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
// Author Ewout Prangsma
//

package k8sutil

import (
	"fmt"
	"os"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"

	monitoringClient "github.com/coreos/prometheus-operator/pkg/client/versioned/typed/monitoring/v1"

	"github.com/arangodb/kube-arangodb/pkg/util"
	"k8s.io/client-go/tools/clientcmd"

	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const Kubeconfig util.EnvironmentVariable = "KUBECONFIG"

// NewKubeConfig loads config from KUBECONFIG or as incluster
func NewKubeConfig() (*rest.Config, error) {
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

// NewKubeClient creates a new k8s client
func NewKubeClient() (kubernetes.Interface, error) {
	cfg, err := NewKubeConfig()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	c, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return c, nil
}

// MustNewKubeClient calls NewKubeClient an panics if it fails
func MustNewKubeClient() kubernetes.Interface {
	i, err := NewKubeClient()
	if err != nil {
		panic(err)
	}
	return i
}

// NewKubeExtClient creates a new k8s api extensions client
func NewKubeExtClient() (apiextensionsclient.Interface, error) {
	cfg, err := NewKubeConfig()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	c, err := apiextensionsclient.NewForConfig(cfg)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return c, nil
}

func NewKubeMonitoringV1Client() (monitoringClient.MonitoringV1Interface, error) {
	cfg, err := NewKubeConfig()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	c, err := monitoringClient.NewForConfig(cfg)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return c, nil
}
