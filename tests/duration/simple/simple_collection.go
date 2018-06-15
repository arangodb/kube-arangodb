//
// DISCLAIMER
//
// Copyright 2018 ArangoDB GmbH, Cologne, Germany
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

package simple

import (
	"context"
	"fmt"

	driver "github.com/arangodb/go-driver"

	"github.com/arangodb/kube-arangodb/tests/duration/test"
)

// createCollection creates a new collection.
// The operation is expected to succeed.
func (t *simpleTest) createCollection(c *collection, numberOfShards, replicationFactor int) error {
	ctx := context.Background()
	opts := &driver.CreateCollectionOptions{
		NumberOfShards:    numberOfShards,
		ReplicationFactor: replicationFactor,
	}
	t.log.Info().Msgf("Creating collection '%s' with numberOfShards=%d, replicationFactor=%d...", c.name, numberOfShards, replicationFactor)
	if _, err := t.db.CreateCollection(ctx, c.name, opts); err != nil {
		// This is a failure
		t.reportFailure(test.NewFailure("Failed to create collection '%s': %v", c.name, err))
		return maskAny(err)
	} else if driver.IsConflict(err) {
		// Duplicate name, check if that is correct
		if exists, checkErr := t.collectionExists(c); checkErr != nil {
			t.log.Error().Msgf("Failed to check if collection exists: %v", checkErr)
			t.reportFailure(test.NewFailure("Failed to create collection '%s': %v and cannot check existance: %v", c.name, err, checkErr))
			return maskAny(err)
		} else if !exists {
			// Collection has not been created, so 409 status is really wrong
			t.reportFailure(test.NewFailure("Failed to create collection '%s': 409 reported but collection does not exist", c.name))
			return maskAny(fmt.Errorf("Create collection reported 409, but collection does not exist"))
		}
	}
	t.log.Info().Msgf("Creating collection '%s' with numberOfShards=%d, replicationFactor=%d succeeded", c.name, numberOfShards, replicationFactor)
	return nil
}

// removeCollection remove an existing collection.
// The operation is expected to succeed.
func (t *simpleTest) removeExistingCollection(c *collection) error {
	ctx := context.Background()
	t.log.Info().Msgf("Removing collection '%s'...", c.name)
	col, err := t.db.Collection(ctx, c.name)
	if err != nil {
		return maskAny(err)
	}
	if err := col.Remove(ctx); err != nil {
		// This is a failure
		t.removeExistingCollectionCounter.failed++
		t.reportFailure(test.NewFailure("Failed to remove collection '%s': %v", c.name, err))
		return maskAny(err)
	}
	t.removeExistingCollectionCounter.succeeded++
	t.log.Info().Msgf("Removing collection '%s' succeeded", c.name)
	t.unregisterCollection(c)
	return nil
}

// collectionExists tries to fetch information about the collection to see if it exists.
func (t *simpleTest) collectionExists(c *collection) (bool, error) {
	ctx := context.Background()
	t.log.Info().Msgf("Checking collection '%s'...", c.name)
	if found, err := t.db.CollectionExists(ctx, c.name); err != nil {
		// This is a failure
		return false, maskAny(err)
	} else {
		return found, nil
	}
}
