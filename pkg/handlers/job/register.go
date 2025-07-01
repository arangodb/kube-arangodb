//
// DISCLAIMER
//
// Copyright 2016-2025 ArangoDB GmbH, Cologne, Germany
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

package job

import (
	"github.com/arangodb/kube-arangodb/pkg/apis/apps"
	appsApi "github.com/arangodb/kube-arangodb/pkg/apis/apps/v1"
	arangoInformer "github.com/arangodb/kube-arangodb/pkg/generated/informers/externalversions"
	operator "github.com/arangodb/kube-arangodb/pkg/operatorV2"
	"github.com/arangodb/kube-arangodb/pkg/operatorV2/event"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
)

func newEventInstance(eventRecorder event.Recorder) event.RecorderInstance {
	return eventRecorder.NewInstance(appsApi.SchemeGroupVersion.Group,
		appsApi.SchemeGroupVersion.Version,
		apps.ArangoJobResourceKind)
}

// RegisterInformer into operator
func RegisterInformer(operator operator.Operator, recorder event.Recorder, client kclient.Client, informer arangoInformer.SharedInformerFactory) error {
	if err := operator.RegisterInformer(informer.Apps().V1().ArangoJobs().Informer(),
		appsApi.SchemeGroupVersion.Group,
		appsApi.SchemeGroupVersion.Version,
		apps.ArangoJobResourceKind); err != nil {
		return err
	}

	h := &handler{
		client:        client.Arango(),
		kubeClient:    client.Kubernetes(),
		eventRecorder: newEventInstance(recorder),

		operator: operator,
	}

	if err := operator.RegisterHandler(h); err != nil {
		return err
	}

	return nil
}
