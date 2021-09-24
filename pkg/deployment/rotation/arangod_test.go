//
// DISCLAIMER
//
// Copyright 2020-2021 ArangoDB GmbH, Cologne, Germany
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
// Author Adam Janikowski
//

package rotation

import (
	"testing"

	core "k8s.io/api/core/v1"
)

func Test_ArangoD_SchedulerName(t *testing.T) {
	testCases := []TestCase{
		{
			name: "Change SchedulerName from Empty",
			spec: buildPodSpec(func(pod *core.PodTemplateSpec) {
				pod.Spec.SchedulerName = ""
			}),
			status: buildPodSpec(func(pod *core.PodTemplateSpec) {
				pod.Spec.SchedulerName = "new"
			}),

			expectedMode: SilentRotation,
		},
		{
			name: "Change SchedulerName into Empty",
			spec: buildPodSpec(func(pod *core.PodTemplateSpec) {
				pod.Spec.SchedulerName = "new"
			}),
			status: buildPodSpec(func(pod *core.PodTemplateSpec) {
				pod.Spec.SchedulerName = ""
			}),

			expectedMode: SilentRotation,
		},
		{
			name: "SchedulerName equals",
			spec: buildPodSpec(func(pod *core.PodTemplateSpec) {
				pod.Spec.SchedulerName = ""
			}),
			status: buildPodSpec(func(pod *core.PodTemplateSpec) {
				pod.Spec.SchedulerName = ""
			}),

			expectedMode: SkippedRotation,
		},
	}

	runTestCases(t)(testCases...)
}
