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
	"context"
	"fmt"
	goStrings "strings"
	"time"

	"github.com/dchest/uniuri"
	core "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/client-go/kubernetes"

	"github.com/arangodb/go-driver/v2/arangodb"
	"github.com/arangodb/go-driver/v2/arangodb/shared"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	permissionApi "github.com/arangodb/kube-arangodb/pkg/apis/permission/v1alpha1"
	sharedApi "github.com/arangodb/kube-arangodb/pkg/apis/shared/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/pod"
	arangoClientSet "github.com/arangodb/kube-arangodb/pkg/generated/clientset/versioned"
	"github.com/arangodb/kube-arangodb/pkg/logging"
	operator "github.com/arangodb/kube-arangodb/pkg/operatorV2"
	"github.com/arangodb/kube-arangodb/pkg/operatorV2/event"
	"github.com/arangodb/kube-arangodb/pkg/operatorV2/operation"
	"github.com/arangodb/kube-arangodb/pkg/util"
	utilConstants "github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/kerrors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/patcher"
	utilToken "github.com/arangodb/kube-arangodb/pkg/util/token"
)

var logger = logging.Global().RegisterAndGetLogger("permission-token-operator", logging.Info)

type handler struct {
	client     arangoClientSet.Interface
	kubeClient kubernetes.Interface

	eventRecorder event.RecorderInstance

	operator operator.Operator

	provider clientProvider
}

func (h *handler) Name() string {
	return Kind()
}

func (h *handler) Handle(ctx context.Context, item operation.Item) error {
	// Get Backup object. It also covers NotFound case
	object, err := util.WithKubernetesContextTimeoutP2A2(ctx, h.client.PermissionV1alpha1().ArangoPermissionTokens(item.Namespace).Get, item.Name, meta.GetOptions{})
	if err != nil {
		if apiErrors.IsNotFound(err) {
			return nil
		}

		return err
	}

	if object.GetDeletionTimestamp() != nil {
		// We are deleting the object
		finalizer, err := h.finalizer(ctx, object)
		if err != nil {
			return err
		}

		if finalizer != "" {
			if changed, err := patcher.EnsureFinalizersGone(ctx, h.client.PermissionV1alpha1().ArangoPermissionTokens(item.Namespace), object, finalizer); err != nil {
				return err
			} else if changed {
				return operator.Reconcile("Finalizers updated")
			}
		}

		return operator.Reconcile("Finalizers pending removal")
	}

	if changed, err := patcher.EnsureFinalizersPresent(ctx, h.client.PermissionV1alpha1().ArangoPermissionTokens(item.Namespace), object, permissionApi.FinalizerArangoPermissionTokenUser); err != nil {
		return err
	} else if changed {
		return operator.Reconcile("Finalizers updated")
	}

	status := object.Status.DeepCopy()

	changed, reconcileErr := operator.HandleP3WithStop(ctx, item, object, status, h.handle)
	if reconcileErr != nil && !operator.IsReconcile(reconcileErr) {
		logger.Err(reconcileErr).Warn("Fail for %s %s/%s",
			item.Kind,
			item.Namespace,
			item.Name)

		return reconcileErr
	}

	if !changed {
		return reconcileErr
	}

	logger.Debug("Updating %s %s/%s",
		item.Kind,
		item.Namespace,
		item.Name)

	if _, err := operator.WithArangoPermissionTokenUpdateStatusInterfaceRetry(context.Background(), h.client.PermissionV1alpha1().ArangoPermissionTokens(object.GetNamespace()), object, *status, meta.UpdateOptions{}); err != nil {
		return err
	}

	return reconcileErr
}

func (h *handler) CanBeHandled(item operation.Item) bool {
	return item.Group == Group() &&
		utilConstants.Version(Version()).IsCompatible(utilConstants.Version(item.Version)) &&
		item.Kind == Kind()
}

