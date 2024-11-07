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

	"github.com/aws/aws-sdk-go/service/s3"

	pbImplStorageV2Shared "github.com/arangodb/kube-arangodb/integrations/storage/v2/shared"
	"github.com/arangodb/kube-arangodb/pkg/util"
)

func (i *ios) Head(ctx context.Context, key string) (*pbImplStorageV2Shared.Info, error) {
	obj, err := i.client.HeadObjectWithContext(ctx, &s3.HeadObjectInput{
		Bucket: util.NewType(i.config.BucketName),
		Key:    util.NewType(i.key(key)),
	})
	if err != nil {
		if IsAWSNotFoundError(err) {
			return nil, nil
		}
		return nil, err
	}

	return &pbImplStorageV2Shared.Info{
		Size:          uint64(util.TypeOrDefault(obj.ContentLength)),
		LastUpdatedAt: util.TypeOrDefault(obj.LastModified),
	}, nil
}
