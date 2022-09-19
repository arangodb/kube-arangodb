//
// DISCLAIMER
//
// Copyright 2016-2022 ArangoDB GmbH, Cologne, Germany
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

package reconcile

import (
	"context"
	"crypto/rand"
	"encoding/hex"

	"github.com/arangodb/go-driver"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

func newBootstrapSetPasswordAction(action api.Action, actionCtx ActionContext) Action {
	a := &actionBootstrapSetPassword{}

	a.actionImpl = newActionImplDefRef(action, actionCtx)

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
		a.log.Warn("User param is not set in action")
		return true, nil
	} else {
		if secret, ok := spec.Bootstrap.PasswordSecretNames[user]; !ok {
			a.log.Warn("User does not exist in password hashes")
			return true, nil
		} else {
			ctxChild, cancel := globals.GetGlobals().Timeouts().ArangoD().WithTimeout(ctx)
			defer cancel()

			if password, err := a.setUserPassword(ctxChild, user, secret.Get()); err != nil {
				return false, err
			} else {
				passwordSha := util.SHA256FromString(password)

				if err := a.actionCtx.WithStatusUpdate(ctx, func(s *api.DeploymentStatus) bool {
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
	a.log.Debug("Bootstrapping user %s, secret %s", user, secret)

	client, err := a.actionCtx.GetMembersState().State().GetDatabaseClient()
	if err != nil {
		return "", errors.WithStack(err)
	}

	password, err := a.ensureUserPasswordSecret(ctx, user, secret)
	if err != nil {
		return "", errors.WithStack(err)
	}

	// Obtain the user
	ctxChild, cancel := globals.GetGlobalTimeouts().ArangoD().WithTimeout(ctx)
	defer cancel()
	if u, err := client.User(ctxChild, user); err != nil {
		if !driver.IsNotFound(err) {
			return "", err
		}

		err = globals.GetGlobalTimeouts().ArangoD().RunWithTimeout(ctx, func(ctxChild context.Context) error {
			_, err := client.CreateUser(ctxChild, user, &driver.UserOptions{Password: password})
			return err
		})

		return password, errors.WithStack(err)
	} else {
		err = globals.GetGlobalTimeouts().ArangoD().RunWithTimeout(ctx, func(ctxChild context.Context) error {
			return u.Update(ctxChild, driver.UserOptions{
				Password: password,
			})
		})

		return password, errors.WithStack(err)
	}
}

func (a actionBootstrapSetPassword) ensureUserPasswordSecret(ctx context.Context, user, secret string) (string, error) {
	if auth, ok := a.actionCtx.ACS().CurrentClusterCache().Secret().V1().GetSimple(secret); !ok {
		// Create new one
		tokenData := make([]byte, 32)
		if _, err := rand.Read(tokenData); err != nil {
			return "", err
		}
		token := hex.EncodeToString(tokenData)
		owner := a.actionCtx.GetAPIObject().AsOwner()

		err := k8sutil.CreateBasicAuthSecret(ctx, a.actionCtx.ACS().CurrentClusterCache().SecretsModInterface().V1(), secret, user, token, &owner)
		if err != nil {
			return "", err
		}

		return token, nil
	} else {
		_, pass, err := k8sutil.GetSecretAuthCredentials(auth)
		if err == nil {
			return pass, nil
		}
		return "", errors.Newf("invalid secret format in secret %s", secret)
	}
}
