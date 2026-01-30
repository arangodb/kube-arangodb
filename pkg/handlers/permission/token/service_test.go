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

package token

import (
	"encoding/hex"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	permissionApi "github.com/arangodb/kube-arangodb/pkg/apis/permission/v1alpha1"
	sharedApi "github.com/arangodb/kube-arangodb/pkg/apis/shared/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/pod"
	"github.com/arangodb/kube-arangodb/pkg/operatorV2/operation"
	"github.com/arangodb/kube-arangodb/pkg/util"
	utilConstants "github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/arangodb/kube-arangodb/pkg/util/tests"
	utilToken "github.com/arangodb/kube-arangodb/pkg/util/token"
)

func extractRoles(t *testing.T, token utilToken.Token) []string {
	absRoles, ok := token.Claims()[utilToken.ClaimRoles].([]any)
	require.True(t, ok)

	ret := make([]string, len(absRoles))

	for id := range absRoles {
		v, ok := absRoles[id].(string)
		require.True(t, ok)
		ret[id] = v
	}

	return ret
}

func Test_ServiceReconcile(t *testing.T) {
	handler := newFakeHandler(t)

	// Arrange
	extension := tests.NewMetaObject[*permissionApi.ArangoPermissionToken](t, tests.FakeNamespace, "example",
		func(t *testing.T, obj *permissionApi.ArangoPermissionToken) {})

	deployment := tests.NewMetaObject[*api.ArangoDeployment](t, tests.FakeNamespace, "example",
		func(t *testing.T, obj *api.ArangoDeployment) {})

	jwt := tests.NewMetaObject[*core.Secret](t, tests.FakeNamespace, pod.JWTSecretFolder("example"), func(t *testing.T, obj *core.Secret) {
		obj.Data = map[string][]byte{}
	})

	tokenData := make([]byte, 32)
	util.Rand().Read(tokenData)
	token := hex.EncodeToString(tokenData)
	util.Rand().Read(tokenData)
	tokenNew := hex.EncodeToString(tokenData)

	refresh := tests.CreateObjects(t, handler.kubeClient, handler.client, &extension, &deployment, &jwt)

	t.Run("Missing deployment section", func(t *testing.T) {
		// Test
		require.NoError(t, tests.Handle(handler, tests.NewItem(t, operation.Update, extension)))

		// Refresh
		refresh(t)

		// Validate
		require.False(t, extension.Status.Conditions.IsTrue(permissionApi.SpecValidCondition))
		require.False(t, extension.Status.Conditions.IsTrue(permissionApi.DeploymentFoundCondition))
		require.False(t, extension.Status.Conditions.IsTrue(permissionApi.ReadyCondition))
	})

	t.Run("Missing deployment", func(t *testing.T) {
		// Test
		tests.Update(t, handler.kubeClient, handler.client, &extension, func(t *testing.T, obj *permissionApi.ArangoPermissionToken) {
			obj.Spec.Deployment = &sharedApi.Object{
				Name: "unknown",
			}
		})

		require.NoError(t, tests.Handle(handler, tests.NewItem(t, operation.Update, extension)))

		// Refresh
		refresh(t)

		// Validate
		require.True(t, extension.Status.Conditions.IsTrue(permissionApi.SpecValidCondition))
		require.False(t, extension.Status.Conditions.IsTrue(permissionApi.DeploymentFoundCondition))
		require.False(t, extension.Status.Conditions.IsTrue(permissionApi.DeploymentReachableCondition))
		require.False(t, extension.Status.Conditions.IsTrue(permissionApi.ReadyCondition))
	})

	t.Run("Existing deployment, invalid TTL", func(t *testing.T) {
		// Test
		tests.Update(t, handler.kubeClient, handler.client, &extension, func(t *testing.T, obj *permissionApi.ArangoPermissionToken) {
			obj.Spec.Deployment = &sharedApi.Object{
				Name: "example",
			}

			obj.Spec.TTL = util.NewType(meta.Duration{Duration: time.Minute})
		})

		require.NoError(t, tests.Handle(handler, tests.NewItem(t, operation.Update, extension)))

		// Refresh
		refresh(t)

		// Validate
		require.False(t, extension.Status.Conditions.IsTrue(permissionApi.SpecValidCondition))
		require.False(t, extension.Status.Conditions.IsTrue(permissionApi.DeploymentFoundCondition))
		require.False(t, extension.Status.Conditions.IsTrue(permissionApi.DeploymentReachableCondition))
		require.False(t, extension.Status.Conditions.IsTrue(permissionApi.ReadyCondition))
	})

	t.Run("Existing deployment, unauthenticated", func(t *testing.T) {
		// Test
		tests.Update(t, handler.kubeClient, handler.client, &extension, func(t *testing.T, obj *permissionApi.ArangoPermissionToken) {
			obj.Spec.Deployment = &sharedApi.Object{
				Name: "example",
			}

			obj.Spec.TTL = nil
		})
		tests.Update(t, handler.kubeClient, handler.client, &deployment, func(t *testing.T, obj *api.ArangoDeployment) {
			obj.Spec.Authentication.JWTSecretName = util.NewType("None")
		})

		require.NoError(t, tests.Handle(handler, tests.NewItem(t, operation.Update, extension)))

		// Refresh
		refresh(t)

		// Validate
		require.True(t, extension.Status.Conditions.IsTrue(permissionApi.SpecValidCondition))
		require.True(t, extension.Status.Conditions.IsTrue(permissionApi.DeploymentFoundCondition))
		require.True(t, extension.Status.Conditions.IsTrue(permissionApi.DeploymentReachableCondition))
		require.True(t, extension.Status.Conditions.IsTrue(permissionApi.ReadyCondition))
		require.NotNil(t, extension.Status.Secret)

		secret := tests.GetObject[*core.Secret](t, handler.kubeClient, handler.client, tests.FakeNamespace, extension.Status.Secret.GetName())

		require.Contains(t, secret.Data, core.ServiceAccountTokenKey)
		require.Len(t, secret.Data[core.ServiceAccountTokenKey], 0)
	})

	t.Run("Existing deployment, authenticated, but missing", func(t *testing.T) {
		// Test
		tests.Update(t, handler.kubeClient, handler.client, &extension, func(t *testing.T, obj *permissionApi.ArangoPermissionToken) {
			obj.Spec.Deployment = &sharedApi.Object{
				Name: "example",
			}
		})
		tests.Update(t, handler.kubeClient, handler.client, &deployment, func(t *testing.T, obj *api.ArangoDeployment) {
			obj.Spec.Authentication.JWTSecretName = nil
		})

		require.EqualError(t, tests.Handle(handler, tests.NewItem(t, operation.Update, extension)), "No '-' data found in secret 'example-jwt-folder'")
	})

	t.Run("Existing deployment", func(t *testing.T) {
		// Test
		tests.Update(t, handler.kubeClient, handler.client, &extension, func(t *testing.T, obj *permissionApi.ArangoPermissionToken) {
			obj.Spec.Deployment = &sharedApi.Object{
				Name: "example",
			}
		})
		tests.Update(t, handler.kubeClient, handler.client, &jwt, func(t *testing.T, obj *core.Secret) {
			obj.Data = map[string][]byte{
				utilConstants.ActiveJWTKey: []byte(token),
			}
		})

		require.NoError(t, tests.Handle(handler, tests.NewItem(t, operation.Update, extension)))

		// Refresh
		refresh(t)

		// Validate
		require.True(t, extension.Status.Conditions.IsTrue(permissionApi.SpecValidCondition))
		require.True(t, extension.Status.Conditions.IsTrue(permissionApi.DeploymentFoundCondition))
		require.True(t, extension.Status.Conditions.IsTrue(permissionApi.DeploymentReachableCondition))
		require.True(t, extension.Status.Conditions.IsTrue(permissionApi.ReadyCondition))
		require.NotNil(t, extension.Status.Secret)

		secret := tests.GetObject[*core.Secret](t, handler.kubeClient, handler.client, tests.FakeNamespace, extension.Status.Secret.GetName())

		validator, err := k8sutil.GetTokenFolderSecret(t.Context(), handler.kubeClient.CoreV1().Secrets(deployment.GetNamespace()), pod.JWTSecretFolder(deployment.GetName()))
		require.NoError(t, err)

		require.Contains(t, secret.Data, core.ServiceAccountTokenKey)
		token, err := validator.Validate(string(secret.Data[core.ServiceAccountTokenKey]))
		require.NoError(t, err)
		require.EqualValues(t, token.Claims()[utilToken.ClaimPreferredUsername], extension.Status.User.GetName())

		roles := extractRoles(t, token)
		require.Len(t, roles, 0)
	})

	t.Run("New deployment - refresh", func(t *testing.T) {
		oldSecret := tests.GetObject[*core.Secret](t, handler.kubeClient, handler.client, tests.FakeNamespace, extension.Status.Secret.GetName()).Data[core.ServiceAccountTokenKey]

		// Test
		tests.Update(t, handler.kubeClient, handler.client, &extension, func(t *testing.T, obj *permissionApi.ArangoPermissionToken) {
			obj.Spec.Deployment = &sharedApi.Object{
				Name: "example",
			}
		})
		tests.Update(t, handler.kubeClient, handler.client, &jwt, func(t *testing.T, obj *core.Secret) {
			obj.Data = map[string][]byte{
				utilConstants.ActiveJWTKey: []byte(token),
			}
		})

		require.NoError(t, tests.Handle(handler, tests.NewItem(t, operation.Update, extension)))

		// Refresh
		refresh(t)

		// Validate
		require.True(t, extension.Status.Conditions.IsTrue(permissionApi.SpecValidCondition))
		require.True(t, extension.Status.Conditions.IsTrue(permissionApi.DeploymentFoundCondition))
		require.True(t, extension.Status.Conditions.IsTrue(permissionApi.DeploymentReachableCondition))
		require.True(t, extension.Status.Conditions.IsTrue(permissionApi.ReadyCondition))
		require.NotNil(t, extension.Status.Secret)

		secret := tests.GetObject[*core.Secret](t, handler.kubeClient, handler.client, tests.FakeNamespace, extension.Status.Secret.GetName())

		validator, err := k8sutil.GetTokenFolderSecret(t.Context(), handler.kubeClient.CoreV1().Secrets(deployment.GetNamespace()), pod.JWTSecretFolder(deployment.GetName()))
		require.NoError(t, err)

		require.Contains(t, secret.Data, core.ServiceAccountTokenKey)
		token, err := validator.Validate(string(secret.Data[core.ServiceAccountTokenKey]))
		require.NoError(t, err)
		require.EqualValues(t, token.Claims()[utilToken.ClaimPreferredUsername], extension.Status.User.GetName())
		require.Equal(t, oldSecret, secret.Data[core.ServiceAccountTokenKey])

		roles := extractRoles(t, token)
		require.Len(t, roles, 0)
	})

	t.Run("New deployment - refresh JWT", func(t *testing.T) {
		oldSecret := tests.GetObject[*core.Secret](t, handler.kubeClient, handler.client, tests.FakeNamespace, extension.Status.Secret.GetName()).Data[core.ServiceAccountTokenKey]

		oldValidator, err := k8sutil.GetTokenFolderSecret(t.Context(), handler.kubeClient.CoreV1().Secrets(deployment.GetNamespace()), pod.JWTSecretFolder(deployment.GetName()))
		require.NoError(t, err)

		_, err = oldValidator.Validate(string(oldSecret))
		require.NoError(t, err)

		// Test
		tests.Update(t, handler.kubeClient, handler.client, &extension, func(t *testing.T, obj *permissionApi.ArangoPermissionToken) {
			obj.Spec.Deployment = &sharedApi.Object{
				Name: "example",
			}
		})
		tests.Update(t, handler.kubeClient, handler.client, &jwt, func(t *testing.T, obj *core.Secret) {
			obj.Data = map[string][]byte{
				utilConstants.ActiveJWTKey: []byte(tokenNew),
			}
		})

		require.NoError(t, tests.Handle(handler, tests.NewItem(t, operation.Update, extension)))

		// Refresh
		refresh(t)

		// Validate
		require.True(t, extension.Status.Conditions.IsTrue(permissionApi.SpecValidCondition))
		require.True(t, extension.Status.Conditions.IsTrue(permissionApi.DeploymentFoundCondition))
		require.True(t, extension.Status.Conditions.IsTrue(permissionApi.DeploymentReachableCondition))
		require.True(t, extension.Status.Conditions.IsTrue(permissionApi.ReadyCondition))
		require.NotNil(t, extension.Status.Secret)

		secret := tests.GetObject[*core.Secret](t, handler.kubeClient, handler.client, tests.FakeNamespace, extension.Status.Secret.GetName())

		validator, err := k8sutil.GetTokenFolderSecret(t.Context(), handler.kubeClient.CoreV1().Secrets(deployment.GetNamespace()), pod.JWTSecretFolder(deployment.GetName()))
		require.NoError(t, err)

		require.Contains(t, secret.Data, core.ServiceAccountTokenKey)
		token, err := validator.Validate(string(secret.Data[core.ServiceAccountTokenKey]))
		require.NoError(t, err)
		require.EqualValues(t, token.Claims()[utilToken.ClaimPreferredUsername], extension.Status.User.GetName())
		require.NotEqual(t, oldSecret, secret.Data[core.ServiceAccountTokenKey])

		roles := extractRoles(t, token)
		require.Len(t, roles, 0)

		// Check validation

		_, err = oldValidator.Validate(string(secret.Data[core.ServiceAccountTokenKey]))
		require.EqualError(t, err, "signature is invalid")

		_, err = validator.Validate(string(oldSecret))
		require.EqualError(t, err, "signature is invalid")
	})

	t.Run("New deployment - refresh JWT roles", func(t *testing.T) {
		oldSecret := tests.GetObject[*core.Secret](t, handler.kubeClient, handler.client, tests.FakeNamespace, extension.Status.Secret.GetName()).Data[core.ServiceAccountTokenKey]

		oldValidator, err := k8sutil.GetTokenFolderSecret(t.Context(), handler.kubeClient.CoreV1().Secrets(deployment.GetNamespace()), pod.JWTSecretFolder(deployment.GetName()))
		require.NoError(t, err)

		_, err = oldValidator.Validate(string(oldSecret))
		require.NoError(t, err)

		// Test
		tests.Update(t, handler.kubeClient, handler.client, &extension, func(t *testing.T, obj *permissionApi.ArangoPermissionToken) {
			obj.Spec.Deployment = &sharedApi.Object{
				Name: "example",
			}

			obj.Spec.Roles = []string{
				"role",
			}
		})
		tests.Update(t, handler.kubeClient, handler.client, &jwt, func(t *testing.T, obj *core.Secret) {
			obj.Data = map[string][]byte{
				utilConstants.ActiveJWTKey: []byte(tokenNew),
			}
		})

		require.NoError(t, tests.Handle(handler, tests.NewItem(t, operation.Update, extension)))

		// Refresh
		refresh(t)

		// Validate
		require.True(t, extension.Status.Conditions.IsTrue(permissionApi.SpecValidCondition))
		require.True(t, extension.Status.Conditions.IsTrue(permissionApi.DeploymentFoundCondition))
		require.True(t, extension.Status.Conditions.IsTrue(permissionApi.DeploymentReachableCondition))
		require.True(t, extension.Status.Conditions.IsTrue(permissionApi.ReadyCondition))
		require.NotNil(t, extension.Status.Secret)

		secret := tests.GetObject[*core.Secret](t, handler.kubeClient, handler.client, tests.FakeNamespace, extension.Status.Secret.GetName())

		validator, err := k8sutil.GetTokenFolderSecret(t.Context(), handler.kubeClient.CoreV1().Secrets(deployment.GetNamespace()), pod.JWTSecretFolder(deployment.GetName()))
		require.NoError(t, err)

		require.Contains(t, secret.Data, core.ServiceAccountTokenKey)
		token, err := validator.Validate(string(secret.Data[core.ServiceAccountTokenKey]))
		require.NoError(t, err)
		require.EqualValues(t, token.Claims()[utilToken.ClaimPreferredUsername], extension.Status.User.GetName())
		require.NotEqual(t, oldSecret, secret.Data[core.ServiceAccountTokenKey])

		roles := extractRoles(t, token)
		require.Len(t, roles, 1)
		require.Contains(t, roles, "role")

		// Check validation

		_, err = oldValidator.Validate(string(secret.Data[core.ServiceAccountTokenKey]))
		require.NoError(t, err)

		_, err = validator.Validate(string(oldSecret))
		require.NoError(t, err, "signature is invalid")
	})

	t.Run("New deployment - refresh JWT roles - remove", func(t *testing.T) {
		oldSecret := tests.GetObject[*core.Secret](t, handler.kubeClient, handler.client, tests.FakeNamespace, extension.Status.Secret.GetName()).Data[core.ServiceAccountTokenKey]

		oldValidator, err := k8sutil.GetTokenFolderSecret(t.Context(), handler.kubeClient.CoreV1().Secrets(deployment.GetNamespace()), pod.JWTSecretFolder(deployment.GetName()))
		require.NoError(t, err)

		_, err = oldValidator.Validate(string(oldSecret))
		require.NoError(t, err)

		// Test
		tests.Update(t, handler.kubeClient, handler.client, &extension, func(t *testing.T, obj *permissionApi.ArangoPermissionToken) {
			obj.Spec.Deployment = &sharedApi.Object{
				Name: "example",
			}

			obj.Spec.Roles = []string{}
		})
		tests.Update(t, handler.kubeClient, handler.client, &jwt, func(t *testing.T, obj *core.Secret) {
			obj.Data = map[string][]byte{
				utilConstants.ActiveJWTKey: []byte(tokenNew),
			}
		})

		require.NoError(t, tests.Handle(handler, tests.NewItem(t, operation.Update, extension)))

		// Refresh
		refresh(t)

		// Validate
		require.True(t, extension.Status.Conditions.IsTrue(permissionApi.SpecValidCondition))
		require.True(t, extension.Status.Conditions.IsTrue(permissionApi.DeploymentFoundCondition))
		require.True(t, extension.Status.Conditions.IsTrue(permissionApi.DeploymentReachableCondition))
		require.True(t, extension.Status.Conditions.IsTrue(permissionApi.ReadyCondition))
		require.NotNil(t, extension.Status.Secret)

		secret := tests.GetObject[*core.Secret](t, handler.kubeClient, handler.client, tests.FakeNamespace, extension.Status.Secret.GetName())

		validator, err := k8sutil.GetTokenFolderSecret(t.Context(), handler.kubeClient.CoreV1().Secrets(deployment.GetNamespace()), pod.JWTSecretFolder(deployment.GetName()))
		require.NoError(t, err)

		require.Contains(t, secret.Data, core.ServiceAccountTokenKey)
		token, err := validator.Validate(string(secret.Data[core.ServiceAccountTokenKey]))
		require.NoError(t, err)
		require.EqualValues(t, token.Claims()[utilToken.ClaimPreferredUsername], extension.Status.User.GetName())
		require.NotEqual(t, oldSecret, secret.Data[core.ServiceAccountTokenKey])

		roles := extractRoles(t, token)
		require.Len(t, roles, 0)

		// Check validation

		_, err = oldValidator.Validate(string(secret.Data[core.ServiceAccountTokenKey]))
		require.NoError(t, err)

		_, err = validator.Validate(string(oldSecret))
		require.NoError(t, err, "signature is invalid")
	})

	t.Run("New deployment - change of TTL", func(t *testing.T) {
		oldSecret := tests.GetObject[*core.Secret](t, handler.kubeClient, handler.client, tests.FakeNamespace, extension.Status.Secret.GetName()).Data[core.ServiceAccountTokenKey]

		oldValidator, err := k8sutil.GetTokenFolderSecret(t.Context(), handler.kubeClient.CoreV1().Secrets(deployment.GetNamespace()), pod.JWTSecretFolder(deployment.GetName()))
		require.NoError(t, err)

		_, err = oldValidator.Validate(string(oldSecret))
		require.NoError(t, err)

		// Test
		tests.Update(t, handler.kubeClient, handler.client, &extension, func(t *testing.T, obj *permissionApi.ArangoPermissionToken) {
			obj.Spec.Deployment = &sharedApi.Object{
				Name: "example",
			}

			obj.Spec.TTL = util.NewType(meta.Duration{Duration: 8 * time.Hour})
		})

		require.NoError(t, tests.Handle(handler, tests.NewItem(t, operation.Update, extension)))

		// Refresh
		refresh(t)

		// Validate
		require.True(t, extension.Status.Conditions.IsTrue(permissionApi.SpecValidCondition))
		require.True(t, extension.Status.Conditions.IsTrue(permissionApi.DeploymentFoundCondition))
		require.True(t, extension.Status.Conditions.IsTrue(permissionApi.DeploymentReachableCondition))
		require.True(t, extension.Status.Conditions.IsTrue(permissionApi.ReadyCondition))
		require.NotNil(t, extension.Status.Secret)

		secret := tests.GetObject[*core.Secret](t, handler.kubeClient, handler.client, tests.FakeNamespace, extension.Status.Secret.GetName())

		validator, err := k8sutil.GetTokenFolderSecret(t.Context(), handler.kubeClient.CoreV1().Secrets(deployment.GetNamespace()), pod.JWTSecretFolder(deployment.GetName()))
		require.NoError(t, err)

		require.Contains(t, secret.Data, core.ServiceAccountTokenKey)
		token, err := validator.Validate(string(secret.Data[core.ServiceAccountTokenKey]))
		require.NoError(t, err)
		require.EqualValues(t, token.Claims()[utilToken.ClaimPreferredUsername], extension.Status.User.GetName())
		require.NotEqual(t, oldSecret, secret.Data[core.ServiceAccountTokenKey])

		roles := extractRoles(t, token)
		require.Len(t, roles, 0)

		// Check validation

		_, err = oldValidator.Validate(string(secret.Data[core.ServiceAccountTokenKey]))
		require.NoError(t, err)

		_, err = validator.Validate(string(oldSecret))
		require.NoError(t, err, "signature is invalid")
	})

	t.Run("New deployment - still valid", func(t *testing.T) {
		oldSecret := tests.GetObject[*core.Secret](t, handler.kubeClient, handler.client, tests.FakeNamespace, extension.Status.Secret.GetName()).Data[core.ServiceAccountTokenKey]

		oldValidator, err := k8sutil.GetTokenFolderSecret(t.Context(), handler.kubeClient.CoreV1().Secrets(deployment.GetNamespace()), pod.JWTSecretFolder(deployment.GetName()))
		require.NoError(t, err)

		_, err = oldValidator.Validate(string(oldSecret))
		require.NoError(t, err)

		// Test
		tests.Update(t, handler.kubeClient, handler.client, &extension, func(t *testing.T, obj *permissionApi.ArangoPermissionToken) {
			obj.Spec.Deployment = &sharedApi.Object{
				Name: "example",
			}
		})

		tests.UpdateStatus(t, handler.kubeClient, handler.client, &extension, func(t *testing.T, obj *permissionApi.ArangoPermissionToken) {
			obj.Status.Refresh = meta.NewTime(time.Now().Add(time.Minute))
		})

		require.NoError(t, tests.Handle(handler, tests.NewItem(t, operation.Update, extension)))

		// Refresh
		refresh(t)

		// Validate
		require.True(t, extension.Status.Conditions.IsTrue(permissionApi.SpecValidCondition))
		require.True(t, extension.Status.Conditions.IsTrue(permissionApi.DeploymentFoundCondition))
		require.True(t, extension.Status.Conditions.IsTrue(permissionApi.DeploymentReachableCondition))
		require.True(t, extension.Status.Conditions.IsTrue(permissionApi.ReadyCondition))
		require.NotNil(t, extension.Status.Secret)

		secret := tests.GetObject[*core.Secret](t, handler.kubeClient, handler.client, tests.FakeNamespace, extension.Status.Secret.GetName())

		validator, err := k8sutil.GetTokenFolderSecret(t.Context(), handler.kubeClient.CoreV1().Secrets(deployment.GetNamespace()), pod.JWTSecretFolder(deployment.GetName()))
		require.NoError(t, err)

		require.Contains(t, secret.Data, core.ServiceAccountTokenKey)
		token, err := validator.Validate(string(secret.Data[core.ServiceAccountTokenKey]))
		require.NoError(t, err)
		require.EqualValues(t, token.Claims()[utilToken.ClaimPreferredUsername], extension.Status.User.GetName())
		require.Equal(t, oldSecret, secret.Data[core.ServiceAccountTokenKey])

		roles := extractRoles(t, token)
		require.Len(t, roles, 0)
	})

	t.Run("New deployment - expiring", func(t *testing.T) {
		oldSecret := tests.GetObject[*core.Secret](t, handler.kubeClient, handler.client, tests.FakeNamespace, extension.Status.Secret.GetName()).Data[core.ServiceAccountTokenKey]

		oldValidator, err := k8sutil.GetTokenFolderSecret(t.Context(), handler.kubeClient.CoreV1().Secrets(deployment.GetNamespace()), pod.JWTSecretFolder(deployment.GetName()))
		require.NoError(t, err)

		_, err = oldValidator.Validate(string(oldSecret))
		require.NoError(t, err)

		// Test
		tests.Update(t, handler.kubeClient, handler.client, &extension, func(t *testing.T, obj *permissionApi.ArangoPermissionToken) {
			obj.Spec.Deployment = &sharedApi.Object{
				Name: "example",
			}
		})

		tests.UpdateStatus(t, handler.kubeClient, handler.client, &extension, func(t *testing.T, obj *permissionApi.ArangoPermissionToken) {
			obj.Status.Refresh = meta.NewTime(time.Now().Add(-time.Minute))
		})

		require.NoError(t, tests.Handle(handler, tests.NewItem(t, operation.Update, extension)))

		// Refresh
		refresh(t)

		// Validate
		require.True(t, extension.Status.Conditions.IsTrue(permissionApi.SpecValidCondition))
		require.True(t, extension.Status.Conditions.IsTrue(permissionApi.DeploymentFoundCondition))
		require.True(t, extension.Status.Conditions.IsTrue(permissionApi.DeploymentReachableCondition))
		require.True(t, extension.Status.Conditions.IsTrue(permissionApi.ReadyCondition))
		require.NotNil(t, extension.Status.Secret)

		secret := tests.GetObject[*core.Secret](t, handler.kubeClient, handler.client, tests.FakeNamespace, extension.Status.Secret.GetName())

		validator, err := k8sutil.GetTokenFolderSecret(t.Context(), handler.kubeClient.CoreV1().Secrets(deployment.GetNamespace()), pod.JWTSecretFolder(deployment.GetName()))
		require.NoError(t, err)

		require.Contains(t, secret.Data, core.ServiceAccountTokenKey)
		token, err := validator.Validate(string(secret.Data[core.ServiceAccountTokenKey]))
		require.NoError(t, err)
		require.EqualValues(t, token.Claims()[utilToken.ClaimPreferredUsername], extension.Status.User.GetName())
		require.NotEqual(t, oldSecret, secret.Data[core.ServiceAccountTokenKey])

		roles := extractRoles(t, token)
		require.Len(t, roles, 0)

		// Check validation

		_, err = oldValidator.Validate(string(secret.Data[core.ServiceAccountTokenKey]))
		require.NoError(t, err)

		_, err = validator.Validate(string(oldSecret))
		require.NoError(t, err, "signature is invalid")
	})
}
