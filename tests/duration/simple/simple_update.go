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

package simple

import (
	"context"
	"fmt"
	"time"

	driver "github.com/arangodb/go-driver"

	"github.com/arangodb/kube-arangodb/tests/duration/test"
)

// updateExistingDocument updates an existing document with an optional explicit revision.
// The operation is expected to succeed.
func (t *simpleTest) updateExistingDocument(c *collection, key string) (string, error) {
	ctx := context.Background()
	col, err := t.db.Collection(ctx, c.name)
	if err != nil {
		return "", maskAny(err)
	}
	newName := fmt.Sprintf("Updated name %s", time.Now())
	t.log.Info().Msgf("Updating existing document '%s' in '%s' (name -> '%s')...", key, c.name, newName)
	delta := map[string]interface{}{
		"name": newName,
	}
	doc := c.existingDocs[key]
	m, err := col.UpdateDocument(ctx, key, delta)
	if err != nil {
		// This is a failure
		t.updateExistingCounter.failed++
		t.reportFailure(test.NewFailure("Failed to update existing document '%s' in collection '%s': %v", key, c.name, err))
		return "", maskAny(err)
	}
	// Update internal doc
	doc.Name = newName
	doc.rev = m.Rev
	c.existingDocs[key] = doc
	t.updateExistingCounter.succeeded++
	t.log.Info().Msgf("Updating existing document '%s' in '%s' (name -> '%s') succeeded", key, c.name, newName)
	return m.Rev, nil
}

// updateNonExistingDocument updates a non-existing document.
// The operation is expected to fail.
func (t *simpleTest) updateNonExistingDocument(collectionName string, key string) error {
	ctx := context.Background()
	col, err := t.db.Collection(ctx, collectionName)
	if err != nil {
		return maskAny(err)
	}
	newName := fmt.Sprintf("Updated non-existing name %s", time.Now())
	t.log.Info().Msgf("Updating non-existing document '%s' in '%s' (name -> '%s')...", key, collectionName, newName)
	delta := map[string]interface{}{
		"name": newName,
	}
	if _, err := col.UpdateDocument(ctx, key, delta); !driver.IsNotFound(err) {
		// This is a failure
		t.updateNonExistingCounter.failed++
		t.reportFailure(test.NewFailure("Failed to update non-existing document '%s' in collection '%s': %v", key, collectionName, err))
		return maskAny(err)
	}
	t.updateNonExistingCounter.succeeded++
	t.log.Info().Msgf("Updating non-existing document '%s' in '%s' (name -> '%s') succeeded", key, collectionName, newName)
	return nil
}
