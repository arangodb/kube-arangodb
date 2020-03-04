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
// Author Ewout Prangsma
//

package arangod

import (
	"context"
	"fmt"

	driver "github.com/arangodb/go-driver"
	"github.com/pkg/errors"
)

// IsDBServerEmpty checks if the dbserver identified by the given ID no longer has any
// data on it.
// The given driver must have all coordinators as endpoints.
// The functions returns an error when the check could not be completed or the dbserver
// is not empty, or nil when the dbserver is found to be empty.
func IsDBServerEmpty(ctx context.Context, id string, client driver.Client) error {
	c, err := client.Cluster(ctx)
	if err != nil {
		return maskAny(errors.Wrapf(err, "Cannot obtain Cluster"))
	}
	dbs, err := client.Databases(ctx)
	if err != nil {
		return maskAny(errors.Wrapf(err, "Cannot fetch databases"))
	}
	for _, db := range dbs {
		inventory, err := c.DatabaseInventory(ctx, db)
		if err != nil {
			return maskAny(errors.Wrapf(err, "Cannot fetch inventory for %s", db.Name()))
		}
		// Go over all collections
		for _, col := range inventory.Collections {
			// Go over all shards of the collection
			for shardID, serverIDs := range col.Parameters.Shards {
				for _, serverID := range serverIDs {
					if string(serverID) == id {
						// DBServer still used in this shard
						return maskAny(fmt.Errorf("DBServer still used in shard %s of %s.%s", shardID, col.Parameters.Name, db.Name()))
					}
				}
			}
		}
	}
	// DBServer is not used in any shard of any database
	return nil
}
