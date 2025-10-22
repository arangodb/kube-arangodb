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

package pack

import (
	"context"
	"sync"

	"github.com/arangodb/kube-arangodb/pkg/logging"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/executor"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/helm"
)

func Registry(ctx context.Context, registry string, m helm.ChartManager, p helm.Package) (helm.Package, error) {
	i := &registryExport{
		registry: registry,
		m:        m,
	}

	if err := executor.Run(ctx, logger, 8, i.run(p)); err != nil {
		return helm.Package{}, err
	}

	return i.p, nil
}

type registryExport struct {
	lock sync.Mutex

	registry string

	p helm.Package

	m helm.ChartManager
}

func (i *registryExport) run(p helm.Package) executor.RunFunc {
	return func(ctx context.Context, log logging.Logger, t executor.Thread, h executor.Handler) error {
		for k, v := range p.Packages {
			h.RunAsync(ctx, i.exportPackage(k, v))
		}

		return nil
	}
}

func (i *registryExport) exportPackage(name string, spec helm.PackageSpec) executor.RunFunc {
	return func(ctx context.Context, log logging.Logger, t executor.Thread, h executor.Handler) error {
		var pkgS helm.PackageSpec

		repo, ok := i.m.Get(name)
		if !ok {
			return errors.Errorf("Chart `%s` not found", name)
		}

		ver, ok := repo.Get(spec.Version)
		if !ok {
			return errors.Errorf("Chart `%s=%s` not found", name, spec.Version)
		}

		c, err := ver.Get(ctx)
		if err != nil {
			return err
		}

		pkgS.Version = spec.Version

		chartData, err := c.Get()
		if err != nil {
			return err
		}

		type valuesInterface struct {
			Images ProtoImages `json:"images,omitempty"`
		}

		protoImages, err := util.JSONRemarshal[map[string]any, valuesInterface](chartData.Chart().Values)
		if err != nil {
			return err
		}

		var versions ProtoValues

		versions.Images = map[string]ProtoImage{}

		for k, v := range protoImages.Images {
			if v.IsTest() {
				logger.Str("image", v.GetImage()).Info("Skip Test Image")
				continue
			}

			versions.Images[k] = ProtoImage{
				Registry: util.NewType(i.registry),
			}
		}

		vData, err := helm.NewValues(versions)
		if err != nil {
			return err
		}

		pkgS.Overrides = vData

		i.withPackage(func(in helm.Package) helm.Package {
			if in.Packages == nil {
				in.Packages = map[string]helm.PackageSpec{}
			}

			in.Packages[name] = pkgS

			return in
		})

		return nil
	}
}

func (i *registryExport) withPackage(mod util.ModR[helm.Package]) {
	i.lock.Lock()
	defer i.lock.Unlock()

	i.p = mod(i.p)
}
