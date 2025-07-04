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
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

func (i *ios) Init(ctx context.Context, opts *pbImplStorageV2Shared.InitOptions) error {
	if opts.GetCreate() {
		b := i.client.Bucket(i.config.BucketName)

		if _, err := b.Attrs(ctx); err != nil {
			if errors.Is(err, storage.ErrBucketNotExist) {
				if err := b.Create(ctx, i.config.Client.ProjectID, &storage.BucketAttrs{}); err != nil {
					return err
				}
			} else {
				return err
			}
		}
	}

	return nil
}
