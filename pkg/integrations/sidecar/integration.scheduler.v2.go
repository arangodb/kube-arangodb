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

package sidecar

import (
	core "k8s.io/api/core/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
)

type IntegrationSchedulerV2 struct {
	Core *Core

	DeploymentName string
	Spec           api.DeploymentSpec
}

func (i IntegrationSchedulerV2) Name() []string {
	return []string{"SCHEDULER", "V2"}
}

func (i IntegrationSchedulerV2) Validate() error {
	return nil
}

func (i IntegrationSchedulerV2) Envs() ([]core.EnvVar, error) {
	var envs = []core.EnvVar{
		{
			Name:  "INTEGRATION_SCHEDULER_V2",
			Value: "true",
		},
		{
			Name: "INTEGRATION_SCHEDULER_V2_NAMESPACE",
			ValueFrom: &core.EnvVarSource{
				FieldRef: &core.ObjectFieldSelector{
					FieldPath: "metadata.namespace",
				},
			},
		},
		{
			Name:  "INTEGRATION_SCHEDULER_V2_DEPLOYMENT",
			Value: i.DeploymentName,
		},
	}

	return i.Core.Envs(i, envs...), nil
}

func (i IntegrationSchedulerV2) GlobalEnvs() ([]core.EnvVar, error) {
	return nil, nil
}

func (i IntegrationSchedulerV2) Volumes() ([]core.Volume, []core.VolumeMount, error) {
	return nil, nil, nil
}
