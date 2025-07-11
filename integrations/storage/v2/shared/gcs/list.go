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
	"io"
	"sync"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"

	pbImplStorageV2Shared "github.com/arangodb/kube-arangodb/integrations/storage/v2/shared"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

func (i *ios) List(ctx context.Context, key string) (util.NextIterator[[]pbImplStorageV2Shared.File], error) {
	return &listIterator{
		iter: i.client.Bucket(i.config.BucketName).Objects(ctx, &storage.Query{
			Prefix: i.key(key),
		}),
		parent: i,
	}, nil
}

type listIterator struct {
	lock sync.Mutex

	parent *ios

	iter *storage.ObjectIterator
}

func (l *listIterator) Next(ctx context.Context) ([]pbImplStorageV2Shared.File, error) {
	l.lock.Lock()
	defer l.lock.Unlock()

	results := make([]pbImplStorageV2Shared.File, 0, util.OptionalType(l.parent.config.MaxListKeys, 1000))

	for {
		attrs, err := l.iter.Next()
		if err != nil {
			if errors.Is(err, iterator.Done) {
				break
			}

			return nil, err
		}

		results = append(results, pbImplStorageV2Shared.File{
			Key: l.parent.clean(attrs.Name),
			Info: pbImplStorageV2Shared.Info{
				Size:          uint64(attrs.Size),
				LastUpdatedAt: attrs.Updated,
			},
		})

		if int64(len(results)) >= util.OptionalType(l.parent.config.MaxListKeys, 1000) {
			break
		}
	}

	if len(results) == 0 {
		return nil, io.EOF
	}

	return results, nil
}
