//
// DISCLAIMER
//
// Copyright 2025 ArangoDB GmbH, Cologne, Germany
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

package service

import (
	_ "embed"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/arangodb/kube-arangodb/pkg/apis/apps"
	appsApi "github.com/arangodb/kube-arangodb/pkg/apis/apps/v1"
	"github.com/arangodb/kube-arangodb/pkg/handlers/platform/chart"
	operator "github.com/arangodb/kube-arangodb/pkg/operatorV2"
	"github.com/arangodb/kube-arangodb/pkg/operatorV2/event"
	"github.com/arangodb/kube-arangodb/pkg/operatorV2/operation"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/helm"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient/external"
)

func newFakeHandler(t *testing.T) (*handler, string, operator.Handler) {
	client, ns := external.ExternalClient(t)

	op := operator.NewOperator("mock", ns, "mock")
	recorder := event.NewEventRecorder("mock", client.Kubernetes())
	h, err := helm.NewClient(helm.Configuration{
		Namespace: ns,
		Config:    client.Config(),
	})
	require.NoError(t, err)

	return &handler{
		client:        client.Arango(),
		kubeClient:    client.Kubernetes(),
		eventRecorder: recorder.NewInstance(Group(), Version(), Kind()),
		operator:      op,
		helm:          h,
	}, ns, chart.Handler(op, recorder, client)
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
