//
// DISCLAIMER
//
// Copyright 2023-2024 ArangoDB GmbH, Cologne, Germany
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

package compare

import (
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
)

type Mode int

const (
	// SkippedRotation Skips the rotation. Returned plan is ignored
	SkippedRotation Mode = iota
	// SilentRotation Propagates changes without a restart. Returned plan is executed in High actions
	SilentRotation
	// InPlaceRotation Silently accept changes without a restart. Returned plan is executed in Normal actions
	InPlaceRotation
	// GracefulRotation Schedule pod restart. Returned plan is ignored
	GracefulRotation
	// EnforcedRotation Enforce pod restart. Returned plan is ignored
	EnforcedRotation
)

const (
	SkippedRotationString  = "Skipped"
	SilentRotationString   = "Silent"
	InPlaceRotationString  = "InPlace"
	GracefulRotationString = "Graceful"
	EnforcedRotationString = "EnforcedSkipped"
)

func (m Mode) String() string {
	switch m {
	case SkippedRotation:
		return SkippedRotationString
	case SilentRotation:
		return SilentRotationString
	case InPlaceRotation:
		return InPlaceRotationString
	case GracefulRotation:
		return GracefulRotationString
	case EnforcedRotation:
		return EnforcedRotationString
	}

	return ""
}

// And returns the higher value of the rotation mode.
func (m Mode) And(b Mode) Mode {
	if m > b {
		return m
	}

	return b
}

func (m Mode) Func() Func {
	return func(builder api.ActionBuilder) (mode Mode, plan api.Plan, err error) {
		return m, nil, nil
	}
}
