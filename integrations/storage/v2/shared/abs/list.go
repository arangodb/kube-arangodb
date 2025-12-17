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

package abs

import (
	"context"
	"io"
	"sync"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/container"

	pbImplStorageV2Shared "github.com/arangodb/kube-arangodb/integrations/storage/v2/shared"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

func (i *ios) List(ctx context.Context, key string) (util.NextIterator[[]pbImplStorageV2Shared.File], error) {
	return &listIterator{
		pager: i.container().NewListBlobsFlatPager(&container.ListBlobsFlatOptions{
			Include:    container.ListBlobsInclude{},
			Marker:     nil,
			MaxResults: i.config.MaxListKeys,
			Prefix:     util.NewType(i.key(key)),
		}),
		parent: i,
	}, nil
}

type listIterator struct {
	lock   sync.Mutex
	parent *ios

	pager *runtime.Pager[container.ListBlobsFlatResponse]
}

func (l *listIterator) Next(ctx context.Context) ([]pbImplStorageV2Shared.File, error) {
	l.lock.Lock()
	defer l.lock.Unlock()

	if !l.pager.More() {
		return nil, io.EOF
	}

	resp, err := l.pager.NextPage(ctx)
	if err != nil {
		return nil, err
	}

	if resp.Segment == nil {
		return nil, errors.Errorf("Invalid segment response")
	}

	data := make([]pbImplStorageV2Shared.File, len(resp.Segment.BlobItems))

	for id, file := range resp.Segment.BlobItems {
		if file == nil || file.Properties == nil {
			return nil, errors.Errorf("Invalid file response")
		}

		data[id] = pbImplStorageV2Shared.File{
			Key: l.parent.clean(*file.Name),
			Info: pbImplStorageV2Shared.Info{
				Size:          uint64(util.OptionalType(file.Properties.ContentLength, 0)),
				LastUpdatedAt: util.OptionalType(file.Properties.LastModified, time.Time{}),
			},
		}
	}

	return data, nil
}
