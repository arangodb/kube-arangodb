//
// DISCLAIMER
//
// Copyright 2023 ArangoDB GmbH, Cologne, Germany
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

package internal

import (
	"fmt"
	"os"
	"path"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/arangodb/go-driver"

	"github.com/arangodb/kube-arangodb/internal/md"
	"github.com/arangodb/kube-arangodb/pkg/util"
)

const minSupportedArangoDBVersion = ">= 3.8.0"

type PlatformsDoc struct {
	Platforms Platforms `json:"platforms,omitempty" yaml:"platforms,omitempty"`
}

type Platforms []Platform

type Platform struct {
	Name     string            `json:"name,omitempty" yaml:"name,omitempty"`
	Versions []PlatformVersion `json:"versions,omitempty" yaml:"versions,omitempty"`
}

type PlatformVersion struct {
	KubernetesVersion *string `json:"kubernetesVersion,omitempty" yaml:"kubernetesVersion,omitempty"`
	ArangoDBVersion   *string `json:"arangoDBVersion,omitempty" yaml:"arangoDBVersion,omitempty"`
	State             *string `json:"state,omitempty" yaml:"state,omitempty"`
	Remarks           *string `json:"remarks,omitempty" yaml:"remarks,omitempty"`
	ProviderRemarks   *string `json:"providerRemarks,omitempty" yaml:"providerRemarks,omitempty"`
}

type FeaturesDoc struct {
	Features Features `json:"features,omitempty" yaml:"features,omitempty"`
}

type Features []Feature

type Feature struct {
	Name     string          `json:"name,omitempty" yaml:"name,omitempty"`
	Releases FeatureReleases `json:"releases,omitempty" yaml:"releases,omitempty"`

	FeatureRelease `json:",inline" yaml:",inline"`
}

type FeatureReleases []FeatureRelease

type FeatureRelease struct {
	Doc *string `json:"doc,omitempty" yaml:"doc,omitempty"`

	OperatorVersion *string `json:"operatorVersion,omitempty" yaml:"operatorVersion,omitempty"`
	ArangoDBVersion *string `json:"arangoDBVersion,omitempty" yaml:"arangoDBVersion,omitempty"`

	OperatorEdition *string `json:"operatorEditions,omitempty" yaml:"operatorEditions,omitempty"`
	ArangoDBEdition *string `json:"arangoDBEditions,omitempty" yaml:"arangoDBEditions,omitempty"`

	State   *string `json:"state,omitempty" yaml:"state,omitempty"`
	Flag    *string `json:"flag,omitempty" yaml:"flag,omitempty"`
	Remarks *string `json:"remarks,omitempty" yaml:"remarks,omitempty"`

	Enabled *bool `json:"enabled,omitempty" yaml:"enabled,omitempty"`
}

type LimitsDoc struct {
	Limits Limits `json:"limits,omitempty" yaml:"limits,omitempty"`
}

type Limits []Limit

type Limit struct {
	Name        string  `json:"name" yaml:"name"`
	Description string  `json:"description" yaml:"description"`
	Community   *string `json:"community,omitempty" yaml:"community,omitempty"`
	Enterprise  *string `json:"enterprise,omitempty" yaml:"enterprise,omitempty"`
}

func GenerateReadme(root string) error {
	readmeSections := map[string]string{}

	if section, err := GenerateReadmePlatforms(root); err != nil {
		return err
	} else {
		readmeSections["kubernetesVersionsTable"] = section
	}

	if section, err := GenerateReadmeFeatures(root, true); err != nil {
		return err
	} else {
		readmeSections["featuresEnterpriseTable"] = section
	}

	if section, err := GenerateReadmeFeatures(root, false); err != nil {
		return err
	} else {
		readmeSections["featuresCommunityTable"] = section
	}

	if section, err := GenerateReadmeLimits(root); err != nil {
		return err
	} else {
		readmeSections["limits"] = section
	}

	if err := md.ReplaceSectionsInFile(path.Join(root, "README.md"), readmeSections); err != nil {
		return err
	}

	return nil
}

