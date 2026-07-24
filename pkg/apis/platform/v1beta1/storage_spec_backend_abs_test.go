//
// DISCLAIMER
//
// Copyright 2026 ArangoDB GmbH, Cologne, Germany
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

// Test_ArangoPlatformStorageSpecBackendAzureBlobStorage_Credentials asserts that exactly one of the
// two credential sources (service-principal secret or client certificate) must be defined.
func Test_ArangoPlatformStorageSpecBackendAzureBlobStorage_Credentials(t *testing.T) {
	base := func() *ArangoPlatformStorageSpecBackendAzureBlobStorage {
		return &ArangoPlatformStorageSpecBackendAzureBlobStorage{
			TenantID:    util.NewType("tenant"),
			AccountName: util.NewType("account"),
			BucketName:  util.NewType("bucket"),
		}
	}

	t.Run("client secret only", func(t *testing.T) {
		s := base()
		s.CredentialsSecret = &sharedApi.Object{Name: "az-secret"}
		require.NoError(t, s.Validate())
	})

	t.Run("client certificate only", func(t *testing.T) {
		s := base()
		s.ClientCertificateSecret = &sharedApi.Object{Name: "az-tls"}
		require.NoError(t, s.Validate())
	})

	t.Run("both is mutually exclusive", func(t *testing.T) {
		s := base()
		s.CredentialsSecret = &sharedApi.Object{Name: "az-secret"}
		s.ClientCertificateSecret = &sharedApi.Object{Name: "az-tls"}
		require.ErrorContains(t, s.Validate(), "mutually exclusive")
	})

	t.Run("neither is required", func(t *testing.T) {
		require.ErrorContains(t, base().Validate(), "one of credentialsSecret or clientCertificateSecret")
	})
}
