//
// DISCLAIMER
//
// Copyright 2023-2025 ArangoDB GmbH, Cologne, Germany
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

//go:build !enterprise

package operator

import (
	"k8s.io/client-go/informers"

	arangoInformer "github.com/arangodb/kube-arangodb/pkg/generated/informers/externalversions"
	operatorV2 "github.com/arangodb/kube-arangodb/pkg/operatorV2"
	"github.com/arangodb/kube-arangodb/pkg/operatorV2/event"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
)

// onStartML starts the operator and run till given channel is closed.
func (o *Operator) onStartML(stop <-chan struct{}) {
	panic("Unable to start ML Operator in Community")
}

func (o *Operator) onStartOperatorV2ML(operator operatorV2.Operator, recorder event.Recorder, client kclient.Client, informer arangoInformer.SharedInformerFactory, kubeInformer informers.SharedInformerFactory) {
	panic("Unable to start ML Operator in Community")
}

// onStartAnalytics starts the operator and run till given channel is closed.
func (o *Operator) onStartAnalytics(stop <-chan struct{}) {
	panic("Unable to start Analytics Operator in Community")
}

func (o *Operator) onStartOperatorV2Analytics(operator operatorV2.Operator, recorder event.Recorder, client kclient.Client, informer arangoInformer.SharedInformerFactory, kubeInformer informers.SharedInformerFactory) {
	panic("Unable to start Analytics Operator in Community")
}
