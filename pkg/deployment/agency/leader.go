//
// DISCLAIMER
//
// Copyright 2023 ArangoDB GmbH, Cologne, Germany
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

package agency

import (
	"context"

	agencyCache "github.com/arangodb-helper/go-helper/pkg/arangod/agency/cache"
	"github.com/arangodb-helper/go-helper/pkg/arangod/conn"
)

func StaticLeaderDiscovery(in conn.Connection) agencyCache.LeaderDiscovery {
	return staticLeaderDiscovery{conn: in}
}

type staticLeaderDiscovery struct {
	conn conn.Connection
}

func (s staticLeaderDiscovery) Discover(ctx context.Context) (conn.Connection, error) {
	return s.conn, nil
}
