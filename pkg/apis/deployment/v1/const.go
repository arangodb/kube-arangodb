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

type LabelsMode string

const (
	// LabelsDisabledMode disable annotations/labels override. Default if there is no annotations/labels set in ArangoDeployment
	LabelsDisabledMode LabelsMode = "disabled"
	// LabelsAppendMode add new annotations/labels without affecting old ones
	LabelsAppendMode LabelsMode = "append"
	// LabelsReplaceMode replace existing annotations/labels
	LabelsReplaceMode LabelsMode = "replace"
)

func (a LabelsMode) New() *LabelsMode {
	return &a
}

func (a *LabelsMode) Get(def LabelsMode) LabelsMode {
	if a == nil {
		return def
	}

	return *a
}
