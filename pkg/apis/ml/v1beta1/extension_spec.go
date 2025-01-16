//
// DISCLAIMER
//
// Copyright 2024-2025 ArangoDB GmbH, Cologne, Germany
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

package v1beta1

import (
	schedulerIntegrationApi "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1beta1/integration"
	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"
	sharedApi "github.com/arangodb/kube-arangodb/pkg/apis/shared/v1"
)

type ArangoMLExtensionSpec struct {
	// MetadataService keeps the MetadataService configuration
	// +doc/immutable: This setting cannot be changed after the MetadataService has been created.
	MetadataService *ArangoMLExtensionSpecMetadataService `json:"metadataService,omitempty"`

	// Storage specifies the ArangoMLStorage used within Extension, if extension storage is used
	Storage *sharedApi.Object `json:"storage,omitempty"`

	// ArangoMLExtensionTemplate define Init job specification
	Init *ArangoMLExtensionTemplate `json:"init,omitempty"`

	// Deployment specifies how the ML extension will be deployed into cluster
	Deployment *ArangoMLExtensionSpecDeployment `json:"deployment,omitempty"`

	// JobsTemplates defines templates for jobs
	JobsTemplates *ArangoMLJobsTemplates `json:"jobsTemplates,omitempty"`

	// IntegrationSidecar define the integration sidecar spec
	IntegrationSidecar *schedulerIntegrationApi.Sidecar `json:"integrationSidecar,omitempty"`

	// ArangoMLExtensionSpecStorageType defines storage used for extension
	// +doc/enum: extension|Uses AragoMLStorage
	// +doc/enum: platform|Uses AragoPlatformStorage
	StorageType *ArangoMLExtensionSpecStorageType `json:"storageType,omitempty"`
}

func (a *ArangoMLExtensionSpec) GetMetadataService() *ArangoMLExtensionSpecMetadataService {
	if a == nil || a.MetadataService == nil {
		return nil
	}

	return a.MetadataService
}

func (a *ArangoMLExtensionSpec) GetInit() *ArangoMLExtensionTemplate {
	if a == nil || a.Init == nil {
		return nil
	}

	return a.Init
}

func (a *ArangoMLExtensionSpec) GetStorage2() *sharedApi.Object {
	if a == nil || a.Storage == nil {
		return nil
	}

	return a.Storage
}

func (a *ArangoMLExtensionSpec) GetStorageType() ArangoMLExtensionSpecStorageType {
	if a == nil {
		return ArangoMLExtensionSpecStorageTypeDefault
	}

	return a.StorageType.Get()
}

func (a *ArangoMLExtensionSpec) GetDeployment() *ArangoMLExtensionSpecDeployment {
	if a == nil || a.Deployment == nil {
		return nil
	}
	return a.Deployment
}

func (a *ArangoMLExtensionSpec) GetJobsTemplates() *ArangoMLJobsTemplates {
	if a == nil || a.JobsTemplates == nil {
		return nil
	}
	return a.JobsTemplates
}

func (a *ArangoMLExtensionSpec) GetIntegrationSidecar() *schedulerIntegrationApi.Sidecar {
	if a == nil || a.IntegrationSidecar == nil {
		return nil
	}
	return a.IntegrationSidecar
}

func (a *ArangoMLExtensionSpec) Validate() error {
	if a == nil {
		a = &ArangoMLExtensionSpec{}
	}

	return shared.WithErrors(shared.PrefixResourceErrors("spec",
		shared.PrefixResourceErrors("metadataService", a.GetMetadataService().Validate()),
		shared.PrefixResourceErrors("storage", shared.ValidateOptionalInterface(a.GetStorage2())),
		shared.PrefixResourceErrors("init", a.GetInit().Validate()),
		shared.PrefixResourceErrors("deployment", a.GetDeployment().Validate()),
		shared.PrefixResourceErrors("jobsTemplates", a.GetJobsTemplates().Validate()),
		shared.PrefixResourceErrors("integrationSidecar", a.GetIntegrationSidecar().Validate()),
	))
}
