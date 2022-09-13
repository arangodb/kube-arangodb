//
// DISCLAIMER
//
// Copyright 2016-2022 ArangoDB GmbH, Cologne, Germany
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

import "github.com/arangodb/go-driver"

const Enabled = "true"

var _ Feature = &feature{}

type Feature interface {
	Name() string
	Description() string
	Version() driver.Version
	EnterpriseRequired() bool
	EnabledByDefault() bool
	Enabled() bool
	EnabledPointer() *bool
	Deprecated() (bool, string)
	Hidden() bool
	Supported(v driver.Version, enterprise bool) bool
}

type feature struct {
	name, description                             string
	version                                       driver.Version
	enterpriseRequired, enabledByDefault, enabled bool
	deprecated                                    string
	constValue                                    *bool
	hidden                                        bool
}

func (f feature) Hidden() bool {
	return f.hidden
}

func (f feature) Supported(v driver.Version, enterprise bool) bool {
	return Supported(&f, v, enterprise)
}

func (f feature) Enabled() bool {
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

func (f feature) Version() driver.Version {
	return f.version
}

func (f feature) EnterpriseRequired() bool {
	return f.enterpriseRequired
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
