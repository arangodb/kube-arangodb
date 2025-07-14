//
// DISCLAIMER
//
// Copyright 2024-2025 ArangoDB GmbH, Cologne, Germany
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

package helm

import (
	"context"
	"fmt"
	goHttp "net/http"
	"sync"

	"helm.sh/helm/v3/pkg/action"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/flowcontrol"

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/kconfig"
)

func NewClient(cfg Configuration) (Client, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	var helm action.Configuration

	if err := helm.Init(kconfig.NewRESTClientGetter(cfg.Namespace, nil, cfg.Config), cfg.Namespace, string(cfg.Driver.Get()), func(format string, v ...interface{}) {
		logger.Debug(format, v...)
	}); err != nil {
		return nil, err
	}

	dClient, err := discovery.NewDiscoveryClientForConfig(cfg.Config)
	if err != nil {
		return nil, err
	}

	return &client{
		cfg:         cfg,
		helm:        &helm,
		discovery:   memory.NewMemCacheClient(dClient),
		restClients: map[schema.GroupVersion]rest.Interface{},
	}, nil
}

type Client interface {
	Namespace() string

	Alive(ctx context.Context) error

	Invalidate()
	DiscoverKubernetesApiVersions(gv schema.GroupVersion) (ApiVersions, error)
	DiscoverKubernetesApiVersionKind(gvk schema.GroupVersionKind) (*meta.APIResource, error)
	NativeGet(ctx context.Context, reqs ...Resource) ([]ResourceObject, error)

	StatusObjects(ctx context.Context, name string, mods ...util.Mod[action.Status]) (*Release, []ResourceObject, error)

	Status(ctx context.Context, name string, mods ...util.Mod[action.Status]) (*Release, error)
	List(ctx context.Context, mods ...util.Mod[action.List]) ([]Release, error)
	Install(ctx context.Context, chart Chart, values Values, mods ...util.Mod[action.Install]) (*Release, error)
	Upgrade(ctx context.Context, name string, chart Chart, values Values, mods ...util.Mod[action.Upgrade]) (*UpgradeResponse, error)
	Uninstall(ctx context.Context, name string, mods ...util.Mod[action.Uninstall]) (*UninstallRelease, error)
	Test(ctx context.Context, name string, mods ...util.Mod[action.ReleaseTesting]) (*Release, error)
}

type client struct {
	lock sync.Mutex

	cfg  Configuration
	helm *action.Configuration

	restClients map[schema.GroupVersion]rest.Interface

	discovery discovery.CachedDiscoveryInterface
}

func (c *client) Namespace() string {
	return c.cfg.Namespace
}

func (c *client) Status(ctx context.Context, name string, mods ...util.Mod[action.Status]) (*Release, error) {
	act := action.NewStatus(c.helm)

	act.ShowResources = true

	util.ApplyMods(act, mods...)

	result, err := act.Run(name)
	if err != nil {
		if err.Error() == "release: not found" {
			return nil, nil
		}
		return nil, err
	}

	if r, err := fromHelmRelease(result); err != nil {
		return nil, err
	} else {
		return &r, nil
	}
}

func (c *client) Uninstall(ctx context.Context, name string, mods ...util.Mod[action.Uninstall]) (*UninstallRelease, error) {
	act := action.NewUninstall(c.helm)

	util.ApplyMods(act, mods...)

	result, err := act.Run(name)
	if err != nil {
		return nil, err
	}

	if result == nil {
		return nil, nil
	}

	var res UninstallRelease

	res.Info = result.Info

	if r, err := fromHelmRelease(result.Release); err != nil {
		return nil, err
	} else {
		res.Release = r
	}

	return &res, nil
}

func (c *client) Test(ctx context.Context, name string, mods ...util.Mod[action.ReleaseTesting]) (*Release, error) {
	act := action.NewReleaseTesting(c.helm)

	util.ApplyMods(act, mods...)

	result, err := act.Run(name)
	if err != nil {
		return nil, err
	}

	if r, err := fromHelmRelease(result); err != nil {
		return nil, err
	} else {
		return &r, nil
	}
}

func (c *client) Install(ctx context.Context, chart Chart, values Values, mods ...util.Mod[action.Install]) (*Release, error) {
	act := action.NewInstall(c.helm)

	util.ApplyMods(act, mods...)

	chartData, err := chart.Get()
	if err != nil {
		return nil, err
	}

	valuesData, err := values.Marshal()
	if err != nil {
		return nil, err
	}

	result, err := act.RunWithContext(ctx, chartData.Chart(), valuesData)
	if err != nil {
		return nil, err
	}

	if r, err := fromHelmRelease(result); err != nil {
		return nil, err
	} else {
		return &r, nil
	}
}

func (c *client) Upgrade(ctx context.Context, name string, chart Chart, values Values, mods ...util.Mod[action.Upgrade]) (*UpgradeResponse, error) {
	act := action.NewUpgrade(c.helm)

	util.ApplyMods(act, mods...)

	release, err := c.Status(ctx, name)
	if err != nil {
		return nil, err
	}

	chartData, err := chart.Get()
	if err != nil {
		return nil, err
	}

	valuesData, err := values.Marshal()
	if err != nil {
		return nil, err
	}

	if release != nil {
		if meta := chartData.Chart().Metadata; meta != nil {
			if release.GetChart().GetMetadata().GetVersion() == meta.Version {
				// We are on the same version
				if release.Values.Equals(values) {
					// We provide same values
					return &UpgradeResponse{
						Before: release,
					}, nil
				}
			}
		}
	}

	result, err := act.RunWithContext(ctx, name, chartData.Chart(), valuesData)
	if err != nil {
		return nil, err
	}

	if r, err := fromHelmRelease(result); err != nil {
		return nil, err
	} else {
		if release == nil {
			return &UpgradeResponse{
				After: &r,
			}, nil
		}
		return &UpgradeResponse{
			Before: release,
			After:  &r,
		}, nil
	}
}

