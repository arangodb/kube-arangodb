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
	"os"
	"path"

	"gopkg.in/yaml.v3"

	"github.com/arangodb/kube-arangodb/internal/md"
	"github.com/arangodb/kube-arangodb/pkg/util"
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

func GenerateReadme(root string) error {
	readmeSections := map[string]string{}

	{
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
			return err
		}

		if err := yaml.Unmarshal(data, &d); err != nil {
			return err
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
					return err
				}
			}
		}

		readmeSections["metricsTable"] = md.WrapWithNewLines(t.Render())
	}

	if err := md.ReplaceSectionsInFile(path.Join(root, "README.md"), readmeSections); err != nil {
		return err
	}

	return nil
}
