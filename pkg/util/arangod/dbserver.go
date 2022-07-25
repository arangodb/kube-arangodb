//
// DISCLAIMER
//
// Copyright 2016-2022 ArangoDB GmbH, Cologne, Germany
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
	"net/http"

	driver "github.com/arangodb/go-driver"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
)

// IsDBServerEmpty checks if the dbserver identified by the given ID no longer has any
// data on it.
// The given driver must have all coordinators as endpoints.
// The functions returns an error when the check could not be completed or the dbserver
// is not empty, or nil when the dbserver is found to be empty.
func IsDBServerEmpty(ctx context.Context, id string, client driver.Client) error {
	ctxChild, cancel := globals.GetGlobalTimeouts().ArangoD().WithTimeout(ctx)
	defer cancel()
	c, err := client.Cluster(ctxChild)
	if err != nil {
		return errors.WithStack(errors.Wrapf(err, "Cannot obtain Cluster"))
	}

	ctxChild, cancel = globals.GetGlobalTimeouts().ArangoD().WithTimeout(ctx)
	defer cancel()
	dbs, err := client.Databases(ctxChild)
	if err != nil {
		return errors.WithStack(errors.Wrapf(err, "Cannot fetch databases"))
	}

	var inventory driver.DatabaseInventory
	for _, db := range dbs {
		err := globals.GetGlobalTimeouts().ArangoD().RunWithTimeout(ctx, func(ctxChild context.Context) error {
			var err error
			inventory, err = c.DatabaseInventory(ctxChild, db)

			return err
		})
		if err != nil {
			return errors.WithStack(errors.Wrapf(err, "Cannot fetch inventory for %s", db.Name()))
		}
		// Go over all collections
		for _, col := range inventory.Collections {
			// Go over all shards of the collection
			for shardID, serverIDs := range col.Parameters.Shards {
				for _, serverID := range serverIDs {
					if string(serverID) == id {
						// DBServer still used in this shard
						return errors.WithStack(errors.Newf("DBServer still used in shard %s of %s.%s", shardID, col.Parameters.Name, db.Name()))
					}
				}
			}
		}
	}
	// DBServer is not used in any shard of any database
	return nil
}

// IsServerAvailable returns true when server is available.
// In active fail-over mode one of the server should be available.
func IsServerAvailable(ctx context.Context, c driver.Client) (bool, error) {
	req, err := c.Connection().NewRequest("GET", "_admin/server/availability")
	if err != nil {
		return false, errors.WithStack(err)
	}

	resp, err := c.Connection().Do(ctx, req)
	if err != nil {
		return false, errors.WithStack(err)
	}

	if err := resp.CheckStatus(http.StatusOK, http.StatusServiceUnavailable); err != nil {
		return false, errors.WithStack(err)
	}

	return resp.StatusCode() == http.StatusOK, nil
}
