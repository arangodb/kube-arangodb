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

package v2

import (
	"context"
	_ "embed"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/util/uuid"

	pbStorageV2 "github.com/arangodb/kube-arangodb/integrations/storage/v2/definition"
	pbImplStorageV2SharedS3 "github.com/arangodb/kube-arangodb/integrations/storage/v2/shared/s3"
	"github.com/arangodb/kube-arangodb/pkg/logging"
	"github.com/arangodb/kube-arangodb/pkg/util"
	awsHelper "github.com/arangodb/kube-arangodb/pkg/util/aws"
	"github.com/arangodb/kube-arangodb/pkg/util/svc"
	"github.com/arangodb/kube-arangodb/pkg/util/tests/tgrpc"
)

const (
	TestAwsProfile util.EnvironmentVariable = "TEST_AWS_PROFILE"
	TestAwsRole    util.EnvironmentVariable = "TEST_AWS_ROLE"
	TestAWSBucket  util.EnvironmentVariable = "TEST_AWS_BUCKET"
)

func getClient(t *testing.T, mods ...Mod) Configuration {
	v, ok := TestAwsProfile.Lookup()
	if !ok {
		t.Skipf("Client does not exists")
	}

	b, ok := TestAWSBucket.Lookup()
	if !ok {
		t.Skipf("Bucket does not exists")
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

	var scfg pbImplStorageV2SharedS3.Configuration

	scfg.Client = c
	scfg.BucketName = b
	scfg.BucketPrefix = fmt.Sprintf("test/%s/", uuid.NewUUID())

	var cfg Configuration

	cfg.Type = ConfigurationTypeS3
	cfg.S3 = scfg

	return cfg.With(mods...)
}

func init() {
	logging.Global().ApplyLogLevels(map[string]logging.Level{
		logging.TopicAll: logging.Debug,
	})
}

func Handler(t *testing.T, mods ...Mod) svc.Handler {
	handler, err := New(getClient(t).With(mods...))
	require.NoError(t, err)

	return handler
}

func Client(t *testing.T, ctx context.Context, mods ...Mod) pbStorageV2.StorageV2Client {
	local := svc.NewService(svc.Configuration{
		Address: "127.0.0.1:0",
	}, Handler(t, mods...))

	start := local.Start(ctx)

	return tgrpc.NewGRPCClient(t, ctx, pbStorageV2.NewStorageV2Client, start.Address())
}
