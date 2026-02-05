//
// DISCLAIMER
//
// Copyright 2026 ArangoDB GmbH, Cologne, Germany
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

package authorization

import (
	"context"
	"fmt"
	"io"

	"google.golang.org/protobuf/proto"

	"github.com/arangodb/go-driver/v2/arangodb"
	"github.com/arangodb/go-driver/v2/arangodb/shared"

	ugrpc "github.com/arangodb/kube-arangodb/pkg/util/grpc"
)

// PoolChanges pools the changes from registry. If no documents found EOF is returned
func PoolChanges[T proto.Message](ctx context.Context, db arangodb.DatabaseQuery, col string, start int, max int) ([]T, error) {
	query := fmt.Sprintf("FOR doc IN %s FILTER doc.sequence > %%start SORT BY doc.sequence ASC LIMIT %%limit", col)

	result, err := db.Query(ctx, query, &arangodb.QueryOptions{
		BatchSize: 1024,
		BindVars:  map[string]interface{}{"start": start, "max": max},
	})
	if err != nil {
		return nil, err
	}

	ret := make([]T, 0, max)

	for {
		var d ugrpc.Object[T]

		if _, err := result.ReadDocument(ctx, &d); err != nil {
			if shared.IsEOF(err) {
				break
			}

			return nil, err
		}

		ret = append(ret, d.Object)
	}

	if len(ret) == 0 {
		return nil, io.EOF
	}

	return ret, nil
}
