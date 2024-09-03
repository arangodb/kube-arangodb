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

package integrations

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/arangodb/kube-arangodb/pkg/ml/storage"
	"github.com/arangodb/kube-arangodb/pkg/util/svc"
)

func init() {
	registerer.Register(storage.Name, func() Integration {
		return &storageV1{}
	})
}

type storageV1 struct {
	Configuration storage.Configuration
}

func (b *storageV1) Name() string {
	return storage.Name
}

func (b *storageV1) Description() string {
	return "StorageBucket Integration"
}

func (b *storageV1) Register(cmd *cobra.Command, arg ArgGen) error {
	f := cmd.Flags()

	f.StringVar((*string)(&b.Configuration.Type), arg("type"), string(storage.S3), "Type of the Storage Integration")
	f.StringVar(&b.Configuration.S3.Endpoint, arg("s3.endpoint"), "", "Endpoint of S3 API implementation")
	f.StringVar(&b.Configuration.S3.CACrtFile, arg("s3.ca-crt"), "", "Path to file containing CA certificate to validate endpoint connection")
	f.StringVar(&b.Configuration.S3.CAKeyFile, arg("s3.ca-key"), "", "Path to file containing keyfile to validate endpoint connection")
	f.BoolVar(&b.Configuration.S3.AllowInsecure, arg("s3.allow-insecure"), false, "If set to true, the Endpoint certificates won't be checked")
	f.BoolVar(&b.Configuration.S3.DisableSSL, arg("s3.disable-ssl"), false, "If set to true, the SSL won't be used when connecting to Endpoint")
	f.StringVar(&b.Configuration.S3.Region, arg("s3.region"), "", "Region")
	f.StringVar(&b.Configuration.S3.BucketName, arg("s3.bucket"), "", "Bucket name")
	f.StringVar(&b.Configuration.S3.AccessKeyFile, arg("s3.access-key"), "", "Path to file containing S3 AccessKey")
	f.StringVar(&b.Configuration.S3.SecretKeyFile, arg("s3.secret-key"), "", "Path to file containing S3 SecretKey")

	return nil
}

func (b *storageV1) Handler(ctx context.Context, cmd *cobra.Command) (svc.Handler, error) {
	return storage.NewService(ctx, b.Configuration)
}

func (*storageV1) Init(ctx context.Context, cmd *cobra.Command) error {
	return nil
}
