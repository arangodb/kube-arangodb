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

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	sharedApi "github.com/arangodb/kube-arangodb/pkg/apis/shared/v1"
	"github.com/arangodb/kube-arangodb/pkg/logging"
	operator "github.com/arangodb/kube-arangodb/pkg/operatorV2"
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

type Client[T Object] interface {
	Get(ctx context.Context, name string, options meta.GetOptions) (T, error)
	Update(ctx context.Context, object T, options meta.UpdateOptions) (T, error)
	Create(ctx context.Context, object T, options meta.CreateOptions) (T, error)
	Delete(ctx context.Context, name string, options meta.DeleteOptions) error
}

type Generate[T Object] func(ctx context.Context, ref *sharedApi.Object) (T, bool, string, error)

func OperatorUpdate[T Object](ctx context.Context, logger logging.Logger, client Client[T], ref **sharedApi.Object, generator Generate[T], decisions ...Decision[T]) (bool, error) {
	changed, err := Update[T](ctx, logger, client, ref, generator, decisions...)
	if err != nil {
		return false, err
	}

	if changed {
		return true, operator.Reconcile("Change in resources")
	}

	return false, nil
}

func Update[T Object](ctx context.Context, logger logging.Logger, client Client[T], ref **sharedApi.Object, generator Generate[T], decisions ...Decision[T]) (bool, error) {
	decision := Decision[T](EmptyDecision[T]).With(decisions...)

	if ref == nil {
		return false, errors.Errorf("Reference is nil")
	}

	currentRef := *ref

	var discoveredObject T
	var discoveredObjectExists bool

	if currentRef != nil {
		object, err := util.WithKubernetesContextTimeoutP2A2(ctx, client.Get, currentRef.GetName(), meta.GetOptions{})
		if err != nil {
			if !kerrors.Is(err, kerrors.NotFound) {
				return false, err
			}

			*ref = nil
			logger.
				Str("name", currentRef.GetName()).
				Str("checksum", currentRef.GetChecksum()).
				Str("uid", string(currentRef.GetUID())).
				Debug("Object has been removed")

			return true, nil
		}

		if object.GetDeletionTimestamp() != nil {
			// Object is currently deleting
			logger.
				Str("name", currentRef.GetName()).
				Str("checksum", currentRef.GetChecksum()).
				Str("uid", string(currentRef.GetUID())).
				Debug("Object is currently deleting")
			return true, nil
		}

		if object.GetUID() != currentRef.GetUID() {
			logger.
				Str("name", currentRef.GetName()).
				Str("checksum", currentRef.GetChecksum()).
				Str("uid", string(currentRef.GetUID())).
				Warn("Recreation Required as UID changed")

			if err := util.WithKubernetesContextTimeoutP1A2(ctx, client.Delete, currentRef.GetName(), meta.DeleteOptions{}); err != nil {
				if !kerrors.Is(err, kerrors.NotFound) {
					return false, err
				}
			}

			return true, nil
		}

		discoveredObject = object
		discoveredObjectExists = true
	}

	object, skip, checksum, err := generator(ctx, currentRef.DeepCopy())
	if err != nil {
		return false, err
	}

	if skip {
		// Skip update as it is not required
		return false, nil
	}

	if object == util.Default[T]() {
		// Object is supposed to be removed
		if currentRef == nil {
			// Nothing to do
			return false, nil
		}

		// Remove object
		if err := util.WithKubernetesContextTimeoutP1A2(ctx, client.Delete, currentRef.GetName(), meta.DeleteOptions{}); err != nil {
			if !kerrors.Is(err, kerrors.NotFound) {
				return false, err
			}
		}

		logger.
			Str("name", currentRef.GetName()).
			Str("checksum", currentRef.GetChecksum()).
			Str("uid", string(currentRef.GetUID())).
			Info("Object deletion has been requested")

		return true, nil
	}

	if !discoveredObjectExists {
		// Let's create Object
		newObject, err := util.WithKubernetesContextTimeoutP2A2(ctx, client.Create, object, meta.CreateOptions{})
		if err != nil {
			return false, err
		}

		currentRef = util.NewType(sharedApi.NewObjectWithChecksum(newObject, checksum))
		*ref = currentRef
		logger.
			Str("name", currentRef.GetName()).
			Str("checksum", currentRef.GetChecksum()).
			Str("uid", string(currentRef.GetUID())).
			Info("Object has been created")

		return true, nil
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
		return false, err
	}

	switch action {
	case ActionOK:
		// Nothing to do
		return false, nil
	case ActionReplace:
		// Object needs to be removed
		logger.
			Str("name", currentRef.GetName()).
			Str("checksum", currentRef.GetChecksum()).
			Str("uid", string(currentRef.GetUID())).
			Info("Object needs to be replaced")

		if err := util.WithKubernetesContextTimeoutP1A2(ctx, client.Delete, currentRef.GetName(), meta.DeleteOptions{}); err != nil {
			if !kerrors.Is(err, kerrors.NotFound) {
				return false, err
			}
		}

		return true, nil
	case ActionUpdate:
		logger.
			Str("name", currentRef.GetName()).
			Str("checksum", currentRef.GetChecksum()).
			Str("uid", string(currentRef.GetUID())).
			Info("Object needs to be updated in-place")

		newObject, err := util.WithKubernetesContextTimeoutP2A2(ctx, client.Update, object, meta.UpdateOptions{})
		if err != nil {
			if !kerrors.Is(err, kerrors.NotFound) {
				return false, err
			}

			// Reconcile if object was not found
			return true, nil
		}

		*ref = util.NewType(sharedApi.NewObjectWithChecksum(newObject, checksum))

		return true, nil

	default:
		return false, errors.Errorf("Unknown action returned")
	}
}
