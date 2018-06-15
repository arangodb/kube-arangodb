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
	"time"

	driver "github.com/arangodb/go-driver"

	"github.com/arangodb/kube-arangodb/tests/duration/test"
)

// queryUpdateDocuments runs an AQL update query.
// The operation is expected to succeed.
func (t *simpleTest) queryUpdateDocuments(c *collection, key string) (string, error) {
	ctx := context.Background()
	ctx = driver.WithQueryCount(ctx)

	t.log.Info().Msgf("Creating update AQL query for collection '%s'...", c.name)
	newName := fmt.Sprintf("AQLUpdate name %s", time.Now())
	query := fmt.Sprintf("UPDATE \"%s\" WITH { name: \"%s\" } IN %s RETURN NEW", key, newName, c.name)
	cursor, err := t.db.Query(ctx, query, nil)
	if err != nil {
		// This is a failure
		t.queryUpdateCounter.failed++
		t.reportFailure(test.NewFailure("Failed to create update AQL cursor in collection '%s': %v", c.name, err))
		return "", maskAny(err)
	}
	var resultDocument UserDocument
	m, err := cursor.ReadDocument(ctx, &resultDocument)
	if err != nil {
		// This is a failure
		t.queryUpdateCounter.failed++
		t.reportFailure(test.NewFailure("Failed to read document from cursor in collection '%s': %v", c.name, err))
		return "", maskAny(err)
	}
	resultCount := cursor.Count()
	cursor.Close()
	if resultCount != 1 {
		// This is a failure
		t.queryUpdateCounter.failed++
		t.reportFailure(test.NewFailure("Failed to create update AQL cursor in collection '%s': expected 1 result, got %d", c.name, resultCount))
		return "", maskAny(fmt.Errorf("Number of documents was %d, expected 1", resultCount))
	}

	// Update document
	c.existingDocs[key] = resultDocument
	t.queryUpdateCounter.succeeded++
	t.log.Info().Msgf("Creating update AQL query for collection '%s' succeeded", c.name)

	return m.Rev, nil
}

// queryUpdateDocumentsLongRunning runs a long running AQL update query.
// The operation is expected to succeed.
func (t *simpleTest) queryUpdateDocumentsLongRunning(c *collection, key string) (string, error) {
	ctx := context.Background()
	ctx = driver.WithQueryCount(ctx)

	t.log.Info().Msgf("Creating long running update AQL query for collection '%s'...", c.name)
	newName := fmt.Sprintf("AQLLongRunningUpdate name %s", time.Now())
	query := fmt.Sprintf("UPDATE \"%s\" WITH { name: \"%s\", unknown: SLEEP(15) } IN %s RETURN NEW", key, newName, c.name)
	cursor, err := t.db.Query(ctx, query, nil)
	if err != nil {
		// This is a failure
		t.queryUpdateLongRunningCounter.failed++
		t.reportFailure(test.NewFailure("Failed to create long running update AQL cursor in collection '%s': %v", c.name, err))
		return "", maskAny(err)
	}
	var resultDocument UserDocument
	m, err := cursor.ReadDocument(ctx, &resultDocument)
	if err != nil {
		// This is a failure
		t.queryUpdateCounter.failed++
		t.reportFailure(test.NewFailure("Failed to read document from cursor in collection '%s': %v", c.name, err))
		return "", maskAny(err)
	}
	resultCount := cursor.Count()
	cursor.Close()
	if resultCount != 1 {
		// This is a failure
		t.queryUpdateLongRunningCounter.failed++
		t.reportFailure(test.NewFailure("Failed to create long running update AQL cursor in collection '%s': expected 1 result, got %d", c.name, resultCount))
		return "", maskAny(fmt.Errorf("Number of documents was %d, expected 1", resultCount))
	}

	// Update document
	c.existingDocs[key] = resultDocument
	t.queryUpdateLongRunningCounter.succeeded++
	t.log.Info().Msgf("Creating long running update AQL query for collection '%s' succeeded", c.name)

	return m.Rev, nil
}
