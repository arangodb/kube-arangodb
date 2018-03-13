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
package versioned

import (
	databasev1alpha "github.com/arangodb/kube-arangodb/pkg/generated/clientset/versioned/typed/deployment/v1alpha"
	storagev1alpha "github.com/arangodb/kube-arangodb/pkg/generated/clientset/versioned/typed/storage/v1alpha"
	glog "github.com/golang/glog"
	discovery "k8s.io/client-go/discovery"
	rest "k8s.io/client-go/rest"
	flowcontrol "k8s.io/client-go/util/flowcontrol"
)

type Interface interface {
	Discovery() discovery.DiscoveryInterface
	DatabaseV1alpha() databasev1alpha.DatabaseV1alphaInterface
	// Deprecated: please explicitly pick a version if possible.
	Database() databasev1alpha.DatabaseV1alphaInterface
	StorageV1alpha() storagev1alpha.StorageV1alphaInterface
	// Deprecated: please explicitly pick a version if possible.
	Storage() storagev1alpha.StorageV1alphaInterface
}

// Clientset contains the clients for groups. Each group has exactly one
// version included in a Clientset.
type Clientset struct {
	*discovery.DiscoveryClient
	databaseV1alpha *databasev1alpha.DatabaseV1alphaClient
	storageV1alpha  *storagev1alpha.StorageV1alphaClient
}

// DatabaseV1alpha retrieves the DatabaseV1alphaClient
func (c *Clientset) DatabaseV1alpha() databasev1alpha.DatabaseV1alphaInterface {
	return c.databaseV1alpha
}

// Deprecated: Database retrieves the default version of DatabaseClient.
// Please explicitly pick a version.
func (c *Clientset) Database() databasev1alpha.DatabaseV1alphaInterface {
	return c.databaseV1alpha
}

// StorageV1alpha retrieves the StorageV1alphaClient
func (c *Clientset) StorageV1alpha() storagev1alpha.StorageV1alphaInterface {
	return c.storageV1alpha
}

// Deprecated: Storage retrieves the default version of StorageClient.
// Please explicitly pick a version.
func (c *Clientset) Storage() storagev1alpha.StorageV1alphaInterface {
	return c.storageV1alpha
}

// Discovery retrieves the DiscoveryClient
func (c *Clientset) Discovery() discovery.DiscoveryInterface {
	if c == nil {
		return nil
	}
	return c.DiscoveryClient
}

// NewForConfig creates a new Clientset for the given config.
func NewForConfig(c *rest.Config) (*Clientset, error) {
	configShallowCopy := *c
	if configShallowCopy.RateLimiter == nil && configShallowCopy.QPS > 0 {
		configShallowCopy.RateLimiter = flowcontrol.NewTokenBucketRateLimiter(configShallowCopy.QPS, configShallowCopy.Burst)
	}
	var cs Clientset
	var err error
	cs.databaseV1alpha, err = databasev1alpha.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}
	cs.storageV1alpha, err = storagev1alpha.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}

	cs.DiscoveryClient, err = discovery.NewDiscoveryClientForConfig(&configShallowCopy)
	if err != nil {
		glog.Errorf("failed to create the DiscoveryClient: %v", err)
		return nil, err
	}
	return &cs, nil
}

// NewForConfigOrDie creates a new Clientset for the given config and
// panics if there is an error in the config.
func NewForConfigOrDie(c *rest.Config) *Clientset {
	var cs Clientset
	cs.databaseV1alpha = databasev1alpha.NewForConfigOrDie(c)
	cs.storageV1alpha = storagev1alpha.NewForConfigOrDie(c)

	cs.DiscoveryClient = discovery.NewDiscoveryClientForConfigOrDie(c)
	return &cs
}

// New creates a new Clientset for the given RESTClient.
func New(c rest.Interface) *Clientset {
	var cs Clientset
	cs.databaseV1alpha = databasev1alpha.New(c)
	cs.storageV1alpha = storagev1alpha.New(c)

	cs.DiscoveryClient = discovery.NewDiscoveryClient(c)
	return &cs
}
