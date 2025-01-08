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
	awsHelper "github.com/arangodb/kube-arangodb/pkg/util/aws"
)

const (
	TestAwsProfile util.EnvironmentVariable = "TEST_AWS_PROFILE"
	TestAwsRole    util.EnvironmentVariable = "TEST_AWS_ROLE"
	TestAWSBucket  util.EnvironmentVariable = "TEST_AWS_BUCKET"
)

func GetAWSS3Bucket(t *testing.T) string {
	b, ok := TestAWSBucket.Lookup()
	if !ok {
		t.Skipf("Bucket does not exists")
	}

	return b
}

func GetAWSClientConfig(t *testing.T) awsHelper.Config {
	v, ok := TestAwsProfile.Lookup()
	if !ok {
		t.Skipf("Client does not exists")
	}

	var c awsHelper.Config
	c.Region = "eu-central-1"

	c.Provider.Config = awsHelper.ProviderConfig{
		Profile: v,
	}
	c.Provider.Type = awsHelper.ProviderTypeConfig

	r, ok := TestAwsRole.Lookup()
	if ok {
		c.Provider.Impersonate = awsHelper.ProviderImpersonate{
			Impersonate: true,
			Role:        r,
			Name:        "Test",
		}
	}

	return c
}
