//
// DISCLAIMER
//
// Copyright 2016-2022 ArangoDB GmbH, Cologne, Germany
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

package clustersync

import (
	"context"

	"github.com/arangodb/kube-arangodb/pkg/apis/deployment"
	arangoClientSet "github.com/arangodb/kube-arangodb/pkg/generated/clientset/versioned"
	operator "github.com/arangodb/kube-arangodb/pkg/operatorV2"
	"github.com/arangodb/kube-arangodb/pkg/operatorV2/event"
	"github.com/arangodb/kube-arangodb/pkg/operatorV2/operation"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"

	deploymentApi "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type handler struct {
	client        arangoClientSet.Interface
	kubeClient    kubernetes.Interface
	eventRecorder event.RecorderInstance

	operator operator.Operator
}

func (*handler) Name() string {
	return deployment.ArangoClusterSynchronizationResourceKind
}

func (h *handler) Handle(item operation.Item) error {
	// Do not act on delete event
	if item.Operation == operation.Delete {
		return nil
	}

	// Get ClusterSynchronizations object. It also covers NotFound case
	clusterSync, err := h.client.DatabaseV1().ArangoClusterSynchronizations(item.Namespace).Get(context.Background(), item.Name, meta.GetOptions{})
	if err != nil {
		if k8sutil.IsNotFound(err) {
			return nil
		}
		h.operator.GetLogger().Error().Msgf("ListSimple fetch error %v", err)
		return err
	}

	// Update status on object
	if _, err = h.client.DatabaseV1().ArangoClusterSynchronizations(item.Namespace).UpdateStatus(context.Background(), clusterSync, meta.UpdateOptions{}); err != nil {
		h.operator.GetLogger().Error().Msgf("ListSimple status update error %v", err)
		return err
	}

	return nil
}

func (*handler) CanBeHandled(item operation.Item) bool {
	return item.Group == deploymentApi.SchemeGroupVersion.Group &&
		item.Version == deploymentApi.SchemeGroupVersion.Version &&
		item.Kind == deployment.ArangoClusterSynchronizationResourceKind
}
