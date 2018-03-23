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

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/stretchr/testify/assert"
	"k8s.io/api/core/v1"
)

func TestDeploymentSpecValidate(t *testing.T) {
	// TODO
}

func TestDeploymentSpecSetDefaults(t *testing.T) {
	def := func(spec DeploymentSpec) DeploymentSpec {
		spec.SetDefaults("test")
		return spec
	}

	assert.Equal(t, "arangodb/arangodb:latest", def(DeploymentSpec{}).GetImage())
}

func TestDeploymentSpecResetImmutableFields(t *testing.T) {
	tests := []struct {
		Original DeploymentSpec
		Target   DeploymentSpec
		Expected DeploymentSpec
		Result   []string
	}{
		// Valid "changes"
		{
			DeploymentSpec{Image: util.NewString("foo")},
			DeploymentSpec{Image: util.NewString("foo2")},
			DeploymentSpec{Image: util.NewString("foo2")},
			nil,
		},
		{
			DeploymentSpec{ImagePullPolicy: util.NewPullPolicy(v1.PullAlways)},
			DeploymentSpec{ImagePullPolicy: util.NewPullPolicy(v1.PullNever)},
			DeploymentSpec{ImagePullPolicy: util.NewPullPolicy(v1.PullNever)},
			nil,
		},

		// Invalid changes
		{
			DeploymentSpec{Mode: NewMode(DeploymentModeSingle)},
			DeploymentSpec{Mode: NewMode(DeploymentModeCluster)},
			DeploymentSpec{Mode: NewMode(DeploymentModeSingle)},
			[]string{"mode"},
		},
	}

	for _, test := range tests {
		result := test.Original.ResetImmutableFields(&test.Target)
		assert.Equal(t, test.Result, result)
		assert.Equal(t, test.Expected, test.Target)
	}
}
