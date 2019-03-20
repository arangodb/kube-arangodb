//
// DISCLAIMER
//
// Copyright 2018 ArangoDB GmbH, Cologne, Germany
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
	"net"
	"os"

	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// NewKubeClient creates a new k8s client
func NewKubeClient() (kubernetes.Interface, error) {
	cfg, err := InClusterConfig()
	if err != nil {
		return nil, maskAny(err)
	}
	c, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, maskAny(err)
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
	cfg, err := InClusterConfig()
	if err != nil {
		return nil, maskAny(err)
	}
	c, err := apiextensionsclient.NewForConfig(cfg)
	if err != nil {
		return nil, maskAny(err)
	}
	return c, nil
}

// InClusterConfig loads the environment into a rest config.
func InClusterConfig() (*rest.Config, error) {
	// Work around https://github.com/kubernetes/kubernetes/issues/40973
	// See https://github.com/coreos/etcd-operator/issues/731#issuecomment-283804819
	if len(os.Getenv("KUBERNETES_SERVICE_HOST")) == 0 {
		addrs, err := net.LookupHost("kubernetes.default.svc")
		if err != nil {
			return nil, maskAny(err)
		}
		os.Setenv("KUBERNETES_SERVICE_HOST", addrs[0])
	}
	if len(os.Getenv("KUBERNETES_SERVICE_PORT")) == 0 {
		os.Setenv("KUBERNETES_SERVICE_PORT", "443")
	}
	cfg, err := rest.InClusterConfig()
	if err != nil {
		return nil, maskAny(err)
	}
	return cfg, nil
}
