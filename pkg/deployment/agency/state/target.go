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

type Target struct {
	// Jobs Section

	JobToDo     Jobs `json:"ToDo,omitempty"`
	JobPending  Jobs `json:"Pending,omitempty"`
	JobFailed   Jobs `json:"Failed,omitempty"`
	JobFinished Jobs `json:"Finished,omitempty"`

	// Servers Section

	CleanedServers     Servers `json:"CleanedServers,omitempty"`
	ToBeCleanedServers Servers `json:"ToBeCleanedServers,omitempty"`

	// HotBackup section

	HotBackup TargetHotBackup `json:"HotBackup,omitempty"`
}

func (s Target) GetJob(id JobID) (Job, JobPhase) {
	if v, ok := s.JobToDo[id]; ok {
		return v, JobPhaseToDo
	}
	if v, ok := s.JobPending[id]; ok {
		return v, JobPhasePending
	}
	if v, ok := s.JobFailed[id]; ok {
		return v, JobPhaseFailed
	}
	if v, ok := s.JobFinished[id]; ok {
		return v, JobPhaseFinished
	}

	return Job{}, JobPhaseUnknown
}

func (s Target) GetJobIDs() []JobID {
	r := make([]JobID, 0, len(s.JobToDo)+len(s.JobPending)+len(s.JobFinished)+len(s.JobFailed))

	for k := range s.JobToDo {
		r = append(r, k)
	}

	for k := range s.JobPending {
		r = append(r, k)
	}

	for k := range s.JobFinished {
		r = append(r, k)
	}

	for k := range s.JobFailed {
		r = append(r, k)
	}

	return r
}

type TargetHotBackup struct {
	Create Timestamp `json:"Create,omitempty"`
}
