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
	"fmt"

	core "k8s.io/api/core/v1"

	"github.com/arangodb/kube-arangodb/pkg/util/constants"
)

type IntegrationShutdownV1 struct {
	Core *Core
}

func (i IntegrationShutdownV1) Annotations() (map[string]string, error) {
	return map[string]string{
		fmt.Sprintf("%s/%s", constants.AnnotationShutdownContainer, ContainerName): ListenPortHealthName,
		constants.AnnotationShutdownManagedContainer:                               "true",
	}, nil
}

func (i IntegrationShutdownV1) Name() []string {
	return []string{"SHUTDOWN", "V1"}
}

func (i IntegrationShutdownV1) Validate() error {
	return nil
}

func (i IntegrationShutdownV1) Envs() ([]core.EnvVar, error) {
	var envs = []core.EnvVar{
		{
			Name:  "INTEGRATION_SHUTDOWN_V1",
			Value: "true",
		},
	}

	return i.Core.Envs(i, envs...), nil
}

func (i IntegrationShutdownV1) GlobalEnvs() ([]core.EnvVar, error) {
	return nil, nil
}

func (i IntegrationShutdownV1) Volumes() ([]core.Volume, []core.VolumeMount, error) {
	return nil, nil, nil
}
