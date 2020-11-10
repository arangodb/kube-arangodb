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

package reconcile

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/arangodb/go-driver"
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/rs/zerolog"
)

func init() {
	registerAction(api.ActionTypeBootstrapSetPassword, newBootstrapSetPasswordAction)
}

func newBootstrapSetPasswordAction(log zerolog.Logger, action api.Action, actionCtx ActionContext) Action {
	a := &actionBootstrapSetPassword{}

	a.actionImpl = newActionImplDefRef(log, action, actionCtx, defaultTimeout)

	return a
}

type actionBootstrapSetPassword struct {
	// actionImpl implement timeout and member id functions
	actionImpl

	actionEmptyCheckProgress
}

func (a actionBootstrapSetPassword) Start(ctx context.Context) (bool, error) {
	spec := a.actionCtx.GetSpec()

	if user, ok := a.action.GetParam("user"); !ok {
		a.log.Warn().Msgf("User param is not set in action")
		return true, nil
	} else {
		if secret, ok := spec.Bootstrap.PasswordSecretNames[user]; !ok {
			a.log.Warn().Msgf("User does not exist in password hashes")
			return true, nil
		} else {
			ctx, c := context.WithTimeout(context.Background(), a.Timeout(spec))
			defer c()
			if password, err := a.setUserPassword(ctx, user, secret.Get()); err != nil {
				return false, err
			} else {
				passwordSha := util.SHA256FromString(password)

				if err := a.actionCtx.WithStatusUpdate(func(s *api.DeploymentStatus) bool {
					if s.SecretHashes == nil {
						s.SecretHashes = &api.SecretHashes{}
					}

					if s.SecretHashes.Users == nil {
						s.SecretHashes.Users = map[string]string{}
					}

					if u, ok := s.SecretHashes.Users[user]; !ok || u != passwordSha {
						s.SecretHashes.Users[user] = passwordSha
						return true
					}
					return false
				}); err != nil {
					return false, err
				}
			}
		}
	}
	return true, nil
}

func (a actionBootstrapSetPassword) setUserPassword(ctx context.Context, user, secret string) (string, error) {
	a.log.Debug().Msgf("Bootstrapping user %s, secret %s", user, secret)

	client, err := a.actionCtx.GetDatabaseClient(ctx)
	if err != nil {
		return "", maskAny(err)
	}

	password, err := a.ensureUserPasswordSecret(user, secret)
	if err != nil {
		return "", maskAny(err)
	}

	// Obtain the user
	if u, err := client.User(context.Background(), user); driver.IsNotFound(err) {
		_, err := client.CreateUser(context.Background(), user, &driver.UserOptions{Password: password})
		return password, maskAny(err)
	} else if err == nil {
		return password, maskAny(u.Update(context.Background(), driver.UserOptions{
			Password: password,
		}))
	} else {
		return "", err
	}
}

func (a actionBootstrapSetPassword) ensureUserPasswordSecret(user, secret string) (string, error) {
	cache := a.actionCtx.GetCachedStatus()

	if auth, ok := cache.Secret(secret); !ok {
		// Create new one
		tokenData := make([]byte, 32)
		if _, err := rand.Read(tokenData); err != nil {
			return "", err
		}
		token := hex.EncodeToString(tokenData)
		owner := a.actionCtx.GetAPIObject().AsOwner()

		if err := k8sutil.CreateBasicAuthSecret(a.actionCtx.SecretsInterface(), secret, user, token, &owner); err != nil {
			return "", err
		}

		return token, nil
	} else {
		user, pass, err := k8sutil.GetSecretAuthCredentials(auth)
		if err == nil && user == user {
			return pass, nil
		}
		return "", fmt.Errorf("invalid secret format in secret %s", secret)
	}
}
