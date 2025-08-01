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
	"k8s.io/client-go/informers"

	arangoInformer "github.com/arangodb/kube-arangodb/pkg/generated/informers/externalversions"
	operator "github.com/arangodb/kube-arangodb/pkg/operatorV2"
	"github.com/arangodb/kube-arangodb/pkg/operatorV2/event"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/helm"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
)

// RegisterInformer into operator
func RegisterInformer(operator operator.Operator, recorder event.Recorder, client kclient.Client, informer arangoInformer.SharedInformerFactory, _ informers.SharedInformerFactory) error {
	if err := operator.RegisterInformer(informer.Platform().V1beta1().ArangoPlatformServices().Informer(),
		Group(),
		Version(),
		Kind()); err != nil {
		return err
	}

	hm, err := helm.NewClient(helm.Configuration{
		Namespace: operator.Namespace(),
		Config:    client.Config(),
		Driver:    util.NewType(helm.ConfigurationDriverSecret),
	})
	if err != nil {
		return err
	}

	h := &handler{
		client:     client.Arango(),
		kubeClient: client.Kubernetes(),

		helm: hm,

		eventRecorder: recorder.NewInstance(Group(), Version(), Kind()),

		operator: operator,
	}

	if err := operator.RegisterHandler(h); err != nil {
		return err
	}

	return nil
}
