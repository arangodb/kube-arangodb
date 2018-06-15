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

import "context"

// isDocumentEqualTo reads an existing document and checks that it is equal to the given document.
// Returns: (isEqual,currentRevision,error)
func (t *simpleTest) isDocumentEqualTo(c *collection, key string, expected UserDocument) (bool, string, error) {
	ctx := context.Background()
	var result UserDocument
	t.log.Info().Msgf("Checking existing document '%s' from '%s'...", key, c.name)
	col, err := t.db.Collection(ctx, c.name)
	if err != nil {
		return false, "", maskAny(err)
	}
	m, err := col.ReadDocument(ctx, key, &result)
	if err != nil {
		// This is a failure
		t.log.Error().Msgf("Failed to read document '%s' from '%s': %v", key, c.name, err)
		return false, "", maskAny(err)
	}
	// Compare document against expected document
	if result.Equals(expected) {
		// Found an exact match
		return true, m.Rev, nil
	}
	t.log.Info().Msgf("Document '%s' in '%s'  returned different values: got %q expected %q", key, c.name, result, expected)
	return false, m.Rev, nil
}
