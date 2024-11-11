//
// DISCLAIMER
//
// Copyright 2024 ArangoDB GmbH, Cologne, Germany
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

package chart

import (
	_ "embed"

	"k8s.io/client-go/kubernetes/fake"

	"github.com/arangodb/kube-arangodb/pkg/apis/apps"
	appsApi "github.com/arangodb/kube-arangodb/pkg/apis/apps/v1"
	fakeClientSet "github.com/arangodb/kube-arangodb/pkg/generated/clientset/versioned/fake"
	operator "github.com/arangodb/kube-arangodb/pkg/operatorV2"
	"github.com/arangodb/kube-arangodb/pkg/operatorV2/event"
	"github.com/arangodb/kube-arangodb/pkg/operatorV2/operation"
)

//go:embed suite/example-1.0.0.tgz
var chart_1_0 []byte

//go:embed suite/example-1.1.0.tgz
var chart_1_1 []byte

func newFakeHandler() *handler {
	f := fakeClientSet.NewSimpleClientset()
	k := fake.NewSimpleClientset()

	h := &handler{
		client:        f,
		kubeClient:    k,
		eventRecorder: event.NewEventRecorder("mock", k).NewInstance(Group(), Version(), Kind()),
		operator:      operator.NewOperator("mock", "mock", "mock"),
	}

	return h
}

func newItem(o operation.Operation, namespace, name string) operation.Item {
	return operation.Item{
		Group:   appsApi.SchemeGroupVersion.Group,
		Version: appsApi.SchemeGroupVersion.Version,
		Kind:    apps.ArangoJobResourceKind,

		Operation: o,

		Namespace: namespace,
		Name:      name,
	}
}
