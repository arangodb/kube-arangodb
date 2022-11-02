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
	"sync"

	"github.com/dchest/uniuri"
	"github.com/pkg/errors"
	monitoring "github.com/prometheus-operator/prometheus-operator/pkg/client/versioned"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/arangodb/kube-arangodb/pkg/generated/clientset/versioned"
)

const (
	unattachedFactoryID = "unattached-factory-id"
)

var (
	factories     = map[string]*factory{}
	factoriesLock sync.Mutex
)

func init() {
	f := GetDefaultFactory()

	f.SetKubeConfigGetter(NewStaticConfigGetter(newKubeConfig))

	if err := f.Refresh(); err != nil {
		println("Error while getting client: ", err.Error())
	}
}

func GetDefaultFactory() Factory {
	return GetFactory("")
}

func GetFactory(name string) Factory {
	factoriesLock.Lock()
	defer factoriesLock.Unlock()

	if f, ok := factories[name]; ok {
		return f
	}

	factories[name] = &factory{
		name: name,
	}

	return factories[name]
}

func GetUnattachedFactory() Factory {
	return &factory{
		name: unattachedFactoryID,
	}
}

type ConfigGetter func() (*rest.Config, string, error)

func NewStaticConfigGetter(f func() (*rest.Config, error)) ConfigGetter {
	u := uniuri.NewLen(32)
	return func() (*rest.Config, string, error) {
		if f == nil {
			return nil, "", errors.Errorf("Provided generator is empty")
		}
		cfg, err := f()
		return cfg, u, err
	}
}

type Factory interface {
	SetKubeConfigGetter(getter ConfigGetter)
	Refresh() error
	SetClient(c Client)

	Name() string

	Client() (Client, bool)
}

type factory struct {
	lock sync.RWMutex

	name string

	getter ConfigGetter

	kubeConfigChecksum string

	client Client
}

func (f *factory) Name() string {
	return f.name
}

func (f *factory) Refresh() error {
	return f.refresh()
}

func (f *factory) SetClient(c Client) {
	f.lock.Lock()
	defer f.lock.Unlock()

	f.client = c
}

func (f *factory) SetKubeConfigGetter(getter ConfigGetter) {
	f.lock.Lock()
	defer f.lock.Unlock()

	f.getter = getter
	f.client = nil
}

func (f *factory) refresh() error {
	if f.getter == nil {
		return errors.Errorf("Getter is nil")
	}

	cfg, checksum, err := f.getter()
	if err != nil {
		return err
	}

	f.lock.Lock()
	defer f.lock.Unlock()

	if f.client != nil && checksum == f.kubeConfigChecksum {
		return nil
	}

	if f.name == unattachedFactoryID {
		cfg.RateLimiter = GetUnattachedRateLimiter()
	} else {
		cfg.RateLimiter = GetRateLimiter(f.name)
	}

	client, err := newClient(f.name, cfg)
	if err != nil {
		return err
	}

	f.client = client
	f.kubeConfigChecksum = checksum

	return nil
}

func (f *factory) Client() (Client, bool) {
	f.lock.RLock()
	defer f.lock.RUnlock()

	if f.client == nil {
		return nil, false
	}

	return f.client, true
}

type Client interface {
	Kubernetes() kubernetes.Interface
	KubernetesExtensions() apiextensionsclient.Interface
	Arango() versioned.Interface
	Monitoring() monitoring.Interface

	Name() string

	Config() *rest.Config
}

func NewStaticClient(kubernetes kubernetes.Interface, kubernetesExtensions apiextensionsclient.Interface, arango versioned.Interface, monitoring monitoring.Interface) Client {
	return &client{
		name:                 "static",
		kubernetes:           kubernetes,
		kubernetesExtensions: kubernetesExtensions,
		arango:               arango,
		monitoring:           monitoring,
	}
}

func newClient(name string, cfg *rest.Config) (*client, error) {
	var c client

	c.config = cfg
	c.name = name

	if q, err := kubernetes.NewForConfig(cfg); err != nil {
		return nil, err
	} else {
		c.kubernetes = q
	}

	if q, err := apiextensionsclient.NewForConfig(cfg); err != nil {
		return nil, err
	} else {
		c.kubernetesExtensions = q
	}

	if q, err := versioned.NewForConfig(cfg); err != nil {
		return nil, err
	} else {
		c.arango = q
	}

	if q, err := monitoring.NewForConfig(cfg); err != nil {
		return nil, err
	} else {
		c.monitoring = q
	}

	return &c, nil
}

type client struct {
	name string

	kubernetes           kubernetes.Interface
	kubernetesExtensions apiextensionsclient.Interface
	arango               versioned.Interface
	monitoring           monitoring.Interface
	config               *rest.Config
}

func (c *client) Name() string {
	return c.name
}

func (c *client) Config() *rest.Config {
	return c.config
}

func (c *client) Kubernetes() kubernetes.Interface {
	return c.kubernetes
}

func (c *client) KubernetesExtensions() apiextensionsclient.Interface {
	return c.kubernetesExtensions
}

func (c *client) Arango() versioned.Interface {
	return c.arango
}

func (c *client) Monitoring() monitoring.Interface {
	return c.monitoring
}
