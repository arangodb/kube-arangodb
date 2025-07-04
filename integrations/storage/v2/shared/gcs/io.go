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
	"crypto/sha256"
	"errors"
	"os"
	"path"
	goStrings "strings"

	"cloud.google.com/go/storage"

	pbImplStorageV2Shared "github.com/arangodb/kube-arangodb/integrations/storage/v2/shared"
)

type ios struct {
	config Configuration

	client *storage.Client
}

func (i *ios) key(keys ...string) string {
	out := path.Join(goStrings.TrimPrefix(i.config.BucketPrefix, "/"), path.Join(keys...))

	if len(keys) > 0 {
		if goStrings.HasSuffix(keys[len(keys)-1], "/") {
			out = out + "/"
		}
	}

	return out
}

func (i *ios) clean(key string) string {
	return goStrings.TrimPrefix(goStrings.TrimPrefix(key, i.key()), "/")
}

func (i *ios) Write(ctx context.Context, key string) (pbImplStorageV2Shared.Writer, error) {
	b := i.client.Bucket(i.config.BucketName)

	obj := b.Object(i.key(key))

	return &writer{
		write:    obj.NewWriter(ctx),
		checksum: sha256.New(),
	}, nil
}

func (i *ios) Read(ctx context.Context, key string) (pbImplStorageV2Shared.Reader, error) {
	b := i.client.Bucket(i.config.BucketName)

	obj := b.Object(i.key(key))

	r, err := obj.NewReader(ctx)
	if err != nil {
		if errors.Is(err, storage.ErrObjectNotExist) {
			return nil, os.ErrNotExist
		}
		return nil, err
	}

	return &reader{
		read:     r,
		checksum: sha256.New(),
	}, nil
}
