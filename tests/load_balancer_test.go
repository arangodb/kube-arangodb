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

package tests

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/dchest/uniuri"
	"github.com/stretchr/testify/assert"

	driver "github.com/arangodb/go-driver"
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/client"
	"github.com/arangodb/kube-arangodb/pkg/util"
)

type queryTest struct {
	Query             string
	BindVars          map[string]interface{}
	ExpectSuccess     bool
	ExpectedDocuments []interface{}
	DocumentType      reflect.Type
}

type queryTestContext struct {
	Context     context.Context
	ExpectCount bool
}

func TestLoadBalancingCursorVST(t *testing.T) {
	// run with VST
	TestLoadBalancingCursorSubtest(t, true)
}

func TestLoadBalancingCursorHTTP(t *testing.T) {
	// run with HTTP
	TestCursorLoadBalancingTestSubtest(t, false)
}

// TestLoadBalancerCursorVST tests cursor forwarding with load-balanced conn.
func TestLoadBalancingCursorSubtest(t *testing.T, useVst bool) {
	c := client.MustNewInCluster()
	kubecli := mustNewKubeClient(t)
	ns := getNamespace(t)

	// Prepare deployment config
	depl := newDeployment("test-lb-" + uniuri.NewLen(4))
	depl.Spec.Mode = api.NewMode(api.DeploymentModeCluster)

	// Create deployment
	_, err := c.DatabaseV1alpha().ArangoDeployments(ns).Create(depl)
	if err != nil {
		t.Fatalf("Create deployment failed: %v", err)
	}
	// Prepare cleanup
	defer removeDeployment(c, depl.GetName(), ns)

	// Wait for deployment to be ready
	apiObject, err := waitUntilDeployment(c, depl.GetName(), ns, deploymentIsReady())
	if err != nil {
		t.Fatalf("Deployment not running in time: %v", err)
	}

	// Create a database client
	ctx := context.Background()
	client := mustNewArangodDatabaseClient(ctx, kubecli, apiObject, t, useVst)

	// Wait for cluster to be available
	if err := waitUntilVersionUp(client, nil); err != nil {
		t.Fatalf("Cluster not running returning version in time: %v", err)
	}

	// Create data set
	collectionData := map[string][]interface{}{
		"books": []interface{}{
			Book{Title: "Book 01"},
			Book{Title: "Book 02"},
			Book{Title: "Book 03"},
			Book{Title: "Book 04"},
			Book{Title: "Book 05"},
			Book{Title: "Book 06"},
			Book{Title: "Book 07"},
			Book{Title: "Book 08"},
			Book{Title: "Book 09"},
			Book{Title: "Book 10"},
			Book{Title: "Book 11"},
			Book{Title: "Book 12"},
			Book{Title: "Book 13"},
			Book{Title: "Book 14"},
			Book{Title: "Book 15"},
			Book{Title: "Book 16"},
			Book{Title: "Book 17"},
			Book{Title: "Book 18"},
			Book{Title: "Book 19"},
			Book{Title: "Book 20"},
		},
		"users": []interface{}{
			UserDoc{Name: "John", Age: 13},
			UserDoc{Name: "Jake", Age: 25},
			UserDoc{Name: "Clair", Age: 12},
			UserDoc{Name: "Johnny", Age: 42},
			UserDoc{Name: "Blair", Age: 67},
			UserDoc{Name: "Zz", Age: 12},
		},
	}
	for colName, colDocs := range collectionData {
		col := ensureCollection(ctx, db, colName, nil, t)
		if _, _, err := col.CreateDocuments(ctx, colDocs); err != nil {
			t.Fatalf("Expected success, got %s", describe(err))
		}
	}

	// Setup tests
	tests := []queryTest{
		queryTest{
			Query:             "FOR d IN books SORT d.Title RETURN d",
			ExpectSuccess:     true,
			ExpectedDocuments: collectionData["books"],
			DocumentType:      reflect.TypeOf(Book{}),
		},
		queryTest{
			Query:             "FOR d IN books FILTER d.Title==@title SORT d.Title RETURN d",
			BindVars:          map[string]interface{}{"title": "Book 02"},
			ExpectSuccess:     true,
			ExpectedDocuments: []interface{}{collectionData["books"][1]},
			DocumentType:      reflect.TypeOf(Book{}),
		},
		queryTest{
			Query:         "FOR d IN books FILTER d.Title==@title SORT d.Title RETURN d",
			BindVars:      map[string]interface{}{"somethingelse": "Book 02"},
			ExpectSuccess: false, // Unknown `@title`
		},
		queryTest{
			Query:             "FOR u IN users FILTER u.age>100 SORT u.name RETURN u",
			ExpectSuccess:     true,
			ExpectedDocuments: []interface{}{},
			DocumentType:      reflect.TypeOf(UserDoc{}),
		},
		queryTest{
			Query:             "FOR u IN users FILTER u.age<@maxAge SORT u.name RETURN u",
			BindVars:          map[string]interface{}{"maxAge": 20},
			ExpectSuccess:     true,
			ExpectedDocuments: []interface{}{collectionData["users"][2], collectionData["users"][0], collectionData["users"][5]},
			DocumentType:      reflect.TypeOf(UserDoc{}),
		},
		queryTest{
			Query:         "FOR u IN users FILTER u.age<@maxAge SORT u.name RETURN u",
			BindVars:      map[string]interface{}{"maxage": 20},
			ExpectSuccess: false, // `@maxage` versus `@maxAge`
		},
		queryTest{
			Query:             "FOR u IN users SORT u.age RETURN u.age",
			ExpectedDocuments: []interface{}{12, 12, 13, 25, 42, 67},
			DocumentType:      reflect.TypeOf(12),
			ExpectSuccess:     true,
		},
		queryTest{
			Query:             "FOR p IN users COLLECT a = p.age WITH COUNT INTO c SORT a RETURN [a, c]",
			ExpectedDocuments: []interface{}{[]int{12, 2}, []int{13, 1}, []int{25, 1}, []int{42, 1}, []int{67, 1}},
			DocumentType:      reflect.TypeOf([]int{}),
			ExpectSuccess:     true,
		},
		queryTest{
			Query:             "FOR u IN users SORT u.name RETURN u.name",
			ExpectedDocuments: []interface{}{"Blair", "Clair", "Jake", "John", "Johnny", "Zz"},
			DocumentType:      reflect.TypeOf("foo"),
			ExpectSuccess:     true,
		},
	}

	// Setup context alternatives
	contexts := []queryTestContext{
		queryTestContext{driver.WithUseVST(nil, true), false},
		queryTestContext{driver.WithUseVST(context.Background(), true), false},
		queryTestContext{driver.WithUseVST(driver.WithQueryCount(nil), true), true},
		queryTestContext{driver.WithUseVST(driver.WithQueryCount(nil, true), true), true},
		queryTestContext{driver.WithUseVST(driver.WithQueryCount(nil, false), true), false},
		queryTestContext{driver.WithUseVST(driver.WithQueryBatchSize(nil, 1), true), false},
		queryTestContext{driver.WithUseVST(driver.WithQueryCache(nil), true), false},
		queryTestContext{driver.WithUseVST(driver.WithQueryCache(nil, true), true), false},
		queryTestContext{driver.WithUseVST(driver.WithQueryCache(nil, false), true), false},
		queryTestContext{driver.WithUseVST(driver.WithQueryMemoryLimit(nil, 60000), true), false},
		queryTestContext{driver.WithUseVST(driver.WithQueryTTL(nil, time.Minute), true), false},
		queryTestContext{driver.WithUseVST(driver.WithQueryBatchSize(driver.WithQueryCount(nil), 1), true), true},
		queryTestContext{driver.WithUseVST(driver.WithQueryCache(driver.WithQueryCount(driver.WithQueryBatchSize(nil, 2))), true), true},
	}

	// Run tests for every context alternative
	for _, qctx := range contexts {
		ctx := qctx.Context
		for i, test := range tests {
			cursor, err := db.Query(ctx, test.Query, test.BindVars)
			if err == nil {
				// Close upon exit of the function
				defer cursor.Close()
			}
			if test.ExpectSuccess {
				if err != nil {
					t.Errorf("Expected success in query %d (%s), got '%s'", i, test.Query, describe(err))
					continue
				}
				count := cursor.Count()
				if qctx.ExpectCount {
					if count != int64(len(test.ExpectedDocuments)) {
						t.Errorf("Expected count of %d, got %d in query %d (%s)", len(test.ExpectedDocuments), count, i, test.Query)
					}
				} else {
					if count != 0 {
						t.Errorf("Expected count of 0, got %d in query %d (%s)", count, i, test.Query)
					}
				}
				var result []interface{}
				for {
					hasMore := cursor.HasMore()
					doc := reflect.New(test.DocumentType)
					if _, err := cursor.ReadDocument(ctx, doc.Interface()); driver.IsNoMoreDocuments(err) {
						if hasMore {
							t.Error("HasMore returned true, but ReadDocument returns a IsNoMoreDocuments error")
						}
						break
					} else if err != nil {
						t.Errorf("Failed to result document %d: %s", len(result), describe(err))
					}
					if !hasMore {
						t.Error("HasMore returned false, but ReadDocument returns a document")
					}
					result = append(result, doc.Elem().Interface())
				}
				if len(result) != len(test.ExpectedDocuments) {
					t.Errorf("Expected %d documents, got %d in query %d (%s)", len(test.ExpectedDocuments), len(result), i, test.Query)
				} else {
					for resultIdx, resultDoc := range result {
						if !reflect.DeepEqual(resultDoc, test.ExpectedDocuments[resultIdx]) {
							t.Errorf("Unexpected document in query %d (%s) at index %d: got %+v, expected %+v", i, test.Query, resultIdx, resultDoc, test.ExpectedDocuments[resultIdx])
						}
					}
				}
				// Close anyway (this tests calling Close more than once)
				if err := cursor.Close(); err != nil {
					t.Errorf("Expected success in Close of cursor from query %d (%s), got '%s'", i, test.Query, describe(err))
				}
			} else {
				if err == nil {
					t.Errorf("Expected error in query %d (%s), got '%s'", i, test.Query, describe(err))
					continue
				}
			}
		}
	}
}