func (h *handler) finalizer(ctx context.Context, extension *permissionApi.ArangoPermissionToken) (string, error) {
	for _, finalizer := range extension.GetFinalizers() {
		switch finalizer {
		case permissionApi.FinalizerArangoPermissionTokenUser:
			if err := h.finalizerUserRemoval(ctx, extension); err != nil {
				return "", err
			}

			return permissionApi.FinalizerArangoPermissionTokenUser, nil
		}
	}

	return "", nil
}

func (h *handler) finalizerUserRemoval(ctx context.Context, extension *permissionApi.ArangoPermissionToken) error {
	if extension.Status.Deployment == nil || extension.Status.User == nil {
		return nil
	}

	depl, err := h.client.DatabaseV1().ArangoDeployments(extension.GetNamespace()).Get(ctx, extension.Status.Deployment.GetName(), meta.GetOptions{})
	if err != nil {
		if kerrors.IsNotFound(err) {
			return nil
		}

		return err
	}

	if !extension.Status.Deployment.Equals(depl) {
		logger.Warn("Deleting of the user not allowed due to change in UUID")
		return nil
	}

	client, err := h.provider.ArangoClient(ctx, h.kubeClient, depl)
	if err != nil {
		logger.Warn("Fail to get client for deleting user")
		return err
	}

	if err := client.RemoveUser(ctx, extension.Status.User.GetName()); err != nil {
		if shared.IsNotFound(err) {
			return nil
		}

		return err
	}

	return nil
}

func (h *handler) handle(ctx context.Context, item operation.Item, extension *permissionApi.ArangoPermissionToken, status *permissionApi.ArangoPermissionTokenStatus) (bool, error) {
	return operator.HandleSharedP3WithCondition(ctx, &status.Conditions, permissionApi.ReadyCondition, item, extension, status, h.HandleSpecValidity, h.HandleDeployment)
}

func (h *handler) HandleSpecValidity(ctx context.Context, item operation.Item, extension *permissionApi.ArangoPermissionToken, status *permissionApi.ArangoPermissionTokenStatus) (bool, error) {
	if err := extension.Spec.Validate(); err != nil {
		// We have received an error in the spec!

		logger.Err(err).Warn("Invalid Spec on %s", item.String())

		if status.Conditions.Update(permissionApi.SpecValidCondition, false, "Spec is invalid", "Spec is invalid") {
			return true, operator.Stop("Invalid spec")
		}
		return false, operator.Stop("Invalid spec")
	}

	if status.Conditions.Update(permissionApi.SpecValidCondition, true, "Spec is valid", "Spec is valid") {
		logger.WrapObj(item).Debug("Spec is valid")
		return true, nil
	}

	return false, nil
}

func (h *handler) HandleDeployment(ctx context.Context, item operation.Item, extension *permissionApi.ArangoPermissionToken, status *permissionApi.ArangoPermissionTokenStatus) (bool, error) {
	logger := logger.WrapObj(item).Str("deployment", extension.Spec.Deployment.GetName())

	if status.Deployment == nil {
		depl, err := h.client.DatabaseV1().ArangoDeployments(extension.GetNamespace()).Get(ctx, extension.Spec.Deployment.GetName(), meta.GetOptions{})
		if err != nil {
			if !apiErrors.IsNotFound(err) {
				return false, err
			}

			if status.Conditions.Update(permissionApi.DeploymentFoundCondition, false, "Deployment not found", "Deployment not found") {
				logger.Warn("Deployment Not Found")
				return true, operator.Reconcile("Conditions updated")
			}

			return false, operator.Stop("Missing deployment")
		}

		status.Deployment = util.NewType(sharedApi.NewObject(depl))

		logger.Info("Deployment Accepted")

		return true, operator.Reconcile("Deployment Accepted")
	}

	depl, err := h.client.DatabaseV1().ArangoDeployments(extension.GetNamespace()).Get(ctx, extension.Status.Deployment.GetName(), meta.GetOptions{})
	if err != nil {
		if !apiErrors.IsNotFound(err) {
			return false, err
		}

		if status.Conditions.Update(permissionApi.DeploymentFoundCondition, false, "Deployment not found", "Deployment not found") {
			logger.Warn("Deployment Not Found")
			return true, nil
		}

		return false, operator.Stop("Missing deployment, recreate object")
	}

	if !extension.Status.Deployment.Equals(depl) {
		if status.Conditions.Update(permissionApi.DeploymentFoundCondition, false, "Deployment changed", "Deployment changed") {
			logger.Warn("Deployment Changed")
			return true, operator.Reconcile("Conditions updated")
		}

		return false, operator.Stop("Invalid deployment, recreate object")
	}

	if status.Conditions.Update(permissionApi.DeploymentFoundCondition, true, "Deployment found", "Deployment found") {
		logger.Debug("Deployment Found")
		return true, nil
	}

	return operator.HandleP4(ctx, item, extension, status, depl, h.HandleDeploymentConnection)
}

