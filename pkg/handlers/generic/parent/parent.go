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

package parent

import (
	"context"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/arangodb/kube-arangodb/pkg/logging"
	operator "github.com/arangodb/kube-arangodb/pkg/operatorV2"
	"github.com/arangodb/kube-arangodb/pkg/operatorV2/operation"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/generic"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/kerrors"
)

var logger = logging.Global().RegisterAndGetLogger("generic-parent-operator", logging.Info)

type NotifyHandlerClientFactory[T meta.Object, C NotifyHandlerClient[T]] func(namespace string) C

type NotifyHandlerClient[T meta.Object] interface {
	generic.GetInterface[T]
}

func NewNotifyHandler[T meta.Object, C NotifyHandlerClient[T]](name string, operator operator.Operator, client NotifyHandlerClientFactory[T, C], gvk schema.GroupVersionKind, notifiable ...schema.GroupVersionKind) operator.Handler {
	return notifyHandler[T, C]{
		name:       name,
		client:     client,
		gvk:        gvk,
		operator:   operator,
		notifiable: notifiable,
	}
}

type notifyHandler[T meta.Object, C NotifyHandlerClient[T]] struct {
	operator   operator.Operator
	name       string
	client     NotifyHandlerClientFactory[T, C]
	gvk        schema.GroupVersionKind
	notifiable []schema.GroupVersionKind
}

func (p notifyHandler[T, C]) Name() string {
	return p.name
}

func (p notifyHandler[T, C]) Handle(ctx context.Context, item operation.Item) error {
	logger := logger.WrapObj(item)
	if item.Operation == operation.Update {
		obj, err := p.client(item.Namespace).Get(ctx, item.Name, meta.GetOptions{})
		if err != nil {
			if kerrors.IsNotFound(err) {
				logger.Debug("Not Found")
				return nil
			}
			logger.Err(err).Warn("Unexpected Error")
			return err
		}

		for _, owner := range obj.GetOwnerReferences() {
			if i, err := operation.NewItemFromGVKObject(item.Operation, schema.FromAPIVersionAndKind(owner.APIVersion, owner.Kind), obj); err == nil {
				if p.isNotifiable(i) {
					logger.Debug("Parent notified")
					p.operator.EnqueueItem(i)
				} else {
					logger.Debug("Parent notify skipped")
				}
			}
		}
	}

	return nil
}

func (p notifyHandler[T, C]) isNotifiable(i operation.Item) bool {
	for _, g := range p.notifiable {
		if version := g.Version; version != "" && version != i.Version {
			continue
		}
		if kind := g.Kind; kind != "" && kind != i.Kind {
			continue
		}
		if group := g.Group; group != "" && group != i.Group {
			continue
		}

		return true
	}

	return false
}

func (p notifyHandler[T, C]) CanBeHandled(item operation.Item) bool {
	return item.GVK(p.gvk)
}
