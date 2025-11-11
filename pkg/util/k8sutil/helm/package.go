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

package helm

import (
	"context"

	"helm.sh/helm/v3/pkg/action"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	platformApi "github.com/arangodb/kube-arangodb/pkg/apis/platform/v1beta1"
	sharedApi "github.com/arangodb/kube-arangodb/pkg/apis/shared/v1"
	utilConstants "github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
)

type Package struct {
	Packages map[string]PackageSpec `json:"packages,omitempty"`

	Releases map[string]PackageRelease `json:"releases,omitempty"`
}

func (pkg *Package) Validate() error {
	return nil
}

type PackageSpec struct {
	Stage *string `json:"stage,omitempty"`

	Version string `json:"version"`

	Chart sharedApi.Data `json:"chart,omitempty"`

	Overrides Values `json:"overrides,omitempty"`
}

func (p PackageSpec) GetStage() string {
	if p.Stage == nil {
		return "prd"
	}

	return *p.Stage
}

type PackageRelease struct {
	Package string `json:"package"`

	Overrides Values `json:"overrides,omitempty"`
}

func NewPackage(ctx context.Context, client kclient.Client, namespace, deployment string) (*Package, error) {
	hclient, err := NewClient(Configuration{
		Namespace: namespace,
		Config:    client.Config(),
		Driver:    nil,
	})
	if err != nil {
		return nil, err
	}

	charts, err := GetLocalCharts(ctx, client, namespace)
	if err != nil {
		return nil, err
	}

	var out Package

	out.Packages = map[string]PackageSpec{}

	out.Releases = map[string]PackageRelease{}

	for name, c := range charts {
		if !c.Status.Conditions.IsTrue(platformApi.ReadyCondition) {
			return nil, errors.Errorf("Chart `%s` is not in ready condition", name)
		}

		if info := c.Status.Info; info != nil {
			if det := info.Details; det != nil {
				out.Packages[name] = PackageSpec{
					Version:   det.GetVersion(),
					Overrides: Values(info.Overrides),
					Chart:     c.Status.Info.Definition,
				}
			}
		}

		existingReleases, err := hclient.List(ctx, func(in *action.List) {
			in.Selector = meta.FormatLabelSelector(&meta.LabelSelector{
				MatchLabels: map[string]string{
					utilConstants.HelmLabelArangoDBManaged:    "true",
					utilConstants.HelmLabelArangoDBDeployment: deployment,
					utilConstants.HelmLabelArangoDBChart:      name,
					utilConstants.HelmLabelArangoDBType:       "platform",
				},
			})
		})
		if err != nil {
			logger.Err(err).Error("Unable to list releases")
			return nil, err
		}

		for _, release := range existingReleases {
			var r PackageRelease

			r.Package = name

			data, err := release.Values.Marshal()
			if err != nil {
				logger.Err(err).Error("Unable to unmarshal values")
				return nil, err
			}

			delete(data, "arangodb_platform")

			if len(data) != 0 {
				values, err := NewValues(data)
				if err != nil {
					logger.Err(err).Error("Unable to marshal values")
					return nil, err
				}

				r.Overrides = values
			}

			out.Releases[release.Name] = r
		}
	}

	return &out, nil
}
