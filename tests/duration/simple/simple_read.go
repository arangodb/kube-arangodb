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

// readExistingDocument reads an existing document with an optional explicit revision.
// The operation is expected to succeed.
func (t *simpleTest) readExistingDocument(c *collection, key, rev string, updateRevision, skipExpectedValueCheck bool) (string, error) {
	ctx := context.Background()
	var result UserDocument
	col, err := t.db.Collection(ctx, c.name)
	if err != nil {
		return "", maskAny(err)
	}
	m, err := col.ReadDocument(ctx, key, &result)
	if err != nil {
		// This is a failure
		t.readExistingCounter.failed++
		t.reportFailure(test.NewFailure("Failed to read existing document '%s' in collection '%s': %v", key, c.name, err))
		return "", maskAny(err)
	}
	// Compare document against expected document
	if !skipExpectedValueCheck {
		expected := c.existingDocs[key]
		if result.Value != expected.Value || result.Name != expected.Name || result.Odd != expected.Odd {
			// This is a failure
			t.readExistingCounter.failed++
			t.reportFailure(test.NewFailure("Read existing document '%s' returned different values '%s': got %q expected %q", key, c.name, result, expected))
			return "", maskAny(fmt.Errorf("Read returned invalid values"))
		}
	}
	if updateRevision {
		// Store read document so we have the last revision
		c.existingDocs[key] = result
	}
	t.readExistingCounter.succeeded++
	t.log.Info().Msgf("Reading existing document '%s' from '%s' succeeded", key, c.name)
	return m.Rev, nil
}

// readNonExistingDocument reads a non-existing document.
// The operation is expected to fail.
func (t *simpleTest) readNonExistingDocument(collectionName string, key string) error {
	ctx := context.Background()
	var result UserDocument
	t.log.Info().Msgf("Reading non-existing document '%s' from '%s'...", key, collectionName)
	col, err := t.db.Collection(ctx, collectionName)
	if err != nil {
		return maskAny(err)
	}
	if _, err := col.ReadDocument(ctx, key, &result); !driver.IsNotFound(err) {
		// This is a failure
		t.readNonExistingCounter.failed++
		t.reportFailure(test.NewFailure("Failed to read non-existing document '%s' in collection '%s': %v", key, collectionName, err))
		return maskAny(err)
	}
	t.readNonExistingCounter.succeeded++
	t.log.Info().Msgf("Reading non-existing document '%s' from '%s' succeeded", key, collectionName)
	return nil
}
