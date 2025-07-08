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

package arango

import (
	"context"
	"io"

	"github.com/rs/zerolog"

	pbImplStorageV2 "github.com/arangodb/kube-arangodb/integrations/storage/v2"
	platformApi "github.com/arangodb/kube-arangodb/pkg/apis/platform/v1alpha1"
	"github.com/arangodb/kube-arangodb/pkg/debug_package/cli"
	"github.com/arangodb/kube-arangodb/pkg/debug_package/shared"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
)

func arangoPlatformV1Alpha1ArangoPlatformStorageDebug(ctx context.Context, logger zerolog.Logger, client kclient.Client, files chan<- shared.File, item *platformApi.ArangoPlatformStorage) error {
	if !cli.GetInput().DebugPackageFiles {
		return nil
	}

	c, err := pbImplStorageV2.NewIOFromObject(ctx, client, item)
	if err != nil {
		return err
	}

	allFiles, err := c.List(ctx, "debug/")
	if err != nil {
		return err
	}

	for {
		t, err := allFiles.Next(ctx)
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		for _, f := range t {
			f := f
			files <- shared.NewFile(f.Key, func() ([]byte, error) {
				reader, err := c.Read(ctx, f.Key)
				if err != nil {
					return nil, err
				}

				data, err := io.ReadAll(reader)
				if err != nil {
					return nil, err
				}

				return data, nil
			})
		}
	}

	return nil
}
