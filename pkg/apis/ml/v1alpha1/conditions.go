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

package v1alpha1

import api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"

const (
	ReadyCondition                         api.ConditionType = "Ready"
	SpecValidCondition                     api.ConditionType = "SpecValid"
	ExtensionFoundCondition                api.ConditionType = "ExtensionFound"
	ExtensionStorageFoundCondition         api.ConditionType = "StorageFound"
	ExtensionDeploymentFoundCondition      api.ConditionType = "DeploymentFound"
	ExtensionBootstrapCompletedCondition   api.ConditionType = "BootstrapCompleted"
	ExtensionMetadataServiceValidCondition api.ConditionType = "MetadataServiceValid"
	ExtensionServiceAccountReadyCondition  api.ConditionType = "ServiceAccountReady"
	ExtensionStatefulSetReadyCondition     api.ConditionType = "ExtensionDeploymentReady"
	ExtensionTLSEnabledCondition           api.ConditionType = "TLSEnabled"
	LicenseValidCondition                  api.ConditionType = "LicenseValid"
	CronJobSyncedCondition                 api.ConditionType = "CronJobSynced"
	BatchJobSyncedCondition                api.ConditionType = "BatchJobSynced"
)
