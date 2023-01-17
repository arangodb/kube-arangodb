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

const (
	// ConditionTypeSecretsChanged indicates that the value of one of more secrets used by
	// the deployment have changed. Once that is the case, the operator will no longer
	// touch the deployment, until the original secrets have been restored.
	ConditionTypeSecretsChanged ConditionType = "SecretsChanged"

	// ConditionTypeBootstrapCompleted indicates that the initial cluster bootstrap has been completed.
	ConditionTypeBootstrapCompleted ConditionType = "BootstrapCompleted"
	// ConditionTypeBootstrapSucceded indicates that the initial cluster bootstrap completed successfully.
	ConditionTypeBootstrapSucceded ConditionType = "BootstrapSucceded"
)
