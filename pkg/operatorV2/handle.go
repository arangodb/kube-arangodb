//
// DISCLAIMER
//
// Copyright 2023-2026 ArangoDB GmbH, Cologne, Germany
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

package operator

import (
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	sharedApi "github.com/arangodb/kube-arangodb/pkg/apis/shared/v1"
)

type Condition struct {
	Status  bool
	Reason  string
	Message string
	Hash    string
}

func WithCondition(conditions *api.ConditionList, condition api.ConditionType, changed bool, err error) (bool, error) {
	var hash string

	if changed || err != nil {
		// Condition should be false
		if conditions.UpdateWithHash(condition, false, "Not ready", "Not ready", hash) {
			changed = true
		}
	} else {
		if conditions.UpdateWithHash(condition, true, "Ready", "Ready", hash) {
			changed = true
		}
	}

	if err == nil || IsStop(err) {
		if changed {
			err = Reconcile("Condition changed")
		}
	}

	return changed, err
}

func WithConditionChange(conditions *api.ConditionList, condition api.ConditionType, c *Condition, changed bool, err error) (bool, error) {
	if c == nil {
		if conditions.Remove(condition) {
			changed = true
		}
	} else {
		if conditions.UpdateWithHash(condition, c.Status, c.Reason, c.Message, c.Hash) {
			changed = true
		}
	}

	if err == nil || IsStop(err) {
		if changed {
			err = Reconcile("Condition changed")
		}
	}

	if err == nil && changed {
		err = Reconcile("Condition changed")
	}

	return changed, err
}

func WithSharedCondition(conditions *sharedApi.ConditionList, condition sharedApi.ConditionType, changed bool, err error) (bool, error) {
	var hash string

	if changed || err != nil {
		// Condition should be false
		if conditions.UpdateWithHash(condition, false, "Not ready", "Not ready", hash) {
			changed = true
		}
	} else {
		if conditions.UpdateWithHash(condition, true, "Ready", "Ready", hash) {
			changed = true
		}
	}

	if err == nil || IsStop(err) {
		if changed {
			err = Reconcile("Condition changed")
		}
	}

	return changed, err
}

func WithSharedConditionChange(conditions *sharedApi.ConditionList, condition sharedApi.ConditionType, c *Condition, changed bool, err error) (bool, error) {
	if c == nil {
		if conditions.Remove(condition) {
			changed = true
		}
	} else {
		if conditions.UpdateWithHash(condition, c.Status, c.Reason, c.Message, c.Hash) {
			changed = true
		}
	}

	if err == nil || IsStop(err) {
		if changed {
			err = Reconcile("Condition changed")
		}
	}

	if err == nil && changed {
		err = Reconcile("Condition changed")
	}

	return changed, err
}
