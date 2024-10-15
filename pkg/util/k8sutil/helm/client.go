//
// DISCLAIMER
//
// Copyright 2024 ArangoDB GmbH, Cologne, Germany
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

	"helm.sh/helm/v3/pkg/action"

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
	"github.com/arangodb/kube-arangodb/pkg/util/tests"
)

func NewClient(cfg Configuration) (Client, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	var helm action.Configuration

	if err := helm.Init(kclient.NewRESTClientGetter(tests.FakeNamespace, nil, cfg.Client.Config()), cfg.Namespace, "configmap", func(format string, v ...interface{}) {
		logger.Debug(format, v...)
	}); err != nil {
		return nil, err
	}

	return &client{
		cfg:  cfg,
		helm: &helm,
	}, nil
}

type Client interface {
	Namespace() string
	Client() kclient.Client

	Alive(ctx context.Context) error

	Status(ctx context.Context, name string, mods ...util.Mod[action.Status]) (*Release, error)
	List(ctx context.Context, mods ...util.Mod[action.List]) ([]Release, error)
	Install(ctx context.Context, chart Chart, values Values, mods ...util.Mod[action.Install]) (*Release, error)
	Upgrade(ctx context.Context, name string, chart Chart, values Values, mods ...util.Mod[action.Upgrade]) (*Release, error)
	Uninstall(ctx context.Context, name string, mods ...util.Mod[action.Uninstall]) (*UninstallRelease, error)
	Test(ctx context.Context, name string, mods ...util.Mod[action.ReleaseTesting]) (*Release, error)
}

type client struct {
	cfg  Configuration
	helm *action.Configuration
}

func (c client) Namespace() string {
	return c.cfg.Namespace
}

func (c client) Client() kclient.Client {
	return c.cfg.Client
}

func (c client) Status(ctx context.Context, name string, mods ...util.Mod[action.Status]) (*Release, error) {
	act := action.NewStatus(c.helm)

	util.ApplyMods(act, mods...)

	result, err := act.Run(name)
	if err != nil {
		if err.Error() == "release: not found" {
			return nil, nil
		}
		return nil, err
	}

	r := fromHelmRelease(result)

	return &r, nil
}

func (c client) Uninstall(ctx context.Context, name string, mods ...util.Mod[action.Uninstall]) (*UninstallRelease, error) {
	act := action.NewUninstall(c.helm)

	util.ApplyMods(act, mods...)

	result, err := act.Run(name)
	if err != nil {
		return nil, err
	}

	var res UninstallRelease

	res.Info = result.Info
	res.Release = fromHelmRelease(result.Release)

	return &res, nil
}

func (c client) Test(ctx context.Context, name string, mods ...util.Mod[action.ReleaseTesting]) (*Release, error) {
	act := action.NewReleaseTesting(c.helm)

	util.ApplyMods(act, mods...)

	result, err := act.Run(name)
	if err != nil {
		return nil, err
	}

	r := fromHelmRelease(result)

	return &r, nil
}

func (c client) Install(ctx context.Context, chart Chart, values Values, mods ...util.Mod[action.Install]) (*Release, error) {
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

	result, err := act.Run(chartData, valuesData)
	if err != nil {
		return nil, err
	}

	r := fromHelmRelease(result)

	return &r, nil
}

func (c client) Upgrade(ctx context.Context, name string, chart Chart, values Values, mods ...util.Mod[action.Upgrade]) (*Release, error) {
	act := action.NewUpgrade(c.helm)

	util.ApplyMods(act, mods...)

	chartData, err := chart.Get()
	if err != nil {
		return nil, err
	}

	valuesData, err := values.Marshal()
	if err != nil {
		return nil, err
	}

	result, err := act.Run(name, chartData, valuesData)
	if err != nil {
		return nil, err
	}

	r := fromHelmRelease(result)

	return &r, nil
}

func (c client) Alive(ctx context.Context) error {
	act := action.NewList(c.helm)

	_, err := act.Run()
	if err != nil {
		return err
	}

	return nil
}

func (c client) List(ctx context.Context, mods ...util.Mod[action.List]) ([]Release, error) {
	act := action.NewList(c.helm)

	util.ApplyMods(act, mods...)

	result, err := act.Run()
	if err != nil {
		return nil, err
	}

	releases := make([]Release, len(result))

	for id := range result {
		releases[id] = fromHelmRelease(result[id])
	}

	return releases, nil
}
