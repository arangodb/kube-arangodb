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

package helpers

import (
	"context"
	"fmt"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	sharedApi "github.com/arangodb/kube-arangodb/pkg/apis/shared/v1"
	"github.com/arangodb/kube-arangodb/pkg/logging"
	operator "github.com/arangodb/kube-arangodb/pkg/operatorV2"
	"github.com/arangodb/kube-arangodb/pkg/operatorV2/event"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/kerrors"
)

type Action int

func (a Action) Or(b Action) Action {
	if b > a {
		return b
	}

	return a
}

const (
	ActionOK Action = iota
	ActionReplace
	ActionUpdate
)

type Object interface {
	comparable
	meta.Object
}

type ClientFactory[T Object] func(namespace string) Client[T]

type Client[T Object] interface {
	Get(ctx context.Context, name string, options meta.GetOptions) (T, error)
	Update(ctx context.Context, object T, options meta.UpdateOptions) (T, error)
	Create(ctx context.Context, object T, options meta.CreateOptions) (T, error)
	Delete(ctx context.Context, name string, options meta.DeleteOptions) error
}

type Generate[T Object] func(ctx context.Context, ref *sharedApi.Object) (T, bool, string, error)

type Config[T Object] struct {
	Events  event.RecorderInstance
	Logger  logging.Logger
	Factory ClientFactory[T]
	Kind    string
}

func NewUpdator[T Object](config Config[T]) Updator[T] {
	return updator[T]{
		config: config,
	}
}

type Updator[T Object] interface {
	OperatorUpdate(ctx context.Context, namespace string, parent meta.Object, ref **sharedApi.Object, generator Generate[T], decisions ...Decision[T]) (T, bool, error)
	Update(ctx context.Context, namespace string, parent meta.Object, ref **sharedApi.Object, generator Generate[T], decisions ...Decision[T]) (T, bool, error)
}

type updator[T Object] struct {
	config Config[T]
}

func (u updator[T]) OperatorUpdate(ctx context.Context, namespace string, parent meta.Object, ref **sharedApi.Object, generator Generate[T], decisions ...Decision[T]) (T, bool, error) {
	obj, changed, err := u.Update(ctx, namespace, parent, ref, generator, decisions...)
	if err != nil {
		return util.Default[T](), false, err
	}

	if changed {
		return obj, true, operator.Reconcile("Change in resources")
	}

	return obj, false, nil
}

