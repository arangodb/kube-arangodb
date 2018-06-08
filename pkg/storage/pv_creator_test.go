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

package storage

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arangodb/kube-arangodb/pkg/storage/provisioner"
	"github.com/arangodb/kube-arangodb/pkg/storage/provisioner/mocks"
)

// TestCreateValidEndpointList tests createValidEndpointList.
func TestCreateValidEndpointList(t *testing.T) {
	tests := []struct {
		Input    *v1.EndpointsList
		Expected []string
	}{
		{
			Input:    &v1.EndpointsList{},
			Expected: []string{},
		},
		{
			Input: &v1.EndpointsList{
				Items: []v1.Endpoints{
					v1.Endpoints{
						Subsets: []v1.EndpointSubset{
							v1.EndpointSubset{
								Addresses: []v1.EndpointAddress{
									v1.EndpointAddress{
										IP: "1.2.3.4",
									},
								},
							},
							v1.EndpointSubset{
								Addresses: []v1.EndpointAddress{
									v1.EndpointAddress{
										IP: "5.6.7.8",
									},
									v1.EndpointAddress{
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

// TestCreateNodeAffinity tests createNodeAffinity.
func TestCreateNodeAffinity(t *testing.T) {
	tests := map[string]string{
		"foo": "{\"requiredDuringSchedulingIgnoredDuringExecution\":{\"nodeSelectorTerms\":[{\"matchExpressions\":[{\"key\":\"kubernetes.io/hostname\",\"operator\":\"In\",\"values\":[\"foo\"]}]}]}}",
		"bar": "{\"requiredDuringSchedulingIgnoredDuringExecution\":{\"nodeSelectorTerms\":[{\"matchExpressions\":[{\"key\":\"kubernetes.io/hostname\",\"operator\":\"In\",\"values\":[\"bar\"]}]}]}}",
	}
	for input, expected := range tests {
		output, err := createNodeAffinity(input)
		assert.NoError(t, err)
		assert.Equal(t, expected, output, "Input: '%s'", input)
	}
}

// TestCreateNodeClientMap tests createNodeClientMap.
func TestCreateNodeClientMap(t *testing.T) {
	GB := int64(1024 * 1024 * 1024)
	foo := mocks.NewProvisioner("foo", 100*GB, 100*GB)
	bar := mocks.NewProvisioner("bar", 100*GB, 100*GB)
	tests := []struct {
		Input    []provisioner.API
		Expected map[string]provisioner.API
	}{
		{
			Input:    nil,
			Expected: map[string]provisioner.API{},
		},
		{
			Input: []provisioner.API{foo, bar},
			Expected: map[string]provisioner.API{
				"bar": bar,
				"foo": foo,
			},
		},
	}
	ctx := context.Background()
	for _, test := range tests {
		output := createNodeClientMap(ctx, test.Input)
		assert.Equal(t, test.Expected, output)
	}
}

// TestGetDeploymentInfo tests getDeploymentInfo.
func TestGetDeploymentInfo(t *testing.T) {
	tests := []struct {
		Input                       v1.PersistentVolumeClaim
		ExpectedDeploymentName      string
		ExpectedRole                string
		ExpectedEnforceAntiAffinity bool
	}{
		{
			Input: v1.PersistentVolumeClaim{},
			ExpectedDeploymentName:      "",
			ExpectedRole:                "",
			ExpectedEnforceAntiAffinity: false,
		},
		{
			Input: v1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
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
			Input: v1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
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
			Input: v1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
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
