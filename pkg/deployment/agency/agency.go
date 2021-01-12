//
// DISCLAIMER
//
// Copyright 2020 ArangoDB GmbH, Cologne, Germany
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
// Author Adam Janikowski
//

package agency

import (
	"context"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"

	"github.com/arangodb/go-driver/agency"
)

type Fetcher func(ctx context.Context, i interface{}, keyParts ...string) error

func NewFetcher(a agency.Agency) Fetcher {
	return func(ctx context.Context, i interface{}, keyParts ...string) error {
		if err := a.ReadKey(ctx, keyParts, i); err != nil {
			return errors.WithStack(err)
		}

		return nil
	}
}
