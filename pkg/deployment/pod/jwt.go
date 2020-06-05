//
// DISCLAIMER
//
// Copyright 2020 ArangoDB GmbH, Cologne, Germany
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
// Author Adam Janikowski
//

package pod

import (
	"github.com/arangodb/kube-arangodb/pkg/deployment/resources/inspector"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
)

func JWT() Builder {
	return jwt{}
}

type jwt struct{}

func (e jwt) Args(i Input) k8sutil.OptionPairs {
	return nil
}

func (e jwt) Volumes(i Input) ([]core.Volume, []core.VolumeMount) {
	return nil, nil
}

func (e jwt) Verify(i Input, cachedStatus inspector.Inspector) error {
	if !i.Deployment.IsAuthenticated() {
		return nil
	}

	secret, exists := cachedStatus.Secret(i.Deployment.Authentication.GetJWTSecretName())
	if !exists {
		return errors.Errorf("Secret for JWT token is missing %s", i.Deployment.Authentication.GetJWTSecretName())
	}

	if err := k8sutil.ValidateTokenFromSecret(secret); err != nil {
		return errors.Wrapf(err, "Cluster JWT secret validation failed")
	}

	return nil
}
