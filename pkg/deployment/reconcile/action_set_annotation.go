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

package reconcile

import (
	"context"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/patch"
)

const (
	SetAnnotationActionKey         string = "key"
	SetAnnotationActionValue       string = "value"
	SetAnnotationActionValueRemove string = "-"
)

func newSetAnnotationAction(action api.Action, actionCtx ActionContext) Action {
	a := &actionSetAnnotation{}

	a.actionImpl = newActionImplDefRef(action, actionCtx)

	return a
}

type actionSetAnnotation struct {
	// actionImpl implement timeout and member id functions
	actionImpl

	actionEmptyCheckProgress
}

// Start starts the action for changing conditions on the provided member.
func (a actionSetAnnotation) Start(ctx context.Context) (bool, error) {
	key, ok := a.action.Params[SetAnnotationActionKey]
	if !ok {
		a.log.Info("key %s is missing in action definition", SetAnnotationActionKey)
		return true, nil
	}

	value, ok := a.action.Params[SetAnnotationActionValue]
	if !ok {
		a.log.Info("key %s is missing in action definition", SetAnnotationActionValue)
		return true, nil
	}

	if value == SetAnnotationActionValueRemove {
		if _, ok := a.actionCtx.GetAPIObject().GetAnnotations()[key]; ok {
			if err := a.actionCtx.ApplyPatch(ctx, patch.ItemRemove(patch.NewPath("metadata", "annotations", key))); err != nil {
				a.log.Str("key", key).Err(err).Warn("Unable to remove annotation")
				return true, nil
			}

			a.log.Str("key", key).Info("Removed annotation")
			return true, nil
		}
		a.log.Str("key", key).Info("Annotation already gone")
		return true, nil
	} else {
		if z, ok := a.actionCtx.GetAPIObject().GetAnnotations()[key]; ok {
			if value != z {
				if err := a.actionCtx.ApplyPatch(ctx, patch.ItemReplace(patch.NewPath("metadata", "annotations", key), value)); err != nil {
					a.log.Str("key", key).Str("value", value).Err(err).Warn("Unable to update annotation")
					return true, nil
				}

				a.log.Str("key", key).Str("value", value).Info("Updated annotation")
				return true, nil
			}
			a.log.Str("key", key).Str("value", value).Info("Annotation update not required")
			return true, nil
		}
		if err := a.actionCtx.ApplyPatch(ctx, patch.ItemAdd(patch.NewPath("metadata", "annotations", key), value)); err != nil {
			a.log.Str("key", key).Str("value", value).Err(err).Warn("Unable to add annotation")
			return true, nil
		}
		a.log.Str("key", key).Str("value", value).Info("Added annotation")
		return true, nil
	}
}
