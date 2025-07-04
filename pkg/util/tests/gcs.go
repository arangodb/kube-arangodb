//
// DISCLAIMER
//
// Copyright 2025 ArangoDB GmbH, Cologne, Germany
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

package tests

import (
	"encoding/base64"
	"testing"

	"github.com/arangodb/kube-arangodb/pkg/util"
	gcsHelper "github.com/arangodb/kube-arangodb/pkg/util/gcs"
)

const (
	TestGcsCredentials     util.EnvironmentVariable = "TEST_GCS_CREDENTIALS"
	TestGcsCredentialsFile util.EnvironmentVariable = "TEST_GCS_CREDENTIALS_FILE"
	TestGcsProjectID       util.EnvironmentVariable = "TEST_GCS_PROJECT_ID"
	TestGcsBucket          util.EnvironmentVariable = "TEST_GCS_BUCKET"
)

func GetGCSBucket(t *testing.T) string {
	b, ok := TestGcsBucket.Lookup()
	if !ok {
		t.Skipf("Bucket does not exist")
	}

	return b
}

func GetGCSClientConfig(t *testing.T) gcsHelper.Config {
	p, ok := TestGcsProjectID.Lookup()
	if !ok {
		t.Skipf("ProjectID does not exist")
	}

	if v, ok := TestGcsCredentials.Lookup(); ok {
		if z, err := base64.StdEncoding.DecodeString(v); err == nil {
			return gcsHelper.Config{
				ProjectID: p,
				Provider: gcsHelper.Provider{
					Type: gcsHelper.ProviderTypeServiceAccount,
					ServiceAccount: gcsHelper.ProviderServiceAccount{
						JSON: string(z),
					},
				},
			}
		}
		return gcsHelper.Config{
			ProjectID: p,
			Provider: gcsHelper.Provider{
				Type: gcsHelper.ProviderTypeServiceAccount,
				ServiceAccount: gcsHelper.ProviderServiceAccount{
					JSON: v,
				},
			},
		}
	}

	if v, ok := TestGcsCredentialsFile.Lookup(); ok {
		return gcsHelper.Config{
			ProjectID: p,
			Provider: gcsHelper.Provider{
				Type: gcsHelper.ProviderTypeServiceAccount,
				ServiceAccount: gcsHelper.ProviderServiceAccount{
					File: v,
				},
			},
		}
	}

	t.Skipf("Credentials do not exist")

	return gcsHelper.Config{}
}