func (u updator[T]) Update(ctx context.Context, namespace string, parent meta.Object, ref **sharedApi.Object, generator Generate[T], decisions ...Decision[T]) (T, bool, error) {
	decision := Decision[T](EmptyDecision[T]).With(decisions...)

	client := u.config.Factory(namespace)

	if ref == nil {
		return util.Default[T](), false, errors.Errorf("Reference is nil")
	}

	currentRef := *ref

	var discoveredObject T
	var discoveredObjectExists bool

	if currentRef != nil {
		object, err := util.WithKubernetesContextTimeoutP2A2(ctx, client.Get, currentRef.GetName(), meta.GetOptions{})
		if err != nil {
			if !kerrors.Is(err, kerrors.NotFound) {
				return util.Default[T](), false, err
			}

			u.config.Logger.
				Str("name", currentRef.GetName()).
				Str("checksum", currentRef.GetChecksum()).
				Str("uid", string(currentRef.GetUID())).
				Debug("Object has been removed")

			if events := u.config.Events; events != nil {
				events.Normal(parent, fmt.Sprintf("%sDeleted", u.config.Kind), "Deleted kubernetes %s %s", u.config.Kind, currentRef.GetName())
			}

			*ref = nil

			return util.Default[T](), true, nil
		}

		if object.GetDeletionTimestamp() != nil {
			// Object is currently deleting
			u.config.Logger.
				Str("name", currentRef.GetName()).
				Str("checksum", currentRef.GetChecksum()).
				Str("uid", string(currentRef.GetUID())).
				Debug("Object is currently deleting")
			return object, true, nil
		}

		if object.GetUID() != currentRef.GetUID() {
			u.config.Logger.
				Str("name", currentRef.GetName()).
				Str("checksum", currentRef.GetChecksum()).
				Str("uid", string(currentRef.GetUID())).
				Warn("Recreation Required as UID changed")

			if events := u.config.Events; events != nil {
				events.Warning(parent, fmt.Sprintf("%sForceDelete", u.config.Kind), "Deletion of kubernetes %s %s requested as UID changed", u.config.Kind, currentRef.GetName())
			}

			if err := util.WithKubernetesContextTimeoutP1A2(ctx, client.Delete, currentRef.GetName(), meta.DeleteOptions{}); err != nil {
				if !kerrors.Is(err, kerrors.NotFound) {
					return util.Default[T](), false, err
				}
			}

			return util.Default[T](), true, nil
		}

		discoveredObject = object
		discoveredObjectExists = true
	}

	object, skip, checksum, err := generator(ctx, currentRef.DeepCopy())
	if err != nil {
		return util.Default[T](), false, err
	}

	if skip {
		// Skip update as it is not required
		return util.Default[T](), false, nil
	}

	if object == util.Default[T]() {
		// Object is supposed to be removed
		if currentRef == nil {
			// Nothing to do
			return util.Default[T](), false, nil
		}

		// Remove object
		if err := util.WithKubernetesContextTimeoutP1A2(ctx, client.Delete, currentRef.GetName(), meta.DeleteOptions{}); err != nil {
			if !kerrors.Is(err, kerrors.NotFound) {
				return util.Default[T](), false, err
			}
		}

		u.config.Logger.
			Str("name", currentRef.GetName()).
			Str("checksum", currentRef.GetChecksum()).
			Str("uid", string(currentRef.GetUID())).
			Info("Object deletion has been requested")

		return util.Default[T](), true, nil
	}

	if !discoveredObjectExists {
		// Let's create Object
		newObject, err := util.WithKubernetesContextTimeoutP2A2(ctx, client.Create, object, meta.CreateOptions{})
		if err != nil {
			return util.Default[T](), false, err
		}

		currentRef = util.NewType(sharedApi.NewObjectWithChecksum(newObject, checksum))
		*ref = currentRef

		u.config.Logger.
			Str("name", currentRef.GetName()).
			Str("checksum", currentRef.GetChecksum()).
			Str("uid", string(currentRef.GetUID())).
			Info("Object has been created")

		if events := u.config.Events; events != nil {
			events.Normal(parent, fmt.Sprintf("%sCreated", u.config.Kind), "Created kubernetes %s %s", u.config.Kind, currentRef.GetName())
		}

		return newObject, true, nil
	}

	// Object exists, lets check if update is required
	action, err := decision(ctx, DecisionObject[T]{
		Checksum: currentRef.GetChecksum(),
		Object:   discoveredObject,
	}, DecisionObject[T]{
		Checksum: checksum,
		Object:   object,
	})
	if err != nil {
		return util.Default[T](), false, err
	}

	switch action {
	case ActionOK:
		// Nothing to do
		return discoveredObject, false, nil
	case ActionReplace:
		// Object needs to be removed
		u.config.Logger.
			Str("name", currentRef.GetName()).
			Str("checksum", currentRef.GetChecksum()).
			Str("uid", string(currentRef.GetUID())).
			Info("Object needs to be replaced")

		if err := util.WithKubernetesContextTimeoutP1A2(ctx, client.Delete, currentRef.GetName(), meta.DeleteOptions{}); err != nil {
			if !kerrors.Is(err, kerrors.NotFound) {
				return util.Default[T](), false, err
			}
		}

		if events := u.config.Events; events != nil {
			events.Normal(parent, fmt.Sprintf("%sReplaced", u.config.Kind), "Replaced kubernetes %s %s", u.config.Kind, currentRef.GetName())
		}

		return util.Default[T](), true, nil
	case ActionUpdate:
		u.config.Logger.
			Str("name", currentRef.GetName()).
			Str("checksum", currentRef.GetChecksum()).
			Str("uid", string(currentRef.GetUID())).
			Info("Object needs to be updated in-place")

		newObject, err := util.WithKubernetesContextTimeoutP2A2(ctx, client.Update, object, meta.UpdateOptions{})
		if err != nil {
			if !kerrors.Is(err, kerrors.NotFound) {
				return util.Default[T](), false, err
			}

			// Reconcile if object was not found
			return util.Default[T](), true, nil
		}

		if events := u.config.Events; events != nil {
			events.Normal(parent, fmt.Sprintf("%sUpdated", u.config.Kind), "Updated kubernetes %s %s", u.config.Kind, currentRef.GetName())
		}

		*ref = util.NewType(sharedApi.NewObjectWithChecksum(newObject, checksum))

		return newObject, true, nil

	default:
		return util.Default[T](), false, errors.Errorf("Unknown action returned")
	}
}
