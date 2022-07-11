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

package v1

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	core "k8s.io/api/core/v1"

	"github.com/arangodb/kube-arangodb/pkg/apis/shared"
)

const (
	labelValidationError = "Validation of label failed: a lowercase RFC 1123 label must consist of lower case alphanumeric characters or '-', and must start and end with an alphanumeric character (e.g. 'my-name',  or '123-abc', regex used for validation is '[a-z0-9]([-a-z0-9]*[a-z0-9])?')"
	invalidName          = "-invalid"
	validName            = "valid"
)

func Test_Volume_Validation(t *testing.T) {
	cases := []struct {
		name         string
		volumes      ServerGroupSpecVolumes
		fail         bool
		failedFields map[string]string
	}{
		{
			name: "Nil definition",
		},
		{
			name: "Invalid name",

			fail: true,
			failedFields: map[string]string{
				"0.name":              labelValidationError,
				"0.secret.secretName": labelValidationError,
			},

			volumes: []ServerGroupSpecVolume{
				{
					Name: invalidName,
					Secret: &ServerGroupSpecVolumeSecret{
						SecretName: invalidName,
					},
				},
			},
		},
		{
			name: "Restricted name",

			fail: true,
			failedFields: map[string]string{
				"": fmt.Sprintf("volume with name %s is restricted", restrictedVolumeNames[0]),
			},

			volumes: []ServerGroupSpecVolume{
				{
					Name: restrictedVolumeNames[0],
					Secret: &ServerGroupSpecVolumeSecret{
						SecretName: validName,
					},
				},
			},
		},
		{
			name: "Defined multiple sources",

			fail: true,
			failedFields: map[string]string{
				"0": "only one option can be defined: secret, configMap or emptyDir",
			},

			volumes: []ServerGroupSpecVolume{
				{
					Name: validName,
					Secret: &ServerGroupSpecVolumeSecret{
						SecretName: validName,
					},
					ConfigMap: &ServerGroupSpecVolumeConfigMap{
						LocalObjectReference: core.LocalObjectReference{
							Name: validName,
						},
					},
				},
			},
		},
		{
			name: "Defined multiple volumes with same name",

			fail: true,
			failedFields: map[string]string{
				"": "volume with name valid defined more than once: 2",
			},

			volumes: []ServerGroupSpecVolume{
				{
					Name: validName,
					Secret: &ServerGroupSpecVolumeSecret{
						SecretName: validName,
					},
				},
				{
					Name: validName,
					Secret: &ServerGroupSpecVolumeSecret{
						SecretName: validName,
					},
				},
			},
		},
		{
			name: "Defined multiple volumes",

			volumes: []ServerGroupSpecVolume{
				{
					Name: validName,
					Secret: &ServerGroupSpecVolumeSecret{
						SecretName: validName,
					},
				},
				{
					Name: "valid-2",
					ConfigMap: &ServerGroupSpecVolumeConfigMap{
						LocalObjectReference: core.LocalObjectReference{
							Name: validName,
						},
					},
				},
			},
		},
		{
			name: "Templating",
			volumes: []ServerGroupSpecVolume{
				{
					Name: validName,
					Secret: &ServerGroupSpecVolumeSecret{
						SecretName: fmt.Sprintf("${%s}-secret", ServerGroupSpecVolumeRenderParamDeploymentName),
					},
				},
			},
		},
		{
			name: "Invalid templating",
			volumes: []ServerGroupSpecVolume{
				{
					Name: validName,
					Secret: &ServerGroupSpecVolumeSecret{
						SecretName: fmt.Sprintf("${%sRANDOM}-secret", ServerGroupSpecVolumeRenderParamDeploymentName),
					},
				},
			},
			fail: true,
			failedFields: map[string]string{
				"0.secret.secretName": labelValidationError,
			},
		},
		{
			name: "Templating with group name",
			volumes: []ServerGroupSpecVolume{
				{
					Name: validName,
					Secret: &ServerGroupSpecVolumeSecret{
						SecretName: fmt.Sprintf("${%s}-${%s}-${%s}-cache", ServerGroupSpecVolumeRenderParamDeploymentName, ServerGroupSpecVolumeRenderParamMemberRole, ServerGroupSpecVolumeRenderParamMemberID),
					},
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := c.volumes.Validate()

			if c.fail {
				require.Error(t, err)

				mergedErr, ok := err.(shared.MergedErrors)
				require.True(t, ok, "Is not MergedError type")

				require.Equal(t, len(mergedErr.Errors()), len(c.failedFields), "Count of expected fields and merged errors does not match")

				for _, fieldError := range mergedErr.Errors() {
					resourceErr, ok := fieldError.(shared.ResourceError)
					if !ok {
						resourceErr = shared.ResourceError{
							Prefix: "",
							Err:    fieldError,
						}
					}

					errValue, ok := c.failedFields[resourceErr.Prefix]
					require.True(t, ok, "unexpected prefix %s", resourceErr.Prefix)

					require.EqualError(t, resourceErr.Err, errValue)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}
