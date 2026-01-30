//
// DISCLAIMER
//
// Copyright 2026 ArangoDB GmbH, Cologne, Germany
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

package client

import (
	"context"

	"k8s.io/client-go/kubernetes"

	"github.com/arangodb/go-driver/v2/connection"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/pod"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	utilToken "github.com/arangodb/kube-arangodb/pkg/util/token"
)

type Authentication interface {
	Authentication(ctx context.Context) (connection.Authentication, bool, error)
}

type AuthenticationFunc func(ctx context.Context) (connection.Authentication, bool, error)

func (f AuthenticationFunc) Authentication(ctx context.Context) (connection.Authentication, bool, error) {
	return f(ctx)
}

func DirectArangoDBAuthentication(client kubernetes.Interface, depl *api.ArangoDeployment) Authentication {
	return AuthenticationFunc(func(ctx context.Context) (connection.Authentication, bool, error) {
		if v := depl.GetAcceptedSpec().Authentication.GetJWTSecretName(); v != api.JWTSecretNameDisabled {
			secrets := client.CoreV1().Secrets(depl.GetNamespace())
			ctxChild, cancel := globals.GetGlobalTimeouts().Kubernetes().WithTimeout(ctx)
			defer cancel()
			s, err := k8sutil.GetTokenFolderSecret(ctxChild, secrets, pod.JWTSecretFolder(depl.GetName()))
			if err != nil {
				return nil, false, errors.WithStack(err)
			}
			jwt, err := utilToken.NewClaims().With(utilToken.WithDefaultClaims(), utilToken.WithServerID("kube-arangodb")).Sign(s)
			if err != nil {
				return nil, false, errors.WithStack(err)
			}
			return connection.NewHeaderAuth("Authorization", "bearer %s", jwt), true, nil
		}

		return nil, false, nil
	})
}
