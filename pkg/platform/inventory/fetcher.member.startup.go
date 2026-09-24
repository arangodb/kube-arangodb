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

package inventory

import (
	"context"
	_ "embed"
	goHttp "net/http"

	adbDriverV2 "github.com/arangodb/go-driver/v2/arangodb"
	adbDriverV2Connection "github.com/arangodb/go-driver/v2/connection"

	"github.com/arangodb/kube-arangodb/pkg/logging"
	"github.com/arangodb/kube-arangodb/pkg/util/arangod"
	"github.com/arangodb/kube-arangodb/pkg/util/executor"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
)

//go:embed queries/startup.aql
var queryStartupAQL string

// The member startup fetcher enumerates the deployment members - all DBServers and Coordinators in
// cluster mode, or the single server in single mode - and joins each to its latest collector startup
// event (matched on the serverID dimension), emitting the startup values as ARANGO_MEMBER_STARTUP
// inventory items. The serverID and nodeName are emitted SHA256-hashed (the raw member/node ids never
// leave the deployment); the pod UID and boot id are emitted as-is. It is a no-op when the _events
// collection is absent (collector disabled).
func init() {
	global.MustRegister("member.startup", func(conn adbDriverV2Connection.Connection, cfg *Configuration, out chan<- *Item) executor.RunFunc {
		return func(ctx context.Context, log logging.Logger, t executor.Thread, h executor.Handler) error {
			members, err := startupMembers(ctx, conn)
			if err != nil {
				return err
			}

			if len(members) == 0 {
				log.Info("No members to collect startup values for")
				return nil
			}

			return ExecuteBasicAQLIfExists("_system", "_events", queryStartupAQL, map[string]any{
				"members": members,
			})(conn, cfg, out)(ctx, log, t, h)
		}
	})
}

// startupMember identifies a deployment member whose startup values should be reported. It carries the
// arangod server id (matching the serverID dimension the collector tags its startup events with) and
// the member role from cluster health, so an item can be emitted even when no startup event exists yet.
type startupMember struct {
	ID   string `json:"id"`
	Role string `json:"role"`
}

// startupMembers returns the members whose startup values should be reported: all DBServers and
// Coordinators in cluster mode, or the single server in single mode.
func startupMembers(ctx context.Context, conn adbDriverV2Connection.Connection) ([]startupMember, error) {
	health, err := arangod.GetRequestWithTimeout[adbDriverV2.ClusterHealth](ctx, globals.GetGlobals().Timeouts().ArangoD().Get(), conn, "_admin", "cluster", "health").
		Do(ctx).
		AcceptCode(goHttp.StatusOK).
		Response()
	if err == nil {
		var members []startupMember
		for id, member := range health.Health {
			switch member.Role {
			case adbDriverV2.ServerRoleDBServer, adbDriverV2.ServerRoleCoordinator:
				members = append(members, startupMember{ID: string(id), Role: string(member.Role)})
			}
		}
		return members, nil
	}

	if c, ok := arangod.IsInvalidCode(err); ok && c.Got == goHttp.StatusForbidden {
		// Single server: the cluster health endpoint is forbidden. Use the single server id.
		id, err := adbDriverV2.NewClient(conn).ServerID(ctx)
		if err != nil || id == "" {
			// Best-effort: without a server id there is nothing to join on.
			return nil, nil
		}
		return []startupMember{{ID: id, Role: string(adbDriverV2.ServerRoleSingle)}}, nil
	}

	return nil, err
}