func GenerateReadmeFeatures(root string, eeOnly bool) (string, error) {
	feature := md.NewColumn("Feature", md.ColumnLeftAlign)
	introduced := md.NewColumn("Introduced", md.ColumnLeftAlign)
	oVersion := md.NewColumn("Operator Version", md.ColumnLeftAlign)
	aVersion := md.NewColumn("ArangoDB Version", md.ColumnLeftAlign)
	aEdition := md.NewColumn("ArangoDB Edition", md.ColumnLeftAlign)
	state := md.NewColumn("State", md.ColumnLeftAlign)
	enabled := md.NewColumn("Enabled", md.ColumnLeftAlign)
	flag := md.NewColumn("Flag", md.ColumnLeftAlign)
	remarks := md.NewColumn("Remarks", md.ColumnLeftAlign)
	t := md.NewTable(
		feature,
		oVersion,
		introduced,
		aVersion,
		aEdition,
		state,
		enabled,
		flag,
		remarks,
	)

	var d FeaturesDoc

	data, err := os.ReadFile(path.Join(root, "internal", "features.yaml"))
	if err != nil {
		return "", err
	}

	if err := yaml.Unmarshal(data, &d); err != nil {
		return "", err
	}

	// Sort list

	sort.Slice(d.Features, func(i, j int) bool {
		{
			av := util.First(util.LastFromList(d.Features[i].Releases).OperatorVersion, d.Features[i].OperatorVersion)
			bv := util.First(util.LastFromList(d.Features[j].Releases).OperatorVersion, d.Features[j].OperatorVersion)

			a := driver.Version(util.TypeOrDefault[string](av, "1.0.0"))
			b := driver.Version(util.TypeOrDefault[string](bv, "1.0.0"))

			if c := a.CompareTo(b); c != 0 {
				return c > 0
			}
		}

		{
			a := driver.Version(util.TypeOrDefault[string](d.Features[i].Releases[0].OperatorVersion, "1.0.0"))
			b := driver.Version(util.TypeOrDefault[string](d.Features[j].Releases[0].OperatorVersion, "1.0.0"))

			if c := a.CompareTo(b); c != 0 {
				return c > 0
			}
		}

		return d.Features[i].Name < d.Features[j].Name
	})

	for _, f := range d.Features {
		r := f.Releases[len(f.Releases)-1]

		if community := strings.Contains(util.TypeOrDefault(util.First(r.OperatorEdition, f.OperatorEdition), "Community, Enterprise"), "Community"); community == eeOnly {
			continue
		}

		n := f.Name

		if v := util.First(r.Doc, f.Doc); v != nil {
			n = fmt.Sprintf("[%s](%s)", n, *v)
		}

		if err := t.AddRow(map[md.Column]string{
			feature:    n,
			oVersion:   util.TypeOrDefault[string](util.First(r.OperatorVersion, f.OperatorVersion), "ANY"),
			introduced: util.TypeOrDefault[string](f.Releases[0].OperatorVersion, "ANY"),
			aVersion:   util.TypeOrDefault[string](util.First(r.ArangoDBVersion, f.ArangoDBVersion), minSupportedArangoDBVersion),
			aEdition:   util.TypeOrDefault[string](util.First(r.ArangoDBEdition, f.ArangoDBEdition), "Community, Enterprise"),
			aEdition:   util.TypeOrDefault[string](util.First(r.ArangoDBEdition, f.ArangoDBEdition), "Community, Enterprise"),
			state:      util.TypeOrDefault[string](util.First(r.State, f.State), "Alpha"),
			enabled:    util.BoolSwitch[string](util.TypeOrDefault[bool](util.First(r.Enabled, f.Enabled), true), "True", "False"),
			flag:       util.TypeOrDefault[string](util.First(r.Flag, f.Flag), "N/A"),
			remarks:    util.TypeOrDefault[string](util.First(r.Remarks, f.Remarks), "N/A"),
		}); err != nil {
			return "", err
		}
	}

	return md.WrapWithNewLines(t.Render()), nil
}

func GenerateReadmeLimits(root string) (string, error) {
	limit := md.NewColumn("Limit", md.ColumnLeftAlign)
	description := md.NewColumn("Description", md.ColumnLeftAlign)
	community := md.NewColumn("Community", md.ColumnLeftAlign)
	enterprise := md.NewColumn("Enterprise", md.ColumnLeftAlign)
	t := md.NewTable(
		limit,
		description,
		community,
		enterprise,
	)

	var d LimitsDoc

	data, err := os.ReadFile(path.Join(root, "internal", "limits.yaml"))
	if err != nil {
		return "", err
	}

	if err := yaml.Unmarshal(data, &d); err != nil {
		return "", err
	}

	for _, l := range d.Limits {
		if err := t.AddRow(map[md.Column]string{
			limit:       l.Name,
			description: l.Description,
			community:   util.TypeOrDefault[string](l.Community, "N/A"),
			enterprise:  util.TypeOrDefault[string](l.Enterprise, "N/A"),
		}); err != nil {
			return "", err
		}
	}

	return md.WrapWithNewLines(t.Render()), nil
}

func GenerateReadmePlatforms(root string) (string, error) {
	platform := md.NewColumn("Platform", md.ColumnLeftAlign)
	kVersion := md.NewColumn("Kubernetes Version", md.ColumnLeftAlign)
	aVersion := md.NewColumn("ArangoDB Version", md.ColumnLeftAlign)
	state := md.NewColumn("State", md.ColumnLeftAlign)
	remarks := md.NewColumn("Remarks", md.ColumnLeftAlign)
	pRemarks := md.NewColumn("Provider Remarks", md.ColumnLeftAlign)
	t := md.NewTable(
		platform,
		kVersion,
		aVersion,
		state,
		remarks,
		pRemarks,
	)

	var d PlatformsDoc

	data, err := os.ReadFile(path.Join(root, "internal", "platforms.yaml"))
	if err != nil {
		return "", err
	}

	if err := yaml.Unmarshal(data, &d); err != nil {
		return "", err
	}

	for _, p := range d.Platforms {
		for _, v := range p.Versions {
			if err := t.AddRow(map[md.Column]string{
				platform: p.Name,
				kVersion: util.TypeOrDefault[string](v.KubernetesVersion, ""),
				aVersion: util.TypeOrDefault[string](v.ArangoDBVersion, ""),
				state:    util.TypeOrDefault[string](v.State, ""),
				remarks:  util.TypeOrDefault[string](v.Remarks, ""),
				pRemarks: util.TypeOrDefault[string](v.ProviderRemarks, ""),
			}); err != nil {
				return "", err
			}
		}
	}

	return md.WrapWithNewLines(t.Render()), nil
}
