//
// DISCLAIMER
//
// Copyright 2024-2025 ArangoDB GmbH, Cologne, Germany
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

	sharedApi "github.com/arangodb/kube-arangodb/pkg/apis/shared/v1"
	"github.com/arangodb/kube-arangodb/pkg/util"
)

func Test_ArangoPlatformStorageSpec(t *testing.T) {
	s := ArangoPlatformStorageSpec{}
	require.Nil(t, s.GetBackend())
	require.Error(t, s.Validate())

	s.Backend = &ArangoPlatformStorageSpecBackend{}
	require.Nil(t, s.Backend.GetS3())
	require.Error(t, s.Validate())

	require.Error(t, s.Backend.S3.Validate())
	s.Backend.S3 = &ArangoPlatformStorageSpecBackendS3{
		Endpoint: util.NewType("http://test.s3.example.com"),
		CredentialsSecret: &sharedApi.Object{
			Name:      "a-secret",
			Namespace: nil,
		},
		BucketName: util.NewType("bucket"),
	}
	require.NoError(t, s.Validate())
}
