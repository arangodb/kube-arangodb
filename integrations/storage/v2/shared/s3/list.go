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

package s3

import (
	"context"
	"io"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go/service/s3"

	pbImplStorageV2Shared "github.com/arangodb/kube-arangodb/integrations/storage/v2/shared"
	"github.com/arangodb/kube-arangodb/pkg/util"
)

func (i *ios) List(_ context.Context, key string) (util.NextIterator[[]pbImplStorageV2Shared.File], error) {
	return &listIterator{
		parent: i,
		key:    key,
	}, nil
}

type listIterator struct {
	lock sync.Mutex

	parent *ios

	key string

	next *string

	done bool
}

func (l *listIterator) Next(ctx context.Context) ([]pbImplStorageV2Shared.File, error) {
	l.lock.Lock()
	defer l.lock.Unlock()

	if l.done {
		return nil, io.EOF
	}

	resp, err := l.parent.client.ListObjectsWithContext(ctx, &s3.ListObjectsInput{
		Bucket:  util.NewType(l.parent.config.BucketName),
		Prefix:  util.NewType(l.parent.key(l.key)),
		MaxKeys: l.parent.config.MaxListKeys,
		Marker:  l.next,
	})
	if err != nil {
		return nil, err
	}

	results := make([]pbImplStorageV2Shared.File, 0, len(resp.Contents))

	for _, obj := range resp.Contents {
		if obj == nil {
			continue
		}

		if obj.Key == nil {
			continue
		}

		var info pbImplStorageV2Shared.Info

		info.Size = uint64(util.TypeOrDefault(obj.Size))
		info.LastUpdatedAt = util.TypeOrDefault(obj.LastModified)

		results = append(results, pbImplStorageV2Shared.File{
			Key:  strings.TrimPrefix(*obj.Key, l.parent.key(l.key)),
			Info: info,
		})

		l.next = util.NewType(*obj.Key)
	}

	if !util.OptionalType(resp.IsTruncated, false) {
		l.done = true
	}

	return results, nil
}