func (c *client) Alive(ctx context.Context) error {
	act := action.NewList(c.helm)

	_, err := act.Run()
	if err != nil {
		return err
	}

	return nil
}

func (c *client) List(ctx context.Context, mods ...util.Mod[action.List]) ([]Release, error) {
	act := action.NewList(c.helm)

	util.ApplyMods(act, mods...)

	result, err := act.Run()
	if err != nil {
		return nil, err
	}

	releases := make([]Release, len(result))

	for id := range result {

		if r, err := fromHelmRelease(result[id]); err != nil {
			return nil, err
		} else {
			releases[id] = r
		}
	}

	return releases, nil
}

func (c *client) Invalidate() {
	c.discovery.Invalidate()
}

func (c *client) DiscoverKubernetesApiVersions(gv schema.GroupVersion) (ApiVersions, error) {
	resp, err := c.discovery.ServerResourcesForGroupVersion(gv.String())
	if err != nil {
		return nil, err
	}

	r := make(ApiVersions, len(resp.APIResources))

	for _, v := range resp.APIResources {
		v.Group = gv.Group
		v.Version = gv.Version
		r[schema.GroupVersionKind{
			Group:   v.Group,
			Version: v.Version,
			Kind:    v.Kind,
		}] = v
	}

	return r, nil
}

func (c *client) DiscoverKubernetesApiVersionKind(gvk schema.GroupVersionKind) (*meta.APIResource, error) {
	v, err := c.DiscoverKubernetesApiVersions(schema.GroupVersion{
		Group:   gvk.Group,
		Version: gvk.Version,
	})
	if err != nil {
		return nil, err
	}

	if z, ok := v[gvk]; ok {
		return &z, nil
	}

	return nil, nil
}

func (c *client) restClientForApiVersion(gv schema.GroupVersion) (rest.Interface, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if v, ok := c.restClients[gv]; ok {
		return v, nil
	}

	configShallowCopy := *c.cfg.Config
	if configShallowCopy.RateLimiter == nil && configShallowCopy.QPS > 0 {
		if configShallowCopy.Burst <= 0 {
			return nil, fmt.Errorf("burst is required to be greater than 0 when RateLimiter is not set and QPS is set to greater than 0")
		}
		configShallowCopy.RateLimiter = flowcontrol.NewTokenBucketRateLimiter(configShallowCopy.QPS, configShallowCopy.Burst)
	}

	configShallowCopy.GroupVersion = &schema.GroupVersion{
		Group:   gv.Group,
		Version: gv.Version,
	}
	configShallowCopy.APIPath = "/api"
	configShallowCopy.NegotiatedSerializer = scheme.Codecs.WithoutConversion()

	if configShallowCopy.UserAgent == "" {
		configShallowCopy.UserAgent = rest.DefaultKubernetesUserAgent()
	}

	rc, err := rest.RESTClientFor(&configShallowCopy)
	if err != nil {
		return nil, err
	}

	c.restClients[gv] = rc

	return rc, nil
}

func (c *client) NativeGet(ctx context.Context, reqs ...Resource) ([]ResourceObject, error) {
	if len(reqs) == 0 {
		return nil, nil
	}

	res := make([]ResourceObject, len(reqs))
	for id := range reqs {
		res[id].Resource = reqs[id]

		gvk, err := c.DiscoverKubernetesApiVersionKind(reqs[id].GroupVersionKind)
		if err != nil {
			return nil, err
		}

		if gvk == nil {
			// NotFound
			continue
		}

		client, err := c.restClientForApiVersion(schema.GroupVersion{
			Group:   gvk.Group,
			Version: gvk.Version,
		})
		if err != nil {
			return nil, err
		}

		resp, err := client.Get().Resource(gvk.Name).
			NamespaceIfScoped(reqs[id].Namespace, gvk.Namespaced).
			Name(reqs[id].Name).DoRaw(ctx)
		if err != nil {
			var e *apiErrors.StatusError
			if errors.As(err, &e) {
				if e.Status().Code == goHttp.StatusNotFound {
					continue
				}
			}

			return nil, err
		}

		res[id].Object = &ResourceObjectData{
			Data: resp,
		}
	}

	return res, nil
}

func (c *client) StatusObjects(ctx context.Context, name string, mods ...util.Mod[action.Status]) (*Release, []ResourceObject, error) {
	mods = append(mods, func(in *action.Status) {
		in.ShowResources = true
	})
	s, err := c.Status(ctx, name, mods...)
	if err != nil {
		return nil, nil, err
	}
	if s == nil {
		return nil, nil, nil
	}

	manifests, err := c.NativeGet(ctx, s.Info.Resources...)
	if err != nil {
		return s, nil, err
	}

	return s, manifests, nil
}
