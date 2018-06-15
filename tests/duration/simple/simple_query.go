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

// queryDocumentsLongRunning runs a long running AQL query.
// The operation is expected to succeed.
func (t *simpleTest) queryDocumentsLongRunning(c *collection) error {
	if len(c.existingDocs) < 10 {
		t.log.Info().Msgf("Skipping query test, we need 10 or more documents")
		return nil
	}

	ctx := context.Background()
	ctx = driver.WithQueryCount(ctx)

	t.log.Info().Msgf("Creating long running AQL query for '%s'...", c.name)
	query := fmt.Sprintf("FOR d IN %s LIMIT 10 RETURN {d:d, s:SLEEP(2)}", c.name)
	cursor, err := t.db.Query(ctx, query, nil)
	if err != nil {
		// This is a failure
		t.queryLongRunningCounter.failed++
		t.reportFailure(test.NewFailure("Failed to create long running AQL cursor in collection '%s': %v", c.name, err))
		return maskAny(err)
	}
	cursor.Close()
	resultCount := cursor.Count()
	t.queryLongRunningCounter.succeeded++
	t.log.Info().Msgf("Creating long running AQL query for collection '%s' succeeded", c.name)

	// We should've fetched all documents, check result count
	if resultCount != 10 {
		t.reportFailure(test.NewFailure("Number of documents was %d, expected 10", resultCount))
		return maskAny(fmt.Errorf("Number of documents was %d, expected 10", resultCount))
	}

	return nil
}
