//
// DISCLAIMER
//
// Copyright 2023 ArangoDB GmbH, Cologne, Germany
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

package v1alpha1

import (
	"testing"

	"github.com/stretchr/testify/require"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/arangodb/kube-arangodb/pkg/util"
)

func Test_ArangoMLStorageSpec(t *testing.T) {
	s := ArangoMLStorageSpec{}
	require.Error(t, s.Validate())

	s.Mode = &ArangoMLStorageSpecMode{}
	s.Backend = &ArangoMLStorageSpecBackend{}
	require.Error(t, s.Validate())

	s.Mode.Sidecar = &ArangoMLStorageSpecModeSidecar{}
	s.Backend.S3 = &ArangoMLStorageSpecBackendS3{
		Endpoint:              util.NewType("http://test.s3.example.com"),
		BucketName:            util.NewType("bucket"),
		CredentialsSecretName: util.NewType("a-secret"),
	}
	require.NoError(t, s.Validate())

	t.Run("default requests and limits assigned", func(t *testing.T) {
		assignedRequirements := core.ResourceRequirements{
			Requests: core.ResourceList{
				core.ResourceCPU:    resource.MustParse("200m"),
				core.ResourceMemory: resource.MustParse("200Mi"),
			},
		}
		s.Mode.Sidecar.Resources = &assignedRequirements

		expectedRequirements := core.ResourceRequirements{
			Requests: assignedRequirements.Requests,
			Limits: core.ResourceList{
				core.ResourceCPU:    resource.MustParse("200m"),
				core.ResourceMemory: resource.MustParse("200Mi"),
			},
		}

		actualRequirements := s.Mode.Sidecar.GetResources()
		require.Equal(t, expectedRequirements, actualRequirements)
	})
}
