//
// DISCLAIMER
//
// Copyright 2016-2023 ArangoDB GmbH, Cologne, Germany
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

import (
	time "time"

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

// ChaosSpec holds configuration for the deployment chaos monkey.
type ChaosSpec struct {
	// Enabled switches the chaos monkey for a deployment on or off.
	Enabled *bool `json:"enabled,omitempty"`
	// Interval is the time between events
	Interval *time.Duration `json:"interval,omitempty"`
	// KillPodProbability is the chance of a pod being killed during an event
	KillPodProbability *Percent `json:"kill-pod-probability,omitempty"`
}

// IsEnabled returns the value of enabled.
func (s ChaosSpec) IsEnabled() bool {
	return util.TypeOrDefault[bool](s.Enabled)
}

// GetInterval returns the value of interval.
func (s ChaosSpec) GetInterval() time.Duration {
	return util.TypeOrDefault[time.Duration](s.Interval)
}

// GetKillPodProbability returns the value of kill-pod-probability.
func (s ChaosSpec) GetKillPodProbability() Percent {
	return PercentOrDefault(s.KillPodProbability)
}

// Validate the given spec
func (s ChaosSpec) Validate() error {
	if s.IsEnabled() {
		if s.GetInterval() <= 0 {
			return errors.WithStack(errors.Wrapf(ValidationError, "Interval must be > 0"))
		}
		if err := s.GetKillPodProbability().Validate(); err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

// SetDefaults fills in missing defaults
func (s *ChaosSpec) SetDefaults() {
	if s.GetInterval() == 0 {
		s.Interval = util.NewType[time.Duration](time.Minute)
	}
	if s.GetKillPodProbability() == 0 {
		s.KillPodProbability = NewPercent(50)
	}
}

// SetDefaultsFrom fills unspecified fields with a value from given source spec.
func (s *ChaosSpec) SetDefaultsFrom(source ChaosSpec) {
	if s.Enabled == nil {
		s.Enabled = util.NewTypeOrNil[bool](source.Enabled)
	}
	if s.Interval == nil {
		s.Interval = util.NewTypeOrNil[time.Duration](source.Interval)
	}
	if s.KillPodProbability == nil {
		s.KillPodProbability = NewPercentOrNil(source.KillPodProbability)
	}
}
