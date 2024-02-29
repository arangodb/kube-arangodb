//
// DISCLAIMER
//
// Copyright 2024 ArangoDB GmbH, Cologne, Germany
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

package cmd

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
	"github.com/arangodb/kube-arangodb/pkg/util/tests"
)

func Test_PodDiscovery(t *testing.T) {
	operatorImageDiscovery.timeout = time.Millisecond

	type testCase struct {
		Name string

		Pod core.Pod

		Image, ServiceAccount string

		Valid bool

		DefaultStatusDiscovery *bool
	}

	var testCases = []testCase{
		{
			Name:  "Empty pod",
			Valid: false,
		},
		{
			Name:  "Not allowed containers",
			Valid: false,
			Pod: core.Pod{
				ObjectMeta: meta.ObjectMeta{
					Name:      "operator",
					Namespace: tests.FakeNamespace,
				},
				Spec: core.PodSpec{
					Containers: []core.Container{
						{
							Name:  "unknown",
							Image: "image1",
						},
					},
				},
				Status: core.PodStatus{
					ContainerStatuses: []core.ContainerStatus{
						{
							Name:    "unknown",
							Image:   "image1",
							ImageID: "image1",
						},
					},
				},
			},
		},
		{
			Name:           "Allowed Status & Spec",
			Valid:          true,
			Image:          "image1",
			ServiceAccount: "sa",
			Pod: core.Pod{
				ObjectMeta: meta.ObjectMeta{
					Name:      "operator",
					Namespace: tests.FakeNamespace,
				},
				Spec: core.PodSpec{
					ServiceAccountName: "sa",
					Containers: []core.Container{
						{
							Name:  "operator",
							Image: "image1",
						},
					},
				},
				Status: core.PodStatus{
					ContainerStatuses: []core.ContainerStatus{
						{
							Name:    "operator",
							Image:   "image1",
							ImageID: "image1",
						},
					},
				},
			},
		},
		{
			Name:           "Allowed Status & Spec",
			Valid:          true,
			Image:          "imageStatusID1",
			ServiceAccount: "sa",
			Pod: core.Pod{
				ObjectMeta: meta.ObjectMeta{
					Name:      "operator",
					Namespace: tests.FakeNamespace,
				},
				Spec: core.PodSpec{
					ServiceAccountName: "sa",
					Containers: []core.Container{
						{
							Name:  "operator",
							Image: "imageSpec1",
						},
					},
				},
				Status: core.PodStatus{
					ContainerStatuses: []core.ContainerStatus{
						{
							Name:    "operator",
							Image:   "imageStatus1",
							ImageID: "imageStatusID1",
						},
					},
				},
			},
		},
		{
			Name:                   "Allowed Status & Spec - From Spec",
			Valid:                  true,
			Image:                  "imageSpec1",
			ServiceAccount:         "sa",
			DefaultStatusDiscovery: util.NewType(false),
			Pod: core.Pod{
				ObjectMeta: meta.ObjectMeta{
					Name:      "operator",
					Namespace: tests.FakeNamespace,
				},
				Spec: core.PodSpec{
					ServiceAccountName: "sa",
					Containers: []core.Container{
						{
							Name:  "operator",
							Image: "imageSpec1",
						},
					},
				},
				Status: core.PodStatus{
					ContainerStatuses: []core.ContainerStatus{
						{
							Name:    "operator",
							Image:   "imageStatus1",
							ImageID: "imageStatusID1",
						},
					},
				},
			},
		},
		{
			Name:           "Allowed Spec",
			Valid:          true,
			Image:          "imageSpec1",
			ServiceAccount: "sa",
			Pod: core.Pod{
				ObjectMeta: meta.ObjectMeta{
					Name:      "operator",
					Namespace: tests.FakeNamespace,
				},
				Spec: core.PodSpec{
					ServiceAccountName: "sa",
					Containers: []core.Container{
						{
							Name:  "operator",
							Image: "imageSpec1",
						},
					},
				},
			},
		},
		{
			Name:           "Allowed Status & Spec - From Second Pod",
			Valid:          true,
			Image:          "imageStatusID2",
			ServiceAccount: "sa",
			Pod: core.Pod{
				ObjectMeta: meta.ObjectMeta{
					Name:      "operator",
					Namespace: tests.FakeNamespace,
				},
				Spec: core.PodSpec{
					ServiceAccountName: "sa",
					Containers: []core.Container{
						{
							Name:  "test",
							Image: "imageSpec1",
						},
						{
							Name:  "operator",
							Image: "imageSpec2",
						},
					},
				},
				Status: core.PodStatus{
					ContainerStatuses: []core.ContainerStatus{
						{
							Name:    "test",
							Image:   "imageStatus1",
							ImageID: "imageStatusID1",
						},
						{
							Name:    "operator",
							Image:   "imageStatus2",
							ImageID: "imageStatusID2",
						},
					},
				},
			},
		},
		{
			Name:                   "Allowed Status & Spec - From Second Pod Spec",
			Valid:                  true,
			Image:                  "imageSpec2",
			ServiceAccount:         "sa",
			DefaultStatusDiscovery: util.NewType(false),
			Pod: core.Pod{
				ObjectMeta: meta.ObjectMeta{
					Name:      "operator",
					Namespace: tests.FakeNamespace,
				},
				Spec: core.PodSpec{
					ServiceAccountName: "sa",
					Containers: []core.Container{
						{
							Name:  "test",
							Image: "imageSpec1",
						},
						{
							Name:  "operator",
							Image: "imageSpec2",
						},
					},
				},
				Status: core.PodStatus{
					ContainerStatuses: []core.ContainerStatus{
						{
							Name:    "test",
							Image:   "imageStatus1",
							ImageID: "imageStatusID1",
						},
						{
							Name:    "operator",
							Image:   "imageStatus2",
							ImageID: "imageStatusID2",
						},
					},
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			operatorImageDiscovery.defaultStatusDiscovery = util.TypeOrDefault(testCase.DefaultStatusDiscovery, true)

			c := kclient.NewFakeClientBuilder().Add(&testCase.Pod).Client()
			image, sa, err := getMyPodInfo(c.Kubernetes(), tests.FakeNamespace, "operator")

			if !testCase.Valid {
				require.Error(t, err)
				require.Empty(t, image)
				require.Empty(t, sa)
			} else {
				require.NoError(t, err)
				require.Equal(t, testCase.Image, image)
				require.Equal(t, testCase.ServiceAccount, sa)
			}
		})
	}
}
