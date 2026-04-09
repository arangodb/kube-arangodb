//
// DISCLAIMER
//
// Copyright 2023-2026 ArangoDB GmbH, Cologne, Germany
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

package leader

import (
	"context"
	"time"

	adbDriverV2Connection "github.com/arangodb/go-driver/v2/connection"
)

type Discovery interface {
	Discover(ctx context.Context) (adbDriverV2Connection.Connection, error)
}

type StateLoader[T interface{}] interface {
	State() (*T, uint64, bool)

	Invalidate()
	Valid() bool

	UpdateTime() time.Time

	Refresh(ctx context.Context, discovery Discovery) error
}

func StaticLeaderDiscovery(in adbDriverV2Connection.Connection) Discovery {
	return staticLeaderDiscovery{conn: in}
}

type staticLeaderDiscovery struct {
	conn adbDriverV2Connection.Connection
}

func (s staticLeaderDiscovery) Discover(ctx context.Context) (adbDriverV2Connection.Connection, error) {
	return s.conn, nil
}
