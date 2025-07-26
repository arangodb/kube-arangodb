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
	"fmt"

	"github.com/rs/zerolog"

	platformApi "github.com/arangodb/kube-arangodb/pkg/apis/platform/v1beta1"
	"github.com/arangodb/kube-arangodb/pkg/debug_package/shared"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
)

func arangoPlatformV1beta1ArangoPlatformChartExtract(ctx context.Context, logger zerolog.Logger, client kclient.Client, files chan<- shared.File, item *platformApi.ArangoPlatformChart) error {
	if i := item.Status.Info; i != nil {
		if i.Valid {
			if d := i.Details; d != nil {
				files <- shared.NewFile(fmt.Sprintf("%s-%s.tgz", d.Name, d.Version), func() ([]byte, error) {
					return i.Definition, nil
				})
			}
		}
	}

	return nil
}
