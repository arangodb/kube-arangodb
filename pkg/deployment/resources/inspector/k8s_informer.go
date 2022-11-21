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

package inspector

import (
	"k8s.io/client-go/tools/cache"

	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/definitions"
)

func (i *inspectorState) eventHandler(component definitions.Component) cache.ResourceEventHandler {
	return eventHandler{
		i:         i,
		component: component,
	}
}

type eventHandler struct {
	i         *inspectorState
	component definitions.Component
}

func (e eventHandler) OnAdd(obj interface{}) {
	e.i.throttles.Invalidate(e.component)
}

func (e eventHandler) OnUpdate(oldObj, newObj interface{}) {
	e.i.throttles.Invalidate(e.component)
}

func (e eventHandler) OnDelete(obj interface{}) {
	e.i.throttles.Invalidate(e.component)
}
