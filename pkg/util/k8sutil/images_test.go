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

package k8sutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	core "k8s.io/api/core/v1"

	"github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

func TestGetArangoDBImageIDFromPod(t *testing.T) {
	type args struct {
		pod *core.Pod
	}
	tests := map[string]struct {
		args    args
		want    string
		wantErr error
	}{
		"pid is nil": {
			wantErr: errors.New("failed to get container statuses from nil pod"),
		},
		"container statuses list is empty": {
			args: args{
				pod: &core.Pod{},
			},
			wantErr: errors.New("empty list of ContainerStatuses"),
		},
		"image ID from the only container": {
			args: args{
				pod: &core.Pod{
					Status: core.PodStatus{
						ContainerStatuses: []core.ContainerStatus{
							{
								ImageID: dockerPullableImageIDPrefix + "test",
							},
						},
					},
				},
			},
			want: "test",
		},
		"image ID from two containers - first one as server": {
			args: args{
				pod: &core.Pod{
					Status: core.PodStatus{
						ContainerStatuses: []core.ContainerStatus{
							{
								Name:    shared.ServerContainerName,
								ImageID: dockerPullableImageIDPrefix + "test_arango",
							},
							{
								Name:    "other",
								ImageID: dockerPullableImageIDPrefix + "test1_arango",
							},
						},
					},
				},
			},
			want: "test_arango",
		},
		"image ID from two containers - second one as server": {
			args: args{
				pod: &core.Pod{
					Status: core.PodStatus{
						ContainerStatuses: []core.ContainerStatus{
							{
								Name:    "other",
								ImageID: dockerPullableImageIDPrefix + "test_arango",
							},
							{
								Name:    shared.ServerContainerName,
								ImageID: dockerPullableImageIDPrefix + "test1_arango",
							},
						},
					},
				},
			},
			want: "test1_arango",
		},
		"image ID from two containers - no server": {
			args: args{
				pod: &core.Pod{
					Status: core.PodStatus{
						ContainerStatuses: []core.ContainerStatus{
							{
								Name:    "other2",
								ImageID: dockerPullableImageIDPrefix + "test_arango",
							},
							{
								Name:    "other",
								ImageID: dockerPullableImageIDPrefix + "test1_arango",
							},
						},
					},
				},
			},
			want: "test_arango",
		},
	}

	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			got, err := GetArangoDBImageIDFromPod(testCase.args.pod)
			if testCase.wantErr != nil {
				require.EqualError(t, err, testCase.wantErr.Error())
				return
			}

			require.NoError(t, err)
			assert.Equalf(t, testCase.want, got, "image ID is not as expected")
		})
	}
}
