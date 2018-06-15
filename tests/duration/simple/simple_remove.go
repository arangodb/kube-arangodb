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

	driver "github.com/arangodb/go-driver"

	"github.com/arangodb/kube-arangodb/tests/duration/test"
)

// removeExistingDocument removes an existing document with an optional explicit revision.
// The operation is expected to succeed.
func (t *simpleTest) removeExistingDocument(collectionName string, key, rev string) error {
	ctx := context.Background()
	col, err := t.db.Collection(ctx, collectionName)
	if err != nil {
		return maskAny(err)
	}
	t.log.Info().Msgf("Removing existing document '%s' from '%s'...", key, collectionName)
	if _, err := col.RemoveDocument(ctx, key); err != nil {
		// This is a failure
		t.deleteExistingCounter.failed++
		t.reportFailure(test.NewFailure("Failed to delete existing document '%s' in collection '%s': %v", key, collectionName, err))
		return maskAny(err)
	}
	t.deleteExistingCounter.succeeded++
	t.log.Info().Msgf("Removing existing document '%s' from '%s' succeeded", key, collectionName)
	return nil
}

// removeNonExistingDocument removes a non-existing document.
// The operation is expected to fail.
func (t *simpleTest) removeNonExistingDocument(collectionName string, key string) error {
	ctx := context.Background()
	col, err := t.db.Collection(ctx, collectionName)
	if err != nil {
		return maskAny(err)
	}
	t.log.Info().Msgf("Removing non-existing document '%s' from '%s'...", key, collectionName)
	if _, err := col.RemoveDocument(ctx, key); !driver.IsNotFound(err) {
		// This is a failure
		t.deleteNonExistingCounter.failed++
		t.reportFailure(test.NewFailure("Failed to delete non-existing document '%s' in collection '%s': %v", key, collectionName, err))
		return maskAny(err)
	}
	t.deleteNonExistingCounter.succeeded++
	t.log.Info().Msgf("Removing non-existing document '%s' from '%s' succeeded", key, collectionName)
	return nil
}
