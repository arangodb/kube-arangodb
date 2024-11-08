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

func (i *ios) Init(ctx context.Context, opts *pbImplStorageV2Shared.InitOptions) error {
	if opts.GetCreate() {
		if _, err := i.client.HeadBucketWithContext(ctx, &s3.HeadBucketInput{
			Bucket: util.NewType(i.config.BucketName),
		}); err != nil {
			if !IsAWSNotFoundError(err) {
				return err
			}

			if _, err := i.client.CreateBucketWithContext(ctx, &s3.CreateBucketInput{
				Bucket: util.NewType(i.config.BucketName),
			}); err != nil {
				return err
			}
		}
	}

	return nil
}
