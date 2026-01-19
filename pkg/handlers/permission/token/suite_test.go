//
// DISCLAIMER
//
// Copyright 2026 ArangoDB GmbH, Cologne, Germany
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

package token

import (
	"context"
	_ "embed"
	"testing"

	kubernetes "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/arangodb/go-driver/v2/arangodb"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	fakeClientSet "github.com/arangodb/kube-arangodb/pkg/generated/clientset/versioned/fake"
	operator "github.com/arangodb/kube-arangodb/pkg/operatorV2"
	"github.com/arangodb/kube-arangodb/pkg/operatorV2/event"
	"github.com/arangodb/kube-arangodb/pkg/operatorV2/operation"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/tests"
)

type arangodbFakeClient struct {
	client arangodb.Client
}

func (a arangodbFakeClient) ArangoClient(ctx context.Context, client kubernetes.Interface, depl *api.ArangoDeployment) (arangodb.Client, error) {
	return a.client, nil
}

func newFakeHandler(t *testing.T) *handler {
	c := arangodbFakeClient{client: tests.TestArangoDBConfig(t).Client(t)}

	f := fakeClientSet.NewSimpleClientset()
	k := fake.NewSimpleClientset()

	h := &handler{
		client:        f,
		kubeClient:    k,
		eventRecorder: event.NewEventRecorder("mock", k).NewInstance(Group(), Version(), Kind()),
		operator:      operator.NewOperator("mock", "mock", util.Image{Image: "mock"}),
		provider:      c,
	}

	return h
}

func newItem(o operation.Operation, namespace, name string) operation.Item {
	return operation.Item{
		Group:   Group(),
		Version: Version(),
		Kind:    Kind(),

		Operation: o,

		Namespace: namespace,
		Name:      name,
	}
}
