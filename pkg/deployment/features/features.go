//
// DISCLAIMER
//
// Copyright 2016-2026 ArangoDB GmbH, Cologne, Germany
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

package features

import (
	"sort"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/version"
)

const (
	Enabled  = "true"
	Disabled = "false"
)

type Features []Feature

func (f Features) Get(name string) (Feature, bool) {
	for _, feature := range features {
		if feature.Name() == name {
			return feature, true
		}
	}

	return nil, false
}

func (f Features) Sort() {
	sort.Slice(f, func(i, j int) bool {
		return f[i].Name() < f[j].Name()
	})
}

var _ Feature = &feature{}

type Feature interface {
	Name() string
	Description() string
	Dependencies() []Feature
	Version() util.VersionConstrain
	EnterpriseRequired() bool
	OperatorEnterpriseRequired() bool
	EnabledByDefault() bool
	Enabled() bool
	EnabledPointer() *bool
	Deprecated() (bool, string)
	Hidden() bool
	Supported(v util.Version, enterprise bool) bool
	ImageSupported(i *api.ImageInfo) bool
	GetDependencies() []string
	Reset()
}

type feature struct {
	name, description                                                         string
	version                                                                   util.VersionConstrain
	enterpriseRequired, operatorEnterpriseRequired, enabledByDefault, enabled bool
	deprecated                                                                string
	constValue                                                                *bool
	hidden                                                                    bool
	dependencies                                                              []Feature
}

func (f feature) Dependencies() []Feature {
	if len(f.dependencies) == 0 {
		return nil
	}

	q := make([]Feature, len(f.dependencies))

	copy(q, f.dependencies)

	return q
}

func (f feature) ImageSupported(i *api.ImageInfo) bool {
	return i != nil && f.Supported(util.Version(i.ArangoDBVersion), i.Enterprise)
}

func (f feature) Hidden() bool {
	return f.hidden
}

// GetDependencies returns direct dependencies' names of features.
func (f feature) GetDependencies() []string {
	if len(f.dependencies) == 0 {
		return nil
	}

	deps := make([]string, 0, len(f.dependencies))
	for _, dependency := range f.dependencies {
		deps = append(deps, dependency.Name())
	}

	return deps
}

func (f feature) Supported(v util.Version, enterprise bool) bool {
	return Supported(&f, v, enterprise)
}

func (f feature) Enabled() bool {
	if f.operatorEnterpriseRequired {
		// Operator Enterprise is required for this feature
		if !version.GetVersionV1().IsEnterprise() {
			return false
		}
	}

	for _, dep := range f.dependencies {
		if !dep.Enabled() {
			return false
		}
	}

	if f.constValue != nil {
		return *f.constValue
	}

	if enableAll {
		return true
	}

	return f.enabled
}

func (f *feature) EnabledPointer() *bool {
	return &f.enabled
}

func (f feature) Version() util.VersionConstrain {
	return f.version
}

func (f feature) EnterpriseRequired() bool {
	return f.enterpriseRequired
}

func (f feature) OperatorEnterpriseRequired() bool {
	return f.operatorEnterpriseRequired
}

func (f feature) EnabledByDefault() bool {
	return f.enabledByDefault
}

func (f feature) Name() string {
	return f.name
}

func (f feature) Description() string {
	return f.description
}

// Deprecated returns true if the feature is deprecated and the reason why it is deprecated.
func (f feature) Deprecated() (bool, string) {
	return len(f.deprecated) > 0, f.deprecated
}

func (f *feature) Reset() {
	f.enabled = f.enabledByDefault
}
