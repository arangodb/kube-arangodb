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
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"

	pbImplStorageV2Shared "github.com/arangodb/kube-arangodb/integrations/storage/v2/shared"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

func (i *ios) Head(ctx context.Context, key string) (*pbImplStorageV2Shared.Info, error) {
	q := i.container().NewBlockBlobClient(i.key(key))

	prop, err := q.GetProperties(ctx, nil)
	if err != nil {
		var respErr *azcore.ResponseError
		if errors.As(err, &respErr) && respErr.StatusCode == 404 {
			return nil, nil
		}
		return nil, err
	}

	return &pbImplStorageV2Shared.Info{
		Size:          uint64(util.OptionalType(prop.ContentLength, 0)),
		LastUpdatedAt: util.OptionalType(prop.LastModified, time.Time{}),
	}, nil
}
