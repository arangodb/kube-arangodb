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

package sidecar

import (
	"context"

	"github.com/spf13/cobra"

	pbImplStorageV2 "github.com/arangodb/kube-arangodb/integrations/storage/v2"
	"github.com/arangodb/kube-arangodb/pkg/util/svc"
)

var storageV2CLI = pbImplStorageV2.NewCLI("storage.v2")

func init() {
	register("storage-v2", func(ctx context.Context, cmd *cobra.Command) (svc.Handler, bool, error) {
		if v, err := flagCentralServicesEnabled.Get(cmd); err != nil {
			return nil, false, err
		} else if !v {
			return nil, false, nil
		}

		cfg, err := storageV2CLI.Configuration(cmd)
		if err != nil {
			return nil, false, err
		}

		if !storageV2BackendConfigured(cfg) {
			return nil, false, nil
		}

		handler, err := pbImplStorageV2.New(ctx, cfg)
		if err != nil {
			return nil, false, err
		}

		return handler, true, nil
	}, storageV2CLI)
}

func storageV2BackendConfigured(cfg pbImplStorageV2.Configuration) bool {
	switch cfg.Type {
	case pbImplStorageV2.ConfigurationTypeS3:
		return cfg.S3.BucketName != ""
	case pbImplStorageV2.ConfigurationTypeGCS:
		return cfg.GCS.BucketName != ""
	case pbImplStorageV2.ConfigurationTypeAzure:
		return cfg.AzureBlobStorage.BucketName != ""
	}
	return false
}
