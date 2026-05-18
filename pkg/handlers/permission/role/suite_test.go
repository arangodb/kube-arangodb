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

package role

import (
	_ "embed"
	"testing"

	"k8s.io/client-go/kubernetes/fake"

	fakeClientSet "github.com/arangodb/kube-arangodb/pkg/generated/clientset/versioned/fake"
	operator "github.com/arangodb/kube-arangodb/pkg/operatorV2"
	"github.com/arangodb/kube-arangodb/pkg/operatorV2/event"
	"github.com/arangodb/kube-arangodb/pkg/operatorV2/operation"
	"github.com/arangodb/kube-arangodb/pkg/util"
)

func newFakeHandler(_ *testing.T) *handler {
	f := fakeClientSet.NewSimpleClientset()
	k := fake.NewSimpleClientset()

	h := &handler{
		client:        f,
		kubeClient:    k,
		eventRecorder: event.NewEventRecorder("mock", k).NewInstance(Group(), Version(), Kind()),
		operator:      operator.NewOperator("mock", "mock", util.Image{Image: "mock"}),
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
