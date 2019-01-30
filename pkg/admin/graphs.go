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

package admin

import (
	"context"
	"fmt"

	driver "github.com/arangodb/go-driver"
	api "github.com/arangodb/kube-arangodb/pkg/apis/admin/v1alpha"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// Graph stores information about a arangodb Graph
type Graph struct {
	api.ArangoGraph
}

func (graph *Graph) GetAPIObject() ArangoResource {
	return graph
}

func (graph *Graph) AsRuntimeObject() runtime.Object {
	return &graph.ArangoGraph
}

func (graph *Graph) SetAPIObject(obj api.ArangoGraph) {
	graph.ArangoGraph = obj
}

func (graph *Graph) Load(kube KubeClient) (runtime.Object, error) {
	return kube.ArangoGraphs(graph.GetNamespace()).Get(graph.GetName(), metav1.GetOptions{})
}

func (graph *Graph) Update(kube KubeClient) error {
	new, err := kube.ArangoGraphs(graph.GetNamespace()).Update(&graph.ArangoGraph)
	if err != nil {
		return err
	}
	graph.SetAPIObject(*new)
	return nil
}

func (graph *Graph) UpdateStatus(kube KubeClient) error {
	_, err := kube.ArangoGraphs(graph.GetNamespace()).UpdateStatus(&graph.ArangoGraph)
	return err
}

func (graph *Graph) GetDeploymentName(resolv DeploymentNameResolver) (string, error) {
	return resolv.DeploymentByDatabase(graph.ArangoGraph.GetDatabaseResourceName())
}

func NewGraphFromObject(object runtime.Object) (*Graph, error) {
	if agraph, ok := object.(*api.ArangoGraph); ok {
		agraph.Spec.SetDefaults(agraph.GetName())
		if err := agraph.Spec.Validate(); err != nil {
			return nil, err
		}
		return &Graph{
			ArangoGraph: *agraph,
		}, nil
	}

	return nil, fmt.Errorf("Not a ArangoGraph")
}

// GetFinalizerName returns the name of the finalizer for this Graph
func (graph *Graph) GetFinalizerName() string {
	return "database-admin-graph-" + graph.Spec.GetName()
}

// Reconcile updates the Graph resource to the given spec
func (graph *Graph) Reconcile(ctx context.Context, admin ReconcileContext) {
	dbn := graph.GetDatabaseResourceName()
	gname := graph.Spec.GetName()
	finalizerName := graph.GetFinalizerName()

	if graph.GetDeletionTimestamp() != nil {
		removeFinalizer := false
		defer func() {
			if removeFinalizer {
				admin.RemoveFinalizer(graph)
				admin.RemoveResourceFinalizer(dbn, finalizerName)
			}
		}()

		client, err := admin.GetArangoDatabaseClient(ctx, dbn)
		if driver.IsNotFound(err) {
			removeFinalizer = true // Database gone
			return
		} else if err != nil {
			admin.ReportError(graph, "Connect to database", err)
			return
		}

		agraph, err := client.Graph(ctx, gname)
		if driver.IsNotFound(err) {
			removeFinalizer = true
			return
		} else if err == nil {
			if err := agraph.Remove(ctx); err != nil {
				admin.ReportError(graph, "Remove graph", err)
				return
			}
			removeFinalizer = true
		}
	} else {
		if !admin.HasFinalizer(graph) {
			admin.AddFinalizer(graph)
		}

		admin.AddResourceFinalizer(dbn, finalizerName)

		client, err := admin.GetArangoDatabaseClient(ctx, dbn)
		if err != nil {
			admin.ReportError(graph, "Connect to database", err)
			return
		}

		_, err = client.Graph(ctx, gname)
		if driver.IsNotFound(err) {
			if admin.GetCreatedAt(graph) != nil {
				admin.ReportWarning(graph, "Collection lost", "The collection was lost and will be recreated")
			}

			_, err := client.CreateGraph(ctx, gname, nil)
			if err != nil {
				admin.ReportError(graph, "Get Graph", err)
				return
			}

			admin.SetCreatedAtNow(graph)
		} else if err == nil {
			// Graph exists

		} else {
			admin.ReportError(graph, "Get Graph", err)
			return
		}

		admin.SetCondition(graph, api.ConditionTypeReady, v1.ConditionTrue, "Graph ready", "Graph is ready")
	}
}
