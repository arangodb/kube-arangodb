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

package cmd

import (
	"context"
	"os"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/arangodb/kube-arangodb/pkg/ml/storage"
	"github.com/arangodb/kube-arangodb/pkg/util"
)

var (
	cmdMLStorage = &cobra.Command{
		Use: "storage",
		Run: func(cmd *cobra.Command, args []string) {
		},
		Hidden: true,
	}

	cmdMLStorageS3 = &cobra.Command{
		Use:   "s3",
		Short: "Run a GRPC service implementing the arangodb.operator.ml.storage.v1 API. Enterprise Edition only",
		Run:   cmdMLStorageS3Run,
	}

	cmdMLStorageS3Options struct {
		storage.ServiceConfig
	}
)

func init() {
	cmdML.AddCommand(cmdMLStorage)
	cmdMLStorage.AddCommand(cmdMLStorageS3)

	f := cmdMLStorageS3.PersistentFlags()
	f.StringVar(&cmdMLStorageS3Options.ListenAddress, "server.address", "", "Address the GRPC service will listen on (IP:port)")

	f.StringVar(&cmdMLStorageS3Options.S3.Endpoint, "s3.endpoint", "", "Endpoint of S3 API implementation")
	f.StringVar(&cmdMLStorageS3Options.S3.CACrtFile, "s3.ca-crt", "", "Path to file containing CA certificate to validate endpoint connection")
	f.StringVar(&cmdMLStorageS3Options.S3.CAKeyFile, "s3.ca-key", "", "Path to file containing keyfile to validate endpoint connection")
	f.BoolVar(&cmdMLStorageS3Options.S3.AllowInsecure, "s3.allow-insecure", false, "If set to true, the Endpoint certificates won't be checked")
	f.BoolVar(&cmdMLStorageS3Options.S3.DisableSSL, "s3.disable-ssl", false, "If set to true, the SSL won't be used when connecting to Endpoint")
	f.StringVar(&cmdMLStorageS3Options.S3.Region, "s3.region", "", "Region")
	f.StringVar(&cmdMLStorageS3Options.S3.BucketName, "s3.bucket", "", "Bucket name")
	f.StringVar(&cmdMLStorageS3Options.S3.AccessKeyFile, "s3.access-key", "", "Path to file containing S3 AccessKey")
	f.StringVar(&cmdMLStorageS3Options.S3.SecretKeyFile, "s3.secret-key", "", "Path to file containing S3 SecretKey")
}

func cmdMLStorageS3Run(cmd *cobra.Command, _ []string) {
	if err := cmdMLStorageS3RunE(cmd); err != nil {
		log.Error().Err(err).Msgf("Fatal")
		os.Exit(1)
	}
}

func cmdMLStorageS3RunE(_ *cobra.Command) error {
	ctx := util.CreateSignalContext(context.Background())

	svc, err := storage.NewService(ctx, storage.StorageTypeS3Proxy, cmdMLStorageS3Options.ServiceConfig)
	if err != nil {
		return err
	}

	return svc.Run(ctx)
}
