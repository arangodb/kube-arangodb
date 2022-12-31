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

package storage

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TestCreateValidEndpointList tests createValidEndpointList.
func TestCreateValidEndpointList(t *testing.T) {
	tests := []struct {
		Input    *core.EndpointsList
		Expected []string
	}{
		{
			Input:    &core.EndpointsList{},
			Expected: []string{},
		},
		{
			Input: &core.EndpointsList{
				Items: []core.Endpoints{
					core.Endpoints{
						Subsets: []core.EndpointSubset{
							core.EndpointSubset{
								Addresses: []core.EndpointAddress{
									core.EndpointAddress{
										IP: "1.2.3.4",
									},
								},
							},
							core.EndpointSubset{
								Addresses: []core.EndpointAddress{
									core.EndpointAddress{
										IP: "5.6.7.8",
									},
									core.EndpointAddress{
										IP: "9.10.11.12",
									},
								},
							},
						},
					},
				},
			},
			Expected: []string{
				"1.2.3.4:8929",
				"5.6.7.8:8929",
				"9.10.11.12:8929",
			},
		},
	}
	for _, test := range tests {
		output := createValidEndpointList(test.Input)
		assert.Equal(t, test.Expected, output)
	}
}

// TestCreateNodeSelector tests createNodeSelector.
func TestCreateNodeSelector(t *testing.T) {
	tests := map[string]string{
		"foo": "{\"nodeSelectorTerms\":[{\"matchExpressions\":[{\"key\":\"kubernetes.io/hostname\",\"operator\":\"In\",\"values\":[\"foo\"]}]}]}",
		"bar": "{\"nodeSelectorTerms\":[{\"matchExpressions\":[{\"key\":\"kubernetes.io/hostname\",\"operator\":\"In\",\"values\":[\"bar\"]}]}]}",
	}
	for input, expected := range tests {
		sel := createNodeSelector(input)
		output, err := json.Marshal(sel)
		assert.NoError(t, err)
		assert.Equal(t, expected, string(output), "Input: '%s'", input)
	}
}

// TestGetDeploymentInfo tests getDeploymentInfo.
func TestGetDeploymentInfo(t *testing.T) {
	tests := []struct {
		Input                       core.PersistentVolumeClaim
		ExpectedDeploymentName      string
		ExpectedRole                string
		ExpectedEnforceAntiAffinity bool
	}{
		{
			Input:                       core.PersistentVolumeClaim{},
			ExpectedDeploymentName:      "",
			ExpectedRole:                "",
			ExpectedEnforceAntiAffinity: false,
		},
		{
			Input: core.PersistentVolumeClaim{
				ObjectMeta: meta.ObjectMeta{
					Annotations: map[string]string{
						"database.arangodb.com/enforce-anti-affinity": "true",
					},
					Labels: map[string]string{
						"arango_deployment": "foo",
						"role":              "r1",
					},
				},
			},
			ExpectedDeploymentName:      "foo",
			ExpectedRole:                "r1",
			ExpectedEnforceAntiAffinity: true,
		},
		{
			Input: core.PersistentVolumeClaim{
				ObjectMeta: meta.ObjectMeta{
					Annotations: map[string]string{
						"database.arangodb.com/enforce-anti-affinity": "false",
					},
					Labels: map[string]string{
						"arango_deployment": "foo",
						"role":              "r1",
					},
				},
			},
			ExpectedDeploymentName:      "foo",
			ExpectedRole:                "r1",
			ExpectedEnforceAntiAffinity: false,
		},
		{
			Input: core.PersistentVolumeClaim{
				ObjectMeta: meta.ObjectMeta{
					Annotations: map[string]string{
						"database.arangodb.com/enforce-anti-affinity": "wrong",
					},
					Labels: map[string]string{
						"arango_deployment": "bar",
						"role":              "r77",
					},
				},
			},
			ExpectedDeploymentName:      "bar",
			ExpectedRole:                "r77",
			ExpectedEnforceAntiAffinity: false,
		},
	}
	for _, test := range tests {
		deploymentName, role, enforceAntiAffinity := getDeploymentInfo(test.Input)
		assert.Equal(t, test.ExpectedDeploymentName, deploymentName)
		assert.Equal(t, test.ExpectedRole, role)
		assert.Equal(t, test.ExpectedEnforceAntiAffinity, enforceAntiAffinity)
	}
}

// TestShortHash tests shortHash.
func TestShortHash(t *testing.T) {
	tests := map[string]string{
		"foo": "0beec7",
		"":    "da39a3",
		"something very very very very very looooooooooooooooooooooooooooooooong": "68ff76",
	}
	for input, expected := range tests {
		output := shortHash(input)
		assert.Equal(t, expected, output, "Input: '%s'", input)
	}
}