func (h *handler) HandleDeploymentConnection(ctx context.Context, item operation.Item, extension *permissionApi.ArangoPermissionToken, status *permissionApi.ArangoPermissionTokenStatus, depl *api.ArangoDeployment) (bool, error) {
	conn, err := h.provider.ArangoClient(ctx, h.kubeClient, depl)
	if err != nil {
		return false, err
	}

	_, err = conn.Version(ctx)
	if err != nil {
		logger.Err(err).Warn("Deployment is not reachable")

		if status.Conditions.Update(permissionApi.DeploymentReachableCondition, false, "Deployment not reachable", "Deployment not reachable") {
			return true, operator.Reconcile("Conditions updated")
		}

		return false, operator.Stop("Deployment not reachable")
	}

	if status.Conditions.Update(permissionApi.DeploymentReachableCondition, true, "Deployment reachable", "Deployment reachable") {
		return true, operator.Reconcile("Conditions updated")
	}

	return operator.HandleP5(ctx, item, extension, status, depl, conn, h.HandleArangoDBUser)
}

func (h *handler) HandleArangoDBUser(ctx context.Context, item operation.Item, extension *permissionApi.ArangoPermissionToken, status *permissionApi.ArangoPermissionTokenStatus, depl *api.ArangoDeployment, conn arangodb.Client) (bool, error) {
	if status.User == nil {
		name := fmt.Sprintf("operator-%s-%s", extension.GetName(), goStrings.ToLower(uniuri.NewLen(6)))
		logger.Str("name", name).Info("Create ArangoDB User")
		if _, err := conn.User(ctx, name); err != nil {
			if !shared.IsNotFound(err) {
				return false, err
			}
		} else {
			return false, operator.Reconcile("ArangoDB User name used")
		}

		user, err := conn.CreateUser(ctx, name, &arangodb.UserOptions{
			Password: string(uuid.NewUUID()),
			Active:   util.NewType(true),
		})
		if err != nil {
			return false, err
		}

		status.User = &sharedApi.Object{
			Name: user.Name(),
		}

		return true, operator.Reconcile("ArangoDB User created")
	}

	if _, err := conn.User(ctx, status.User.GetName()); err != nil {
		if !shared.IsNotFound(err) {
			return false, err
		}

		logger.Str("name", status.User.GetName()).Warn("User Not Found, Recreate")
		status.User = nil
		return true, operator.Reconcile("ArangoDB User gone")
	}

	return operator.HandleP4(ctx, item, extension, status, depl, h.HandleArangoSecret)
}

