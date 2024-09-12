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
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
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

func (b *storageV1) Register(cmd *cobra.Command, fs FlagEnvHandler) error {
	return errors.Errors(
		fs.StringVar((*string)(&b.Configuration.Type), "type", string(storage.S3), "Type of the Storage Integration"),
		fs.StringVar(&b.Configuration.S3.Endpoint, "s3.endpoint", "", "Endpoint of S3 API implementation"),
		fs.StringVar(&b.Configuration.S3.CACrtFile, "s3.ca-crt", "", "Path to file containing CA certificate to validate endpoint connection"),
		fs.StringVar(&b.Configuration.S3.CAKeyFile, "s3.ca-key", "", "Path to file containing keyfile to validate endpoint connection"),
		fs.BoolVar(&b.Configuration.S3.AllowInsecure, "s3.allow-insecure", false, "If set to true, the Endpoint certificates won't be checked"),
		fs.BoolVar(&b.Configuration.S3.DisableSSL, "s3.disable-ssl", false, "If set to true, the SSL won't be used when connecting to Endpoint"),
		fs.StringVar(&b.Configuration.S3.Region, "s3.region", "", "Region"),
		fs.StringVar(&b.Configuration.S3.BucketName, "s3.bucket", "", "Bucket name"),
		fs.StringVar(&b.Configuration.S3.AccessKeyFile, "s3.access-key", "", "Path to file containing S3 AccessKey"),
		fs.StringVar(&b.Configuration.S3.SecretKeyFile, "s3.secret-key", "", "Path to file containing S3 SecretKey"),
	)
}

func (b *storageV1) Handler(ctx context.Context, cmd *cobra.Command) (svc.Handler, error) {
	return storage.NewService(ctx, b.Configuration)
}

func (*storageV1) Init(ctx context.Context, cmd *cobra.Command) error {
	return nil
}
