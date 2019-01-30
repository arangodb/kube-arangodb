//
// DISCLAIMER
//
// Copyright 2018 ArangoDB GmbH, Cologne, Germany
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

package admin

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	driver "github.com/arangodb/go-driver"
	api "github.com/arangodb/kube-arangodb/pkg/apis/admin/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// User stores information about a arangodb user
type User struct {
	api.ArangoUser
}

func (user *User) AsRuntimeObject() runtime.Object {
	return &user.ArangoUser
}

func (user *User) GetAPIObject() ArangoResource {
	return user
}

func (user *User) SetAPIObject(obj api.ArangoUser) {
	user.ArangoUser = obj
}

func (user *User) Load(kube KubeClient) (runtime.Object, error) {
	return kube.ArangoUsers(user.GetNamespace()).Get(user.GetName(), metav1.GetOptions{})
}

func (user *User) Update(kube KubeClient) error {
	new, err := kube.ArangoUsers(user.GetNamespace()).Update(&user.ArangoUser)
	if err != nil {
		return err
	}
	user.SetAPIObject(*new)
	return nil
}

func (user *User) UpdateStatus(kube KubeClient) error {
	_, err := kube.ArangoUsers(user.GetNamespace()).UpdateStatus(&user.ArangoUser)
	return err
}

func (user *User) GetDeploymentName(resolv DeploymentNameResolver) (string, error) {
	return user.ArangoUser.GetDeploymentName(), nil
}

// NewUserFromObject creates a new User from a runtime.Object, if possible
func NewUserFromObject(object runtime.Object) (*User, error) {
	if auser, ok := object.(*api.ArangoUser); ok {
		auser.Spec.SetDefaults(auser.GetName())
		if err := auser.Spec.Validate(); err != nil {
			return nil, err
		}
		return &User{
			ArangoUser: *auser,
		}, nil
	}

	return nil, fmt.Errorf("Not a ArangoUser")
}

// ensurePasswordSecret ensures that a secret containing a password for the user exists
// if not it is created with a random password. The password is returned as string
func (user *User) ensurePasswordSecret(ctx context.Context, admin ReconcileContext) (string, error) {
	secretName := user.Spec.GetPasswordSecretName()
	data, err := admin.GetKubeSecret(user, secretName)
	if k8sutil.IsNotFound(err) {
		tokenData := make([]byte, 32)
		rand.Read(tokenData)
		token := hex.EncodeToString(tokenData)

		if err := admin.CreateKubeSecret(user, secretName, map[string][]byte{
			constants.SecretUsername: []byte(user.Spec.GetName()),
			constants.SecretPassword: []byte(token),
		}); err != nil {
			return "", err
		}

		return token, nil

	} else if err == nil {
		username, ok := data[constants.SecretUsername]
		if ok && string(username) == user.Spec.GetName() {
			pass, ok := data[constants.SecretPassword]
			if ok {
				return string(pass), nil
			}
		}
		return "", fmt.Errorf("invalid secret format in secret %s", secretName)
	}

	return "", err
}

// Reconcile updates the database resource to the given spec
func (user *User) Reconcile(ctx context.Context, admin ReconcileContext) {
	name := user.Spec.GetName()

	if user.GetDeletionTimestamp() != nil {
		removeFinalizer := false
		defer func() {
			if removeFinalizer {
				admin.RemoveFinalizer(user)
				admin.RemoveDeploymentFinalizer(user)
			}
		}()

		client, err := admin.GetArangoClient(ctx, user)
		if err == nil {
			auser, err := client.User(ctx, name)
			if driver.IsNotFound(err) {
				// cool, user is gone
				removeFinalizer = true
			} else if err == nil {
				if err := auser.Remove(ctx); err != nil {
					admin.ReportError(user, "Remove user", err)
				} else {
					removeFinalizer = true
				}
			} else {
				admin.ReportError(user, "Get user", err)
			}

		} else {
			admin.ReportError(user, "Connect to deployment", err)
			return
		}
	} else {

		client, err := admin.GetArangoClient(ctx, user)
		if err == nil {

			if !admin.HasFinalizer(user) {
				admin.AddFinalizer(user)
			}

			auser, err := client.User(ctx, name)
			if driver.IsNotFound(err) {
				// Get user credentials
				passwd, err := user.ensurePasswordSecret(ctx, admin)
				if err == nil {
					if _, err := client.CreateUser(ctx, name, &driver.UserOptions{Password: passwd}); err != nil {
						admin.ReportError(user, "Create user", err)
					} else {
						admin.SetCreatedAtNow(user)
					}
				} else {
					admin.ReportError(user, "Get credentials", err)
				}

			} else if err == nil {
				// User was found
				passwd, err := user.ensurePasswordSecret(ctx, admin)
				if err == nil {
					if err := auser.Update(ctx, driver.UserOptions{Password: passwd}); err != nil {
						admin.ReportError(user, "Update user", err)
					}

					admin.SetCondition(user, api.ConditionTypeReady, v1.ConditionTrue, "User updated", "User is ready")
				}
			} else {
				admin.ReportError(user, "Get user", err)
			}
		} else {
			admin.ReportError(user, "Connect to deployment", err)
			return
		}

		admin.SetCondition(user, api.ConditionTypeReady, v1.ConditionTrue, "User ready", "User is ready")
	}
}
