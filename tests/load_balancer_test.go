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

	driver "github.com/arangodb/go-driver"
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/client"
)

func TestLoadBalancingCursorVST(t *testing.T) {
	longOrSkip(t)
	// run with VST
	loadBalancingCursorSubtest(t, true)
}

func TestLoadBalancingCursorHTTP(t *testing.T) {
	longOrSkip(t)
	// run with HTTP
	loadBalancingCursorSubtest(t, false)
}

func wasForwarded(r driver.Response) bool {
	h := r.Header("x-arango-request-forwarded-to")
	return h != ""
}

// tests cursor forwarding with load-balanced conn.
func loadBalancingCursorSubtest(t *testing.T, useVst bool) {
	c := client.MustNewInCluster()
	kubecli := mustNewKubeClient(t)
	ns := getNamespace(t)

	// Prepare deployment config
	namePrefix := "test-lb-"
	if useVst {
		namePrefix += "vst-"
	} else {
		namePrefix += "http-"
	}
	depl := newDeployment(namePrefix + uniuri.NewLen(4))
	depl.Spec.Mode = api.NewMode(api.DeploymentModeCluster)

	// Create deployment
	_, err := c.DatabaseV1().ArangoDeployments(ns).Create(depl)
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
	clOpts := &DatabaseClientOptions{
		UseVST:       useVst,
		ShortTimeout: true,
	}
	client := mustNewArangodDatabaseClient(ctx, kubecli, apiObject, t, clOpts)

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

	db := ensureDatabase(ctx, client, "lb_cursor_test", nil, t)
	for colName, colDocs := range collectionData {
		col := ensureCollection(ctx, db, colName, nil, t)
		if _, _, err := col.CreateDocuments(ctx, colDocs); err != nil {
			t.Fatalf("Expected success, got %s", err)
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
	}

	var r driver.Response
	// Setup context
	ctx = driver.WithResponse(driver.WithQueryBatchSize(nil, 1), &r)

	// keep track of whether at least one request was forwarded internally to the
	// correct coordinator behind the load balancer
	someRequestsForwarded := false
	someRequestsNotForwarded := false

	// Run tests for every context alternative
	for i, test := range tests {
		cursor, err := db.Query(ctx, test.Query, test.BindVars)
		if err == nil {
			// Close upon exit of the function
			defer cursor.Close()
		}
		if test.ExpectSuccess {
			if err != nil {
				t.Errorf("Expected success in query %d (%s), got '%s'", i, test.Query, err)
				continue
			}
			if count := cursor.Count(); count != 0 {
				t.Errorf("Expected count of 0, got %d in query %d (%s)", count, i, test.Query)
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
					t.Errorf("Failed to result document %d: %s", len(result), err)
				}
				if !hasMore {
					t.Error("HasMore returned false, but ReadDocument returns a document")
				}
				result = append(result, doc.Elem().Interface())
				if wasForwarded(r) {
					someRequestsForwarded = true
				} else {
					someRequestsNotForwarded = true
				}
				time.Sleep(200 * time.Millisecond)
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
				t.Errorf("Expected success in Close of cursor from query %d (%s), got '%s'", i, test.Query, err)
			}
		} else {
			if err == nil {
				t.Errorf("Expected error in query %d (%s), got '%s'", i, test.Query, err)
				continue
			}
		}
	}

	if !someRequestsForwarded {
		t.Error("Did not detect any request being forwarded behind load balancer!")
	}
	if !someRequestsNotForwarded {
		t.Error("Did not detect any request NOT being forwarded behind load balancer!")
	}
}
