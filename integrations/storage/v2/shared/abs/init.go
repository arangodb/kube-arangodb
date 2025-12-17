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

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/container"

	pbImplStorageV2Shared "github.com/arangodb/kube-arangodb/integrations/storage/v2/shared"
)

func (i *ios) Init(ctx context.Context, opts *pbImplStorageV2Shared.InitOptions) error {
	c := i.container()

	_, err := c.GetProperties(ctx, &container.GetPropertiesOptions{})
	if err != nil {
		return err
	}

	return nil
}
