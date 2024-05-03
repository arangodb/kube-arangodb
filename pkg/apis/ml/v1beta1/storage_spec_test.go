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

package v1beta1

import (
	"testing"

	"github.com/stretchr/testify/require"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	schedulerContainerApi "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1beta1/container"
	schedulerContainerResourcesApi "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1beta1/container/resources"
	sharedApi "github.com/arangodb/kube-arangodb/pkg/apis/shared/v1"
	"github.com/arangodb/kube-arangodb/pkg/util"
)

func Test_ArangoMLStorageSpec(t *testing.T) {
	s := ArangoMLStorageSpec{}
	require.Nil(t, s.GetMode())
	require.Nil(t, s.GetBackend())
	require.Error(t, s.Validate())

	s.Mode = &ArangoMLStorageSpecMode{}
	require.Nil(t, s.Mode.GetSidecar())

	s.Backend = &ArangoMLStorageSpecBackend{}
	require.Nil(t, s.Backend.GetS3())
	require.Error(t, s.Validate())

	s.Mode.Sidecar = &ArangoMLStorageSpecModeSidecar{}
	require.Nil(t, s.Mode.Sidecar.GetResources())
	require.NotNil(t, s.Mode.Sidecar.GetListenPort())

	require.Error(t, s.Backend.S3.Validate())
	s.Backend.S3 = &ArangoMLStorageSpecBackendS3{
		Endpoint: util.NewType("http://test.s3.example.com"),
		CredentialsSecret: &sharedApi.Object{
			Name:      "a-secret",
			Namespace: nil,
		},
	}
	s.BucketName = util.NewType("bucket")
	require.NoError(t, s.Validate())

	t.Run("default requests and limits assigned", func(t *testing.T) {
		assignedRequirements := core.ResourceRequirements{
			Requests: core.ResourceList{
				core.ResourceCPU:    resource.MustParse("100m"),
				core.ResourceMemory: resource.MustParse("128Mi"),
			},
		}
		s.Mode.Sidecar.Container = &schedulerContainerApi.Container{}
		s.Mode.Sidecar.Resources = &schedulerContainerResourcesApi.Resources{Resources: &assignedRequirements}

		expectedRequirements := core.ResourceRequirements{
			Requests: assignedRequirements.Requests,
			Limits: core.ResourceList{
				core.ResourceCPU:    resource.MustParse("200m"),
				core.ResourceMemory: resource.MustParse("256Mi"),
			},
		}

		actualRequirements := s.Mode.Sidecar.GetResources().With(&schedulerContainerResourcesApi.Resources{Resources: &core.ResourceRequirements{
			Limits: core.ResourceList{
				core.ResourceCPU:    resource.MustParse("200m"),
				core.ResourceMemory: resource.MustParse("256Mi"),
			},
			Requests: core.ResourceList{
				core.ResourceCPU:    resource.MustParse("100m"),
				core.ResourceMemory: resource.MustParse("128Mi"),
			},
		}})

		require.Equal(t, expectedRequirements, actualRequirements.GetResources())
	})
}
