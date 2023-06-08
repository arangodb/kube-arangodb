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

package state

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"
)

func caseJobPerformance(t *testing.T, jobs int) {
	j := NewJobsGenerator()

	currentJobs := jobs

	for _, p := range []JobPhase{
		JobPhaseToDo,
		JobPhasePending,
		JobPhaseFinished,
	} {
		z := rand.Intn(currentJobs + 1)

		j = j.Jobs(p, z)
		currentJobs -= z
	}
	j = j.Jobs(JobPhaseFailed, currentJobs)

	gen := j.Add()

	t.Run(fmt.Sprintf("Jobs %d", jobs), func(t *testing.T) {
		var s State
		var jids []JobID

		runWithMeasure(t, "Generate", func(t *testing.T) {
			s = GenerateState(t, gen)
		})

		runWithMeasure(t, "Count", func(t *testing.T) {
			jids = s.Target.GetJobIDs()
			i := len(jids)
			t.Logf("Count %d", i)
			require.Equal(t, jobs, i)
		})

		runCountWithMeasure(t, 16, "Lookup", func(t *testing.T) {
			id := jids[rand.Intn(len(jids))]

			_, z := s.Target.GetJob(id)
			require.NotEqual(t, JobPhaseUnknown, z)
		})
	})
}

func TestJobPerformance(t *testing.T) {
	caseJobPerformance(t, 16)
	caseJobPerformance(t, 256)
	caseJobPerformance(t, 1024)
	caseJobPerformance(t, 2048)
	caseJobPerformance(t, 2048*16)
}
