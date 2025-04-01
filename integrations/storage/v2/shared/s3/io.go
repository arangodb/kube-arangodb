//
// DISCLAIMER
//
// Copyright 2024-2025 ArangoDB GmbH, Cologne, Germany
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
	"path"
	goStrings "strings"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"

	pbImplStorageV2Shared "github.com/arangodb/kube-arangodb/integrations/storage/v2/shared"
	"github.com/arangodb/kube-arangodb/pkg/util"
)

type ios struct {
	config     Configuration
	client     s3iface.S3API
	uploader   *s3manager.Uploader
	downloader *s3manager.Downloader
}

func (i *ios) key(keys ...string) string {
	return path.Join(goStrings.TrimPrefix(i.config.BucketPrefix, "/"), path.Join(keys...))
}

func (i *ios) clean(key string) string {
	return goStrings.TrimPrefix(goStrings.TrimPrefix(key, i.key()), "/")
}

func (i *ios) Write(ctx context.Context, key string) (pbImplStorageV2Shared.Writer, error) {
	w := newWriter(i)

	w.start(ctx, &s3manager.UploadInput{
		Bucket: util.NewType(i.config.BucketName),
		Key:    util.NewType(i.key(key)),
	})

	return w, nil
}

func (i *ios) Read(ctx context.Context, key string) (pbImplStorageV2Shared.Reader, error) {
	r := newReader(i)

	r.start(ctx, &s3.GetObjectInput{
		Bucket: util.NewType(i.config.BucketName),
		Key:    util.NewType(i.key(key)),
	})

	return r, nil
}
