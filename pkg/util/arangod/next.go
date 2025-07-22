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

package arangod

import (
	"context"
	"io"

	"github.com/arangodb/go-driver/v2/arangodb"
	"github.com/arangodb/go-driver/v2/arangodb/shared"

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

func QueryV2NextIterator[T any](in arangodb.Cursor, batch int) util.NextIterator[[]T] {
	return queryV1NextIterator[T]{in: in, batch: batch}
}

type queryV1NextIterator[T any] struct {
	in    arangodb.Cursor
	batch int
}

func (q queryV1NextIterator[T]) Next(ctx context.Context) ([]T, error) {
	if q.batch <= 0 {
		return nil, errors.Errorf("batch out of range")
	}

	r := make([]T, 0, q.batch)

	for {
		var o T

		if _, err := q.in.ReadDocument(ctx, &o); err != nil {
			if _, ok := errors.ExtractCause[shared.NoMoreDocumentsError](err); ok {
				break
			}

			return nil, err
		}

		r = append(r, o)

		if len(r) >= q.batch {
			break
		}
	}

	if len(r) > 0 {
		return r, nil
	}

	return nil, io.EOF
}
