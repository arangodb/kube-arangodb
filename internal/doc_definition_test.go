//
// DISCLAIMER
//
// Copyright 2023-2025 ArangoDB GmbH, Cologne, Germany
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
	"sort"
	goStrings "strings"

	stringslices "k8s.io/utils/strings/slices"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/strings"
)

type DocDefinitions []DocDefinition

type DocDefinition struct {
	Path string
	Type string

	File string
	Line int

	Include bool

	Docs []string

	Grade *DocDefinitionGradeDefinition

	Links []string

	Important *string

	Enum []string

	Immutable *string

	Skip []string

	Required *string

	Default *string
	Example []string
}

func (d DocDefinitions) Sort() {
	sort.Slice(d, func(i, j int) bool {
		a, b := goStrings.ToLower(d[i].Path), goStrings.ToLower(d[j].Path)
		if a == b {
			return d[i].Path < d[j].Path
		}
		return a < b
	})
}

func (d DocDefinitions) Find(path string) []DocDefinition {
	var r []DocDefinition

	for _, d := range d {
		if d.Path == path {
			r = append(r, d)
		}
	}

	return r
}

func (d DocDefinitions) Skipped(path string) bool {
	if path == "." {
		return false
	}

	parts := strings.Split(path, ".")
	if len(parts) == 1 {
		return false
	}

	for _, current := range d.Find(path) {
		if stringslices.Contains(current.Skip, "") {
			return true
		}
	}

	pPath := strings.Join(parts[0:len(parts)-1], ".")
	name := parts[len(parts)-1]

	for _, parent := range d.Find(pPath) {
		if stringslices.Contains(parent.Skip, name) {
			return true
		}
	}

	return d.Skipped(pPath)
}

func NewDocDefinitionGradeDefinition(lines ...string) (*DocDefinitionGradeDefinition, error) {
	if len(lines) == 0 {
		return nil, nil
	}

	start := lines[0]

	var ret DocDefinitionGradeDefinition

	grade, err := DocDefinitionGradeFromString(start)
	if err != nil {
		return nil, err
	}

	ret.Grade = grade

	if len(lines) > 1 {
		ret.Message = lines[1:]
	}

	return &ret, nil
}

type DocDefinitionGradeDefinition struct {
	Grade   DocDefinitionGrade
	Message []string
}

type DocDefinitionGrade int

const (
	DocDefinitionGradeAlpha DocDefinitionGrade = iota
	DocDefinitionGradeBeta
	DocDefinitionGradeProduction
	DocDefinitionGradeDeprecating
	DocDefinitionGradeDeprecated
)

func DocDefinitionGradeFromString(in string) (DocDefinitionGrade, error) {
	switch strings.ToLower(in) {
	case "alpha":
		return DocDefinitionGradeAlpha, nil
	case "beta":
		return DocDefinitionGradeBeta, nil
	case "production":
		return DocDefinitionGradeProduction, nil
	case "deprecating":
		return DocDefinitionGradeDeprecating, nil
	case "deprecated":
		return DocDefinitionGradeAlpha, nil
	default:
		return 0, errors.Errorf("Unable to parse grade: %s", in)
	}
}
