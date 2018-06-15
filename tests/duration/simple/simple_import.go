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
	"bytes"
	"context"
	"fmt"

	"github.com/arangodb/kube-arangodb/tests/duration/test"
)

// createImportDocument creates a #document based import file.
func (t *simpleTest) createImportDocument() ([]byte, []UserDocument) {
	buf := &bytes.Buffer{}
	docs := make([]UserDocument, 0, 10000)
	fmt.Fprintf(buf, `[ "_key", "value", "name", "odd" ]`)
	fmt.Fprintln(buf)
	for i := 0; i < 10000; i++ {
		key := fmt.Sprintf("docimp%05d", i)
		userDoc := UserDocument{
			Key:   key,
			Value: i,
			Name:  fmt.Sprintf("Imported %d", i),
			Odd:   i%2 == 0,
		}
		docs = append(docs, userDoc)
		fmt.Fprintf(buf, `[ "%s", %d, "%s", %v ]`, userDoc.Key, userDoc.Value, userDoc.Name, userDoc.Odd)
		fmt.Fprintln(buf)
	}
	return buf.Bytes(), docs
}

// importDocuments imports a bulk set of documents.
// The operation is expected to succeed.
func (t *simpleTest) importDocuments(c *collection) error {
	ctx := context.Background()
	col, err := t.db.Collection(ctx, c.name)
	if err != nil {
		return maskAny(err)
	}
	_, docs := t.createImportDocument()
	t.log.Info().Msgf("Importing %d documents ('%s' - '%s') into '%s'...", len(docs), docs[0].Key, docs[len(docs)-1].Key, c.name)
	_, errs, err := col.CreateDocuments(ctx, docs)
	if err != nil {
		// This is a failure
		t.importCounter.failed++
		t.reportFailure(test.NewFailure("Failed to import documents in collection '%s': %v", c.name, err))
		return maskAny(err)
	}
	for i, d := range docs {
		if errs[i] == nil {
			c.existingDocs[d.Key] = d
		}
	}
	t.importCounter.succeeded++
	t.log.Info().Msgf("Importing %d documents ('%s' - '%s') into '%s' succeeded", len(docs), docs[0].Key, docs[len(docs)-1].Key, c.name)
	return nil
}
