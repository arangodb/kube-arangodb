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
	"errors"

	"cloud.google.com/go/storage"

	pbImplStorageV2Shared "github.com/arangodb/kube-arangodb/integrations/storage/v2/shared"
)

func (i *ios) Head(ctx context.Context, key string) (*pbImplStorageV2Shared.Info, error) {
	b := i.client.Bucket(i.config.BucketName)

	obj := b.Object(i.key(key))

	attr, err := obj.Attrs(ctx)

	if err != nil {
		if errors.Is(err, storage.ErrObjectNotExist) {
			return nil, nil
		}
		return nil, err
	}

	return &pbImplStorageV2Shared.Info{
		Size:          uint64(attr.Size),
		LastUpdatedAt: attr.Updated,
	}, nil
}
