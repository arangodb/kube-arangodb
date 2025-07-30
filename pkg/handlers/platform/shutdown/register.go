//
// DISCLAIMER
//
// Copyright 2024-2025 ArangoDB GmbH, Cologne, Germany
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

package shutdown

import (
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"

	operator "github.com/arangodb/kube-arangodb/pkg/operatorV2"
	"github.com/arangodb/kube-arangodb/pkg/operatorV2/event"
	utilConstants "github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
)

// RegisterInformer into operator
func RegisterInformer(operator operator.Operator, recorder event.Recorder, client kclient.Client, informer informers.SharedInformerFactory) error {
	if err := operator.RegisterInformer(informer.Core().V1().Pods().Informer(),
		Group(),
		Version(),
		Kind(), func(obj meta.Object) bool {
			if anns := obj.GetAnnotations(); len(anns) != 0 {
				if _, ok := anns[utilConstants.AnnotationShutdownManagedContainer]; ok {
					return true
				}
			}
			return false
		}); err != nil {
		return err
	}

	h := &handler{
		kubeClient: client.Kubernetes(),

		eventRecorder: recorder.NewInstance(Group(), Version(), Kind()),

		operator: operator,
	}

	if err := operator.RegisterHandler(h); err != nil {
		return err
	}

	return nil
}
