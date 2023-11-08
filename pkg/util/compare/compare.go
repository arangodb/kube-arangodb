//
// DISCLAIMER
//
// Copyright 2023 ArangoDB GmbH, Cologne, Germany
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
	"github.com/arangodb/kube-arangodb/pkg/logging"
)

func P2[T interface{}, P1, P2 interface{}](
	log logging.Logger,
	p1 P1, p2 P2,
	actionBuilder api.ActionBuilder,
	checksum Checksum[T],
	spec, status Template[T],
	evaluators ...GenP2[T, P1, P2]) (mode Mode, plan api.Plan, err error) {
	if spec.GetChecksum() == status.GetChecksum() {
		return SkippedRotation, nil, nil
	}

	mode = SilentRotation

	// Try to fill fields
	newStatus := status.GetTemplate()
	currentSpec := spec.GetTemplate()

	g := NewFuncGenP2[T](p1, p2, currentSpec, newStatus)

	evaluatorsFunc := make([]Func, len(evaluators))

	for id := range evaluators {
		evaluatorsFunc[id] = g(evaluators[id])
	}

	if m, p, err := Evaluate(actionBuilder, evaluatorsFunc...); err != nil {
		log.Err(err).Error("Error while getting diff")
		return SkippedRotation, nil, err
	} else {
		mode = mode.And(m)
		plan = append(plan, p...)
	}

	// Diff has been generated! Proceed with calculations

	checksumString, err := checksum(newStatus)
	if err != nil {
		log.Err(err).Error("Error while getting checksum")
		return SkippedRotation, nil, err
	}

	if checksumString != spec.GetChecksum() {
		line := log

		// Rotate anyway!
		specData, statusData, diff, err := Diff(currentSpec, newStatus)
		if err == nil {

			if diff != "" {
				line = line.Str("diff", diff)
			}

			if specData != "" {
				line = line.Str("spec", specData)
			}

			if statusData != "" {
				line = line.Str("status", statusData)
			}
		}

		line.Info("Pod needs rotation - templates does not match")

		return GracefulRotation, nil, nil
	}

	return
}
