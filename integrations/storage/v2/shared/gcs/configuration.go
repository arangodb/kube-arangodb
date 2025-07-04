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

package gcs

import (
	"context"

	"cloud.google.com/go/storage"

	pbImplStorageV2Shared "github.com/arangodb/kube-arangodb/integrations/storage/v2/shared"
	gcsHelper "github.com/arangodb/kube-arangodb/pkg/util/gcs"
)

type Configuration struct {
	BucketName   string
	BucketPrefix string

	MaxListKeys *int64

	Client gcsHelper.Config
}

func (c Configuration) New(ctx context.Context) (pbImplStorageV2Shared.IO, error) {
	opts, err := c.Client.GetClientOptions()
	if err != nil {
		return nil, err
	}

	client, err := storage.NewClient(ctx, opts...)
	if err != nil {
		return nil, err
	}

	return &ios{
		config: c,
		client: client,
	}, nil
}
