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

package resources

import (
	core "k8s.io/api/core/v1"

	"github.com/arangodb/kube-arangodb/pkg/util"
)

func MergeProbes(a, b *core.Probe) *core.Probe {
	if a == nil && b == nil {
		return nil
	}
	if a == nil {
		return b.DeepCopy()
	}
	if b == nil {
		return a.DeepCopy()
	}

	return &core.Probe{
		ProbeHandler:                  util.FirstNotDefault(b.ProbeHandler, a.ProbeHandler),
		InitialDelaySeconds:           util.FirstNotDefault(b.InitialDelaySeconds, a.InitialDelaySeconds),
		TimeoutSeconds:                util.FirstNotDefault(b.TimeoutSeconds, a.TimeoutSeconds),
		PeriodSeconds:                 util.FirstNotDefault(b.PeriodSeconds, a.PeriodSeconds),
		SuccessThreshold:              util.FirstNotDefault(b.SuccessThreshold, a.SuccessThreshold),
		FailureThreshold:              util.FirstNotDefault(b.FailureThreshold, a.FailureThreshold),
		TerminationGracePeriodSeconds: util.FirstNotDefault(b.TerminationGracePeriodSeconds, a.TerminationGracePeriodSeconds),
	}
}
