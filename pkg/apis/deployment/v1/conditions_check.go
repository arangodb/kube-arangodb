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

package v1

import "time"

type ConditionCheck interface {
	Evaluate() bool

	Exists() ConditionCheck
	IsTrue() ConditionCheck
	IsFalse() ConditionCheck

	LastTransition(d time.Duration) ConditionCheck
}

var _ ConditionCheck = conditionCheck{}

type conditionCheck struct {
	condition Condition
	exists    bool
}

func (c conditionCheck) LastTransition(d time.Duration) ConditionCheck {
	if c.exists && (!c.condition.LastTransitionTime.IsZero() && time.Since(c.condition.LastTransitionTime.Time) >= d) {
		return c
	}

	return newConditionCheckConst(false)
}

func (c conditionCheck) IsTrue() ConditionCheck {
	if c.condition.IsTrue() {
		return c
	}

	return newConditionCheckConst(false)
}

func (c conditionCheck) IsFalse() ConditionCheck {
	if !c.condition.IsTrue() {
		return c
	}

	return newConditionCheckConst(false)
}

func (c conditionCheck) Evaluate() bool {
	return true
}

func (c conditionCheck) Exists() ConditionCheck {
	if c.exists {
		return c
	}

	return newConditionCheckConst(false)
}

func newConditionCheckConst(c bool) ConditionCheck {
	return conditionCheckConst(c)
}

type conditionCheckConst bool

func (c conditionCheckConst) LastTransition(d time.Duration) ConditionCheck {
	return c
}

func (c conditionCheckConst) IsTrue() ConditionCheck {
	return c
}

func (c conditionCheckConst) IsFalse() ConditionCheck {
	return c
}

func (c conditionCheckConst) Evaluate() bool {
	return bool(c)
}

func (c conditionCheckConst) Exists() ConditionCheck {
	return c
}
