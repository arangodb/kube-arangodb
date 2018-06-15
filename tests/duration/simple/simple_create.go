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

	"github.com/arangodb/kube-arangodb/tests/duration/test"
)

// createDocument creates a new document.
// The operation is expected to succeed.
func (t *simpleTest) createDocument(c *collection, document interface{}, key string) (string, error) {
	ctx := context.Background()
	col, err := t.db.Collection(ctx, c.name)
	if err != nil {
		return "", maskAny(err)
	}
	t.log.Info().Msgf("Creating document '%s' in '%s'...", key, c.name)
	m, err := col.CreateDocument(ctx, document)
	if err != nil {
		// This is a failure
		t.createCounter.failed++
		t.reportFailure(test.NewFailure("Failed to create document with key '%s' in collection '%s': %v", key, c.name, err))
		return "", maskAny(err)
	}
	t.createCounter.succeeded++
	t.log.Info().Msgf("Creating document '%s' in '%s' succeeded", key, c.name)
	return m.Rev, nil
}
