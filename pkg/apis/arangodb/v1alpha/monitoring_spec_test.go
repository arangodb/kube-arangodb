//
// DISCLAIMER
//
// Copyright 2018 ArangoDB GmbH, Cologne, Germany
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
// Author Ewout Prangsma
//

package v1alpha

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMonitoringSpecValidate(t *testing.T) {
	// Valid
	assert.Nil(t, MonitoringSpec{TokenSecretName: ""}.Validate())
	assert.Nil(t, MonitoringSpec{TokenSecretName: "foo"}.Validate())
	assert.Nil(t, MonitoringSpec{TokenSecretName: "foo"}.Validate())

	// Not valid
	assert.Error(t, MonitoringSpec{TokenSecretName: "Foo"}.Validate())
}

func TestMonitoringSpecSetDefaults(t *testing.T) {
	def := func(spec MonitoringSpec) MonitoringSpec {
		spec.SetDefaults()
		return spec
	}

	assert.Equal(t, "", def(MonitoringSpec{}).TokenSecretName)
	assert.Equal(t, "foo", def(MonitoringSpec{TokenSecretName: "foo"}).TokenSecretName)
}
