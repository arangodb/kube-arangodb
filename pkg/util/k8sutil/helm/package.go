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
	"encoding/base64"
	goStrings "strings"

	"helm.sh/helm/v3/pkg/action"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	platformApi "github.com/arangodb/kube-arangodb/pkg/apis/platform/v1beta1"
	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/util"
	utilConstants "github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
)

type PackageType int

const (
	PackageTypePlatform PackageType = iota
	PackageTypeFile
	PackageTypeRemote
	PackageTypeOCI
	PackageTypeInline
	PackageTypeIndex
)

type Package struct {
	// Packages keeps the map of Packages to be installed
	Packages map[string]PackageSpec `json:"packages,omitempty"`

	// Releases keeps the map of Releases to be installed
	Releases map[string]PackageRelease `json:"releases,omitempty"`
}

func (pkg *Package) Validate() error {
	if pkg == nil {
		return nil
	}
	return errors.Errors(
		shared.PrefixResourceErrors("packages", shared.ValidateMap(pkg.Packages, func(s string, spec PackageSpec) error {
			return spec.Validate()
		})),
		shared.PrefixResourceErrors("releases", shared.ValidateMap(pkg.Releases, func(s string, spec PackageRelease) error {
			return spec.Validate()
		})),
	)
}

type PackageSpec struct {
	// Stage defines stage used in the fetch from LicenseManager
	Stage *string `json:"stage,omitempty"`

	// Version keeps the version of the PackageSpec
	Version string `json:"version"`

	// Chart defines override of the PackageSpec
	// It supports multiple modes:
	// - If undefined, LicenseManager OCI Repository is used
	// - If starts with `file://` chart is fetched from local FileSystem
	// - If starts with `http://` or `https://` chart is fetched from the remote URL
	// - If starts with `index://` chart is fetched using Helm YAML Index File stricture (using version and name)
	// - If Starts with `oci://` chart is fetched from Registry Compatible OCI Repository
	// - If none above match, chart is decoded using Base64 encoding
	Chart *string `json:"chart,omitempty"`

	// Overrides defines Values to override the Helm Chart Defaults (merged with Service Overrides)
	Overrides Values `json:"overrides,omitempty"`
}

func (p PackageSpec) PackageType() PackageType {
	if c := p.Chart; c != nil {
		// File
		if goStrings.HasPrefix(*c, "file://") {
			return PackageTypeFile
		}

		// HTTP
		if goStrings.HasPrefix(*c, "https://") || goStrings.HasPrefix(*c, "http://") {
			return PackageTypeRemote
		}

		// OCI
		if goStrings.HasPrefix(*c, "oci://") {
			return PackageTypeOCI
		}

		// Helm Index File
		if goStrings.HasPrefix(*c, "index://") {
			return PackageTypeIndex
		}

		return PackageTypeInline
	}
	return PackageTypePlatform
}

func (p PackageSpec) Validate() error {
	if c := p.Chart; c != nil {
		// File
		if goStrings.HasPrefix(*c, "file://") {
			return nil
		}

		// HTTP
		if goStrings.HasPrefix(*c, "https://") || goStrings.HasPrefix(*c, "http://") {
			return nil
		}

		// OCI
		if goStrings.HasPrefix(*c, "oci://") {
			return nil
		}

		// Helm Index File
		if goStrings.HasPrefix(*c, "index://") {
			return nil
		}

		// Base64
		if _, err := base64.StdEncoding.DecodeString(*c); err != nil {
			return errors.Wrapf(err, "Unable to decode chart data")
		}
	} else {
		if p.Version == "" {
			return errors.Errorf("Version is required if chart is not provided")
		}
	}

	return nil
}

func (p PackageSpec) GetStage() string {
	if p.Stage == nil {
		return "prd"
	}

	return *p.Stage
}

type PackageRelease struct {
	// Package keeps the name of the Chart used from the installation script.
	// References to value provided in Packages
	Package string `json:"package"`

	// Overrides defines Values to override the Helm Chart Defaults during installation
	Overrides Values `json:"overrides,omitempty"`
}

func (p PackageRelease) Validate() error {
	return nil
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
					Chart:     util.NewType(base64.StdEncoding.EncodeToString(c.Status.Info.Definition)),
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
