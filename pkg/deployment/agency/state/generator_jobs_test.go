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
)

func NewJobsGenerator() JobsGeneratorInterface {
	return &jobsGenerator{
		jobs: map[JobPhase]map[JobID]Job{},
	}
}

type jobsGenerator struct {
	id   int
	jobs map[JobPhase]map[JobID]Job
}

func (j *jobsGenerator) Jobs(phase JobPhase, jobs int, jobTypes ...string) JobsGeneratorInterface {
	if len(jobTypes) == 0 {
		jobTypes = []string{"moveShard"}
	}

	z := j.jobs[phase]
	if z == nil {
		z = map[JobID]Job{}
	}

	for i := 0; i < jobs; i++ {
		q := j.id
		j.id++
		id := fmt.Sprintf("s%07d", q)
		z[JobID(id)] = Job{
			Type: jobTypes[rand.Intn(len(jobTypes))],
		}
	}

	j.jobs[phase] = z

	return j
}

func (j *jobsGenerator) Add() Generator {
	return func(t *testing.T, s *State) {
		if m := j.jobs[JobPhaseToDo]; len(m) > 0 {
			if s.Target.JobToDo == nil {
				s.Target.JobToDo = map[JobID]Job{}
			}

			for k, v := range m {
				s.Target.JobToDo[k] = v
			}
		}
		if m := j.jobs[JobPhasePending]; len(m) > 0 {
			if s.Target.JobPending == nil {
				s.Target.JobPending = map[JobID]Job{}
			}

			for k, v := range m {
				s.Target.JobPending[k] = v
			}
		}
		if m := j.jobs[JobPhaseFailed]; len(m) > 0 {
			if s.Target.JobFailed == nil {
				s.Target.JobFailed = map[JobID]Job{}
			}

			for k, v := range m {
				s.Target.JobFailed[k] = v
			}
		}
		if m := j.jobs[JobPhaseFinished]; len(m) > 0 {
			if s.Target.JobFinished == nil {
				s.Target.JobFinished = map[JobID]Job{}
			}

			for k, v := range m {
				s.Target.JobFinished[k] = v
			}
		}
	}
}

type JobsGeneratorInterface interface {
	Jobs(phase JobPhase, jobs int, jobTypes ...string) JobsGeneratorInterface
	Add() Generator
}
