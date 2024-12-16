//
// DISCLAIMER
//
// Copyright 2023-2024 ArangoDB GmbH, Cologne, Germany
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
	"bytes"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/arangodb/go-driver"

	"github.com/arangodb/kube-arangodb/cmd"
	"github.com/arangodb/kube-arangodb/pkg/deployment/features"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/pretty"
)

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

	const basePath = ""
	if section, err := GenerateReadmeFeatures(root, basePath, true); err != nil {
		return err
	} else {
		readmeSections["featuresEnterpriseTable"] = section
	}

	if section, err := GenerateReadmeFeatures(root, basePath, false); err != nil {
		return err
	} else {
		readmeSections["featuresCommunityTable"] = section
	}

	if section, err := GenerateReadmeLimits(root); err != nil {
		return err
	} else {
		readmeSections["limits"] = section
	}

	if section, err := GenerateHelp(cmd.Command()); err != nil {
		return err
	} else {
		readmeSections["operatorArguments"] = section
	}

	if err := pretty.ReplaceSectionsInFile(path.Join(root, "README.md"), readmeSections); err != nil {
		return err
	}

	return nil
}

func GenerateHelp(cmd *cobra.Command, args ...string) (string, error) {
	var lines []string

	lines = append(lines, "```", "Flags:")

	help, err := GenerateHelpRaw(cmd, args...)
	if err != nil {
		return "", err
	}

	for _, line := range strings.Split(help, "\n") {
		if strings.HasPrefix(line, "      --") {
			lines = append(lines, line)
		}
	}

	lines = append(lines, "```")

	return pretty.WrapWithNewLines(pretty.WrapWithNewLines(strings.Join(lines, "\n"))), nil
}

func GenerateHelpQuoted(cmd *cobra.Command, args ...string) (string, error) {
	h, err := GenerateHelpRaw(cmd, args...)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("```\n%s```\n", h), nil
}

func GenerateHelpRaw(cmd *cobra.Command, args ...string) (string, error) {
	buff := bytes.NewBuffer(nil)

	cmd.SetOut(buff)

	cmd.SetArgs(append(args, "--help"))

	if err := cmd.Execute(); err != nil {
		return "", err
	}

	return buff.String(), nil
}

func GenerateReadmeFeatures(root, basePath string, eeOnly bool) (string, error) {
	type tableRow struct {
		Feature         string `table:"Feature" table_align:"left"`
		OperatorVersion string `table:"Operator Version" table_align:"left"`
		Introduced      string `table:"Introduced" table_align:"left"`
		ArangoDBVersion string `table:"ArangoDB Version" table_align:"left"`
		ArangoDBEdition string `table:"ArangoDB Edition" table_align:"left"`
		State           string `table:"State" table_align:"left"`
		Enabled         string `table:"Enabled" table_align:"left"`
		Flag            string `table:"Flag" table_align:"left"`
		Remarks         string `table:"Remarks" table_align:"left"`
	}

	tb, err := pretty.NewTable[tableRow]()
	if err != nil {
		return "", err
	}

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
			p, err := filepath.Rel(basePath, *v)
			if err != nil {
				return "", err
			}

			n = fmt.Sprintf("[%s](%s)", n, p)
		}

		tb.Add(tableRow{
			Feature:         n,
			Introduced:      util.TypeOrDefault[string](f.Releases[0].OperatorVersion, "ANY"),
			OperatorVersion: util.TypeOrDefault[string](util.First(r.OperatorVersion, f.OperatorVersion)),
			ArangoDBVersion: util.TypeOrDefault[string](util.First(r.ArangoDBVersion, f.ArangoDBVersion), fmt.Sprintf(">= %s", features.MinSupportedArangoDBVersion)),
			ArangoDBEdition: util.TypeOrDefault[string](util.First(r.ArangoDBEdition, f.ArangoDBEdition), "Community, Enterprise"),
			State:           util.TypeOrDefault[string](util.First(r.State, f.State), "Alpha"),
			Enabled:         util.BoolSwitch[string](util.TypeOrDefault[bool](util.First(r.Enabled, f.Enabled), true), "True", "False"),
			Flag:            util.TypeOrDefault[string](util.First(r.Flag, f.Flag), "N/A"),
			Remarks:         util.TypeOrDefault[string](util.First(r.Remarks, f.Remarks), "N/A"),
		})
	}

	return pretty.WrapWithNewLines(tb.RenderMarkdown()), nil
}

func GenerateReadmeLimits(root string) (string, error) {
	type tableRow struct {
		Limit       string `table:"Limit" table_align:"left"`
		Description string `table:"Description" table_align:"left"`
		Community   string `table:"Community" table_align:"left"`
		Enterprise  string `table:"Enterprise" table_align:"left"`
	}
	tb, err := pretty.NewTable[tableRow]()
	if err != nil {
		return "", err
	}

	var d LimitsDoc

	data, err := os.ReadFile(path.Join(root, "internal", "limits.yaml"))
	if err != nil {
		return "", err
	}

	if err := yaml.Unmarshal(data, &d); err != nil {
		return "", err
	}

	for _, l := range d.Limits {
		tb.Add(tableRow{
			Limit:       l.Name,
			Description: l.Description,
			Community:   util.TypeOrDefault[string](l.Community, "N/A"),
			Enterprise:  util.TypeOrDefault[string](l.Enterprise, "N/A"),
		})
	}

	return pretty.WrapWithNewLines(tb.RenderMarkdown()), nil
}

func GenerateReadmePlatforms(root string) (string, error) {
	type tableRow struct {
		Platform          string `table:"Platform" table_align:"left"`
		State             string `table:"State" table_align:"left"`
		KubernetesVersion string `table:"Kubernetes Version" table_align:"left"`
		ArangoDBVersion   string `table:"ArangoDB Version" table_align:"left"`
		Remarks           string `table:"Remarks" table_align:"left"`
		ProviderRemarks   string `table:"Provider Remarks" table_align:"left"`
	}
	tb, err := pretty.NewTable[tableRow]()
	if err != nil {
		return "", err
	}

	var d PlatformsDoc

	data, err := os.ReadFile(path.Join(root, "internal", "platforms.yaml"))
	if err != nil {
		return "", err
	}

	if err := yaml.Unmarshal(data, &d); err != nil {
		return "", err
	}

	for _, p := range d.Platforms {
		for id, v := range p.Versions {
			n := ""
			if id == 0 {
				n = p.Name
			}
			tb.Add(tableRow{
				Platform:          n,
				KubernetesVersion: util.TypeOrDefault[string](v.KubernetesVersion, ""),
				ArangoDBVersion:   util.TypeOrDefault[string](v.ArangoDBVersion, ""),
				State:             util.TypeOrDefault[string](v.State, ""),
				Remarks:           util.TypeOrDefault[string](v.Remarks, ""),
				ProviderRemarks:   util.TypeOrDefault[string](v.ProviderRemarks, ""),
			})
		}
	}

	return pretty.WrapWithNewLines(tb.RenderMarkdown()), nil
}
