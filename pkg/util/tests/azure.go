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
	"testing"

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/azure"
)

const (
	TestAzureClientTenant util.EnvironmentVariable = "TEST_AZURE_CLIENT_TENANT"
	TestAzureClientID     util.EnvironmentVariable = "TEST_AZURE_CLIENT_ID"
	TestAzureClientSecret util.EnvironmentVariable = "TEST_AZURE_CLIENT_SECRET"
	TestAzureAccountName  util.EnvironmentVariable = "TEST_AZURE_ACCOUNT_NAME"
	TestAzureEndpoint     util.EnvironmentVariable = "TEST_AZURE_ENDPOINT"
	TestAzureContainer    util.EnvironmentVariable = "TEST_AZURE_CONTAINER"
)

func GetAzureBlobStorageContainer(t *testing.T) string {
	b, ok := TestAzureContainer.Lookup()
	if !ok {
		t.Skipf("Bucket does not exist")
	}

	return b
}

func GetAzureConfig(t *testing.T) azure.Config {
	p := GetAzureProvider(t)

	var z azure.Config

	z.AccountName = TestAzureAccountName.GetOrDefault("")
	z.Endpoint = TestAzureEndpoint.GetOrDefault("")
	z.Provider = p

	return z
}

func GetAzureProvider(t *testing.T) azure.Provider {
	v, ok := TestAzureClientTenant.Lookup()
	if !ok {
		t.Skipf("Tenant Not Provided")
	}

	var p azure.Provider

	p.TenantID = v

	p.Type = azure.ProviderTypeSecret

	p.Secret.ClientID = TestAzureClientID.Require(t)
	p.Secret.ClientSecret = TestAzureClientSecret.Require(t)

	return p
}
