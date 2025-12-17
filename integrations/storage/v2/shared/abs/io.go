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
	"path"
	goStrings "strings"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/container"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/service"
)

type ios struct {
	config Configuration
	client *service.Client
}

func (i *ios) container() *container.Client {
	return i.client.NewContainerClient(i.config.BucketName)
}

func (i *ios) clean(key string) string {
	return goStrings.TrimPrefix(goStrings.TrimPrefix(key, i.key()), "/")
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
