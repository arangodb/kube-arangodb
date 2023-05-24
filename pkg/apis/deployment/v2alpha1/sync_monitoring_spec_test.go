//
// DISCLAIMER
//
// Copyright 2016-2023 ArangoDB GmbH, Cologne, Germany
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

package v2alpha1

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/arangodb/kube-arangodb/pkg/util"
)

func TestMonitoringSpecValidate(t *testing.T) {
	// Valid
	assert.Nil(t, MonitoringSpec{TokenSecretName: nil}.Validate())
	assert.Nil(t, MonitoringSpec{TokenSecretName: util.NewType[string]("")}.Validate())
	assert.Nil(t, MonitoringSpec{TokenSecretName: util.NewType[string]("foo")}.Validate())
	assert.Nil(t, MonitoringSpec{TokenSecretName: util.NewType[string]("foo")}.Validate())

	// Not valid
	assert.Error(t, MonitoringSpec{TokenSecretName: util.NewType[string]("Foo")}.Validate())
}

func TestMonitoringSpecSetDefaults(t *testing.T) {
	def := func(spec MonitoringSpec) MonitoringSpec {
		spec.SetDefaults("")
		return spec
	}
	def2 := func(spec MonitoringSpec) MonitoringSpec {
		spec.SetDefaults("def2")
		return spec
	}

	assert.Equal(t, "", def(MonitoringSpec{}).GetTokenSecretName())
	assert.Equal(t, "def2", def2(MonitoringSpec{}).GetTokenSecretName())
	assert.Equal(t, "foo", def(MonitoringSpec{TokenSecretName: util.NewType[string]("foo")}).GetTokenSecretName())
}
