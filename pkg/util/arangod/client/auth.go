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
	"time"

	"k8s.io/client-go/kubernetes"

	"github.com/arangodb/go-driver/v2/connection"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/pod"
	"github.com/arangodb/kube-arangodb/pkg/util/cache"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	utilToken "github.com/arangodb/kube-arangodb/pkg/util/token"
	utilTokenLoader "github.com/arangodb/kube-arangodb/pkg/util/token/loader"
)

type Authentication interface {
	Authentication(ctx context.Context) (connection.Authentication, bool, error)
}

type AuthenticationFunc func(ctx context.Context) (connection.Authentication, bool, error)

func (f AuthenticationFunc) Authentication(ctx context.Context) (connection.Authentication, bool, error) {
	return f(ctx)
}

func DisabledAuth() Authentication {
	return AuthenticationFunc(func(ctx context.Context) (connection.Authentication, bool, error) {
		return nil, false, nil
	})
}

func FolderArangoDBAuthentication(path string) Authentication {
	if path == "" {
		return DisabledAuth()
	}

	folder := cache.NewObject[utilToken.Secret](utilTokenLoader.SecretCacheDirectory(path, 15*time.Second))

	return AuthenticationFunc(func(ctx context.Context) (connection.Authentication, bool, error) {
		secret, err := folder.Get(ctx)
		if err != nil {
			return nil, false, err
		}

		jwt, err := utilToken.NewClaims().With(utilToken.WithDefaultClaims(), utilToken.WithServerID("kube-arangodb"), utilToken.WithRelativeDuration(time.Minute)).Sign(secret)
		if err != nil {
			return nil, false, errors.WithStack(err)
		}
		return connection.NewHeaderAuth("Authorization", "bearer %s", jwt), true, nil
	})
}

func DirectArangoDBAuthentication(client kubernetes.Interface, depl *api.ArangoDeployment) Authentication {
	if !depl.GetAcceptedSpec().Authentication.IsAuthenticated() {
		return DisabledAuth()
	}

	secret := cache.NewObject[utilToken.Secret](utilTokenLoader.SecretCacheSecretAPI(client.CoreV1().Secrets(depl.GetNamespace()), pod.JWTSecretFolder(depl.GetName()), 15*time.Second))

	return AuthenticationFunc(func(ctx context.Context) (connection.Authentication, bool, error) {
		secret, err := secret.Get(ctx)
		if err != nil {
			return nil, false, err
		}

		jwt, err := utilToken.NewClaims().With(utilToken.WithDefaultClaims(), utilToken.WithServerID("kube-arangodb")).Sign(secret)
		if err != nil {
			return nil, false, errors.WithStack(err)
		}
		return connection.NewHeaderAuth("Authorization", "bearer %s", jwt), true, nil
	})
}