func (h *handler) HandleArangoSecret(ctx context.Context, item operation.Item, extension *permissionApi.ArangoPermissionToken, status *permissionApi.ArangoPermissionTokenStatus, depl *api.ArangoDeployment) (bool, error) {
	if status.Secret == nil {
		name := fmt.Sprintf("%s-%s", extension.GetName(), goStrings.ToLower(uniuri.NewLen(6)))

		secret := &core.Secret{
			ObjectMeta: meta.ObjectMeta{
				Name:      name,
				Namespace: extension.GetNamespace(),
				OwnerReferences: []meta.OwnerReference{
					extension.AsOwner(),
				},
			},
			Type: core.SecretTypeOpaque,
		}
		logger.Str("name", name).Info("Create ArangoDB Secret")

		secret, err := h.kubeClient.CoreV1().Secrets(extension.GetNamespace()).Create(ctx, secret, meta.CreateOptions{})
		if err != nil {
			return false, err
		}

		status.Secret = util.NewType(sharedApi.NewObject(secret))

		return true, operator.Reconcile("Secret Created")
	}

	secret, err := h.kubeClient.CoreV1().Secrets(extension.GetNamespace()).Get(ctx, status.Secret.GetName(), meta.GetOptions{})
	if err != nil {
		if apiErrors.IsNotFound(err) {
			logger.Str("name", status.Secret.GetName()).Warn("Secret Not Found, Recreate")
			status.Secret = nil
			return true, operator.Reconcile("Secret Gone")
		}

		return false, err
	}

	if !status.Secret.Equals(secret) {
		logger.Str("name", status.Secret.GetName()).Warn("Secret Changed, Recreate")
		status.Secret = nil
		return true, operator.Reconcile("Secret Changed")
	}

	// Add the JWT Token

	if !depl.GetAcceptedSpec().Authentication.IsAuthenticated() {
		// Unauthenticated
		if v, ok := secret.Data[core.ServiceAccountTokenKey]; !ok || len(v) != 0 {
			if _, _, err := patcher.Patcher[*core.Secret](ctx, h.kubeClient.CoreV1().Secrets(secret.GetNamespace()), secret, meta.PatchOptions{},
				patcher.PatchSecretData(map[string][]byte{
					core.ServiceAccountTokenKey: {},
				}),
			); err != nil {
				return false, err
			}
		}

		// Done
		return false, nil
	}

	secretManager, err := k8sutil.GetTokenFolderSecret(ctx, h.kubeClient.CoreV1().Secrets(depl.GetNamespace()), pod.JWTSecretFolder(depl.GetName()))
	if err != nil {
		return false, err
	}

	hash := util.SHA256FromStringArray(status.User.GetName(), secretManager.SigningHash(), extension.Spec.GetTTL().String(), util.SHA256FromStringArray(extension.Spec.Roles...))

	if status.Secret.GetChecksum() == hash && time.Now().Before(status.Refresh.Time) {
		// Done
		return false, nil
	}

	logger.Info("Generating Token")

	token, err := utilToken.NewClaims().With(
		utilToken.WithKey("id", uuid.NewUUID()),
		utilToken.WithDefaultClaims(),
		utilToken.WithUsername(status.User.GetName()),
		utilToken.WithCurrentIAT(),
		utilToken.WithExp(time.Now().Add(extension.Spec.GetTTL())),
		utilToken.WithRoles(extension.Spec.Roles...),
	).Sign(secretManager)
	if err != nil {
		return false, err
	}

	if _, _, err := patcher.Patcher[*core.Secret](ctx, h.kubeClient.CoreV1().Secrets(secret.GetNamespace()), secret, meta.PatchOptions{},
		patcher.PatchSecretData(map[string][]byte{
			core.ServiceAccountTokenKey: []byte(token),
		}),
	); err != nil {
		return false, err
	}

	status.Secret.Checksum = util.NewType(hash)
	status.Refresh = meta.NewTime(time.Now().Add(extension.Spec.GetTTL() / 2))

	h.eventRecorder.Normal(extension, "Token Generated", "Token Generated with Hash %s and Expiration %s", hash, status.Refresh.Time.String())

	return true, nil
}
