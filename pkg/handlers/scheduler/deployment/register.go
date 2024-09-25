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

package deployment

import (
	apps "k8s.io/api/apps/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"

	arangoClientSet "github.com/arangodb/kube-arangodb/pkg/generated/clientset/versioned"
	arangoInformer "github.com/arangodb/kube-arangodb/pkg/generated/informers/externalversions"
	"github.com/arangodb/kube-arangodb/pkg/handlers/generic/parent"
	operator "github.com/arangodb/kube-arangodb/pkg/operatorV2"
	"github.com/arangodb/kube-arangodb/pkg/operatorV2/event"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

// RegisterInformer into operator
func RegisterInformer(operator operator.Operator, recorder event.Recorder, client arangoClientSet.Interface,
	kubeClient kubernetes.Interface, informer arangoInformer.SharedInformerFactory, kubeInformer informers.SharedInformerFactory) error {

	if err := operator.RegisterInformer(informer.Scheduler().V1beta1().ArangoSchedulerDeployments().Informer(),
		Group(),
		Version(),
		Kind()); err != nil {
		return err
	}

	h := &handler{
		client:     client,
		kubeClient: kubeClient,

		eventRecorder: recorder.NewInstance(Group(), Version(), Kind()),

		operator: operator,
	}

	h.init()

	if err := operator.RegisterHandler(h); err != nil {
		return err
	}

	{
		deployment := k8sutil.AppsV1DeploymentGVK()

		if err := operator.RegisterInformer(kubeInformer.Apps().V1().Deployments().Informer(),
			deployment.Group,
			deployment.Version,
			deployment.Kind); err != nil {
			return err
		}

		deploymentHandler := parent.NewNotifyHandler[*apps.Deployment]("apps-deployment-v1-parent", operator, kubeClient.AppsV1().Deployments, deployment, GVK())

		if err := operator.RegisterHandler(deploymentHandler); err != nil {
			return err
		}
	}

	return nil
}
