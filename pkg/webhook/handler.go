//
// DISCLAIMER
//
// Copyright 2020 ArangoDB GmbH, Cologne, Germany
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
// Author Adam Janikowski
//

package webhook

import (
	"sync"

	"github.com/arangodb/kube-arangodb/pkg/util/log"
	"github.com/pkg/errors"
	admission "k8s.io/api/admission/v1beta1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Handlers []Handler

func (h Handlers) Filter(gvk meta.GroupVersionKind) Handlers {
	l := make(Handlers, 0, len(h))

	for _, handler := range h {
		if !handler.CanBeHandled(gvk) {
			continue
		}

		l = append(l, handler)
	}

	return l
}

func (h Handlers) AsCreateHandler() []ValidationCreateHandler {
	l := make([]ValidationCreateHandler, 0, len(h))

	for _, handler := range h {
		v, ok := handler.(ValidationCreateHandler)
		if !ok {
			continue
		}

		l = append(l, v)
	}

	return l
}

func (h Handlers) AsUpdateHandler() []ValidationUpdateHandler {
	l := make([]ValidationUpdateHandler, 0, len(h))

	for _, handler := range h {
		v, ok := handler.(ValidationUpdateHandler)
		if !ok {
			continue
		}

		l = append(l, v)
	}

	return l
}

func (h Handlers) AsDeleteHandler() []ValidationDeleteHandler {
	l := make([]ValidationDeleteHandler, 0, len(h))

	for _, handler := range h {
		v, ok := handler.(ValidationDeleteHandler)
		if !ok {
			continue
		}

		l = append(l, v)
	}

	return l
}

var (
	handlers     Handlers
	handlersLock sync.Mutex
)

func RegisterHandler(h Handler) error {
	handlersLock.Lock()
	defer handlersLock.Unlock()

	if h == nil {
		return errors.Errorf("Handler cannot be nil")
	}

	for _, existingHandler := range handlers {
		if existingHandler.Name() == h.Name() {
			return errors.Errorf("Handler with name %s already registered", h.Name())
		}
	}

	handlers = append(handlers, h)

	return nil
}

type Handler interface {
	CanBeHandled(gvk meta.GroupVersionKind) bool

	Name() string
}

type ValidationCreateHandler interface {
	Handler

	ValidateCreate(log log.Factory, request *admission.AdmissionRequest) (bool, string)
}

type ValidationUpdateHandler interface {
	Handler

	ValidateUpdate(log log.Factory, request *admission.AdmissionRequest) (bool, string)
}

type ValidationDeleteHandler interface {
	Handler

	ValidateDelete(log log.Factory, request *admission.AdmissionRequest) (bool, string)
}
