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
	"encoding/json"
	"time"

	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/release"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

func fromHelmRelease(in *release.Release) (Release, error) {
	var r Release

	if in == nil {
		return Release{}, errors.Errorf("Nil release not allowed")
	}

	r.Name = in.Name
	r.Version = in.Version
	r.Namespace = in.Namespace
	r.Labels = in.Labels

	r.Chart = fromHelmReleaseChart(in.Chart)

	if d, err := NewValues(in.Config); err != nil {
		return Release{}, nil
	} else {
		r.Values = d
	}

	if i, err := fromHelmReleaseInfo(in.Info); err != nil {
		return Release{}, err
	} else {
		r.Info = i
	}

	return r, nil
}

func fromHelmReleaseChart(in *chart.Chart) *ReleaseChart {
	if in == nil {
		return nil
	}

	var r ReleaseChart

	r.Metadata = fromHelmReleaseChartMetadata(in.Metadata)

	return &r
}

func fromHelmReleaseChartMetadata(in *chart.Metadata) *ReleaseChartMetadata {
	if in == nil {
		return nil
	}

	var r ReleaseChartMetadata

	r.Version = in.Version
	r.Name = in.Name

	return &r
}

func fromHelmReleaseInfo(in *release.Info) (ReleaseInfo, error) {
	var r ReleaseInfo

	if in == nil {
		return ReleaseInfo{}, errors.Errorf("Nil release info not allowed")
	}

	r.FirstDeployed = in.FirstDeployed.Time
	r.LastDeployed = in.LastDeployed.Time
	r.Deleted = in.Deleted.Time
	r.Description = in.Description
	r.Status = in.Status
	r.Notes = in.Notes

	if m, err := fromHelmReleaseInfoResources(in.Resources); err != nil {
		return ReleaseInfo{}, err
	} else {
		r.Resources = m
	}

	return r, nil
}

func fromHelmReleaseInfoResources(in map[string][]runtime.Object) (Resources, error) {
	if len(in) == 0 {
		return nil, nil
	}

	var r Resources

	for _, v := range in {
		for _, obj := range v {
			d, err := util.JSONRemarshal[runtime.Object, internalResourceObject](obj)
			if err != nil {
				return nil, err
			}

			r = append(r, Resource{
				GroupVersionKind: schema.FromAPIVersionAndKind(d.APIVersion, d.Kind),
				Name:             d.GetName(),
				Namespace:        d.GetNamespace(),
			})
		}
	}

	return r, nil
}

type Release struct {
	Name string `json:"name,omitempty"`

	Info  ReleaseInfo   `json:"info"`
	Chart *ReleaseChart `json:"chart,omitempty"`

	Values Values `json:"values,omitempty"`

	Version   int               `json:"version,omitempty"`
	Namespace string            `json:"namespace,omitempty"`
	Labels    map[string]string `json:"labels,omitempty"`
}

func (r *Release) GetChart() *ReleaseChart {
	if r == nil {
		return nil
	}

	return r.Chart
}

type ReleaseChart struct {
	Metadata *ReleaseChartMetadata `json:"metadata,omitempty"`
}

func (r *ReleaseChart) GetMetadata() *ReleaseChartMetadata {
	if r == nil {
		return nil
	}

	return r.Metadata
}

type ReleaseChartMetadata struct {
	Version string `json:"version,omitempty"`
	Name    string `json:"name,omitempty"`
}

func (r *ReleaseChartMetadata) GetName() string {
	if r == nil {
		return ""
	}

	return r.Name
}

func (r *ReleaseChartMetadata) GetVersion() string {
	if r == nil {
		return ""
	}

	return r.Version
}

type ReleaseInfo struct {
	FirstDeployed time.Time      `json:"first_deployed,omitempty"`
	LastDeployed  time.Time      `json:"last_deployed,omitempty"`
	Deleted       time.Time      `json:"deleted,omitempty"`
	Description   string         `json:"description,omitempty"`
	Status        release.Status `json:"status,omitempty"`
	Notes         string         `json:"notes,omitempty"`
	Resources     Resources      `json:"resources,omitempty"`
}

type Resources []Resource

type internalResourceObject struct {
	meta.TypeMeta   `json:",inline"`
	meta.ObjectMeta `json:"metadata,omitempty"`
}

type Resource struct {
	schema.GroupVersionKind
	Name, Namespace string
}

type ResourceObject struct {
	Resource

	Object *ResourceObjectData
}

func (r *ResourceObjectData) Unmarshal(obj any) error {
	if r == nil {
		return errors.Errorf("Object not returned")
	}

	return json.Unmarshal(r.Data, &obj)
}

type ResourceObjectData struct {
	Data []byte
}

type UninstallRelease struct {
	Release Release
	Info    string
}

type UpgradeResponse struct {
	Before, After *Release
}

type ApiVersions map[schema.GroupVersionKind]meta.APIResource
