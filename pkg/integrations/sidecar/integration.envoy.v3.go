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
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

type IntegrationEnvoyV3 struct {
	Core       *Core
	Deployment *api.ArangoDeployment
}

func (i IntegrationEnvoyV3) Name() (string, string) {
	return "ENVOY", "V3"
}

func (i IntegrationEnvoyV3) Validate() error {
	if i.Deployment == nil {
		return errors.Errorf("Deployment is nil")
	}

	return nil
}

func (i IntegrationEnvoyV3) Args() (k8sutil.OptionPairs, error) {
	options := k8sutil.CreateOptionPairs()

	options.Add("--integration.authentication.v1", true)
	options.Add("--integration.authentication.v1.enabled", i.Deployment.GetAcceptedSpec().IsAuthenticated())
	options.Add("--integration.authentication.v1.path", shared.ClusterJWTSecretVolumeMountDir)

	options.Merge(i.Core.Args(i))

	return options, nil
}
