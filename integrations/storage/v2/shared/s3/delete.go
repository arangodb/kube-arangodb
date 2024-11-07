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

	"github.com/arangodb/kube-arangodb/pkg/util"
)

func (i *ios) Delete(ctx context.Context, key string) (bool, error) {
	_, err := i.client.DeleteObjectWithContext(ctx, &s3.DeleteObjectInput{
		Key:    util.NewType(i.key(key)),
		Bucket: util.NewType(i.config.BucketName),
	})
	if err != nil {
		if IsAWSNotFoundError(err) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}
