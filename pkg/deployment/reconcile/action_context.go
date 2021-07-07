//
// DISCLAIMER
//
// Copyright 2020-2021 ArangoDB GmbH, Cologne, Germany
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
// Author Ewout Prangsma
// Author Tomasz Mielech
//

package reconcile

import (
	"context"

	"github.com/arangodb/kube-arangodb/pkg/deployment/resources"

	inspectorInterface "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"

	backupApi "github.com/arangodb/kube-arangodb/pkg/apis/backup/v1"

	"github.com/arangodb/go-driver/agency"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	core "k8s.io/api/core/v1"

	"github.com/arangodb/arangosync-client/client"
	driver "github.com/arangodb/go-driver"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
)

// ActionContext provides methods to the Action implementations
// to control their context.
type ActionContext interface {
	resources.DeploymentStatusUpdate
	resources.DeploymentAgencyMaintenance
	resources.ArangoMemberContext
	resources.DeploymentPodRenderer

	// GetAPIObject returns the deployment as k8s object.
	GetAPIObject() k8sutil.APIObject
	// Gets the specified mode of deployment
	GetMode() api.DeploymentMode
	// GetDatabaseClient returns a cached client for the entire database (cluster coordinators or single server),
	// creating one if needed.
	GetDatabaseClient(ctx context.Context) (driver.Client, error)
	// GetServerClient returns a cached client for a specific server.
	GetServerClient(ctx context.Context, group api.ServerGroup, id string) (driver.Client, error)
	// GetAgencyClients returns a client connection for every agency member.
	GetAgencyClients(ctx context.Context) ([]driver.Connection, error)
	// GetAgency returns a connection to the entire agency.
	GetAgency(ctx context.Context) (agency.Agency, error)
	// GetSyncServerClient returns a cached client for a specific arangosync server.
	GetSyncServerClient(ctx context.Context, group api.ServerGroup, id string) (client.API, error)
	// CreateEvent creates a given event.
	// On error, the error is logged.
	CreateEvent(evt *k8sutil.Event)
	// GetMemberStatusByID returns the current member status
	// for the member with given id.
	// Returns member status, true when found, or false
	// when no such member is found.
	GetMemberStatusByID(id string) (api.MemberStatus, bool)
	// GetMemberStatusAndGroupByID returns the current member status and group
	// for the member with given id.
	// Returns member status, true when found, or false
	// when no such member is found.
	GetMemberStatusAndGroupByID(id string) (api.MemberStatus, api.ServerGroup, bool)
	// CreateMember adds a new member to the given group.
	// If ID is non-empty, it will be used, otherwise a new ID is created.
	CreateMember(ctx context.Context, group api.ServerGroup, id string) (string, error)
	// UpdateMember updates the deployment status wrt the given member.
	UpdateMember(ctx context.Context, member api.MemberStatus) error
	// RemoveMemberByID removes a member with given id.
	RemoveMemberByID(ctx context.Context, id string) error
	// GetPod returns pod.
	GetPod(ctx context.Context, podName string) (*core.Pod, error)
	// DeletePod deletes a pod with given name in the namespace
	// of the deployment. If the pod does not exist, the error is ignored.
	DeletePod(ctx context.Context, podName string) error
	// DeletePvc deletes a persistent volume claim with given name in the namespace
	// of the deployment. If the pvc does not exist, the error is ignored.
	DeletePvc(ctx context.Context, pvcName string) error
	// GetPvc returns PVC info about PVC with given name in the namespace
	// of the deployment.
	GetPvc(ctx context.Context, pvcName string) (*core.PersistentVolumeClaim, error)
	// UpdatePvc update PVC with given name in the namespace
	// of the deployment.
	UpdatePvc(ctx context.Context, pvc *core.PersistentVolumeClaim) error
	// RemovePodFinalizers removes all the finalizers from the Pod with given name in the namespace
	// of the deployment. If the pod does not exist, the error is ignored.
	RemovePodFinalizers(ctx context.Context, podName string) error
	// DeleteTLSKeyfile removes the Secret containing the TLS keyfile for the given member.
	// If the secret does not exist, the error is ignored.
	DeleteTLSKeyfile(ctx context.Context, group api.ServerGroup, member api.MemberStatus) error
	// DeleteTLSCASecret removes the Secret containing the TLS CA certificate.
	DeleteTLSCASecret(ctx context.Context) error
	// GetImageInfo returns the image info for an image with given name.
	// Returns: (info, infoFound)
	GetImageInfo(imageName string) (api.ImageInfo, bool)
	// GetImageInfo returns the image info for an current image.
	// Returns: (info, infoFound)
	GetCurrentImageInfo() (api.ImageInfo, bool)
	// SetCurrentImage changes the CurrentImage field in the deployment
	// status to the given image.
	SetCurrentImage(ctx context.Context, imageInfo api.ImageInfo) error
	// GetDeploymentHealth returns a copy of the latest known state of cluster health
	GetDeploymentHealth() (driver.ClusterHealth, error)
	// GetShardSyncStatus returns true if all shards are in sync
	GetShardSyncStatus() bool
	// InvalidateSyncStatus resets the sync state to false and triggers an inspection
	InvalidateSyncStatus()
	// GetSpec returns a copy of the spec
	GetSpec() api.DeploymentSpec
	// GetStatus returns a copy of the status
	GetStatus() api.DeploymentStatus
	// DisableScalingCluster disables scaling DBservers and coordinators
	DisableScalingCluster(ctx context.Context) error
	// EnableScalingCluster enables scaling DBservers and coordinators
	EnableScalingCluster(ctx context.Context) error
	// WithStatusUpdate update status of ArangoDeployment with defined modifier. If action returns True action is taken
	UpdateClusterCondition(ctx context.Context, conditionType api.ConditionType, status bool, reason, message string) error
	SecretsInterface() k8sutil.SecretInterface
	// GetBackup receives information about a backup resource
	GetBackup(ctx context.Context, backup string) (*backupApi.ArangoBackup, error)
	// GetName receives information about a deployment name
	GetName() string
	// GetCachedStatus current cached state of deployment
	GetCachedStatus() inspectorInterface.Inspector
	// SelectImage select currently used image by pod
	SelectImage(spec api.DeploymentSpec, status api.DeploymentStatus) (api.ImageInfo, bool)
}

// newActionContext creates a new ActionContext implementation.
func newActionContext(log zerolog.Logger, context Context, cachedStatus inspectorInterface.Inspector) ActionContext {
	return &actionContext{
		log:          log,
		context:      context,
		cachedStatus: cachedStatus,
	}
}

// actionContext implements ActionContext
type actionContext struct {
	context      Context
	log          zerolog.Logger
	cachedStatus inspectorInterface.Inspector
}

func (ac *actionContext) GetAgencyMaintenanceMode(ctx context.Context) (bool, error) {
	return ac.context.GetAgencyMaintenanceMode(ctx)
}

func (ac *actionContext) SetAgencyMaintenanceMode(ctx context.Context, enabled bool) error {
	return ac.context.SetAgencyMaintenanceMode(ctx, enabled)
}

func (ac *actionContext) WithArangoMemberUpdate(ctx context.Context, namespace, name string, action resources.ArangoMemberUpdateFunc) error {
	return ac.context.WithArangoMemberUpdate(ctx, namespace, name, action)
}

func (ac *actionContext) WithArangoMemberStatusUpdate(ctx context.Context, namespace, name string, action resources.ArangoMemberStatusUpdateFunc) error {
	return ac.context.WithArangoMemberStatusUpdate(ctx, namespace, name, action)
}

func (ac *actionContext) RenderPodForMember(ctx context.Context, cachedStatus inspectorInterface.Inspector, spec api.DeploymentSpec, status api.DeploymentStatus, memberID string, imageInfo api.ImageInfo) (*core.Pod, error) {
	return ac.context.RenderPodForMember(ctx, cachedStatus, spec, status, memberID, imageInfo)
}

func (ac *actionContext) RenderPodTemplateForMember(ctx context.Context, cachedStatus inspectorInterface.Inspector, spec api.DeploymentSpec, status api.DeploymentStatus, memberID string, imageInfo api.ImageInfo) (*core.PodTemplateSpec, error) {
	return ac.context.RenderPodTemplateForMember(ctx, cachedStatus, spec, status, memberID, imageInfo)
}

func (ac *actionContext) SelectImage(spec api.DeploymentSpec, status api.DeploymentStatus) (api.ImageInfo, bool) {
	return ac.context.SelectImage(spec, status)
}

func (ac *actionContext) GetCachedStatus() inspectorInterface.Inspector {
	return ac.cachedStatus
}

func (ac *actionContext) GetName() string {
	return ac.context.GetName()
}

func (ac *actionContext) GetStatus() api.DeploymentStatus {
	a, _ := ac.context.GetStatus()

	s := a.DeepCopy()

	return *s
}

func (ac *actionContext) GetBackup(ctx context.Context, backup string) (*backupApi.ArangoBackup, error) {
	return ac.context.GetBackup(ctx, backup)
}

func (ac *actionContext) WithStatusUpdate(ctx context.Context, action resources.DeploymentStatusUpdateFunc, force ...bool) error {
	return ac.context.WithStatusUpdate(ctx, action, force...)
}

func (ac *actionContext) SecretsInterface() k8sutil.SecretInterface {
	return ac.context.SecretsInterface()
}

func (ac *actionContext) GetShardSyncStatus() bool {
	return ac.context.GetShardSyncStatus()
}

func (ac *actionContext) UpdateClusterCondition(ctx context.Context, conditionType api.ConditionType, status bool, reason, message string) error {
	return ac.context.WithStatusUpdate(ctx, func(s *api.DeploymentStatus) bool {
		return s.Conditions.Update(conditionType, status, reason, message)
	})
}

func (ac *actionContext) GetAPIObject() k8sutil.APIObject {
	return ac.context.GetAPIObject()
}

func (ac *actionContext) UpdatePvc(ctx context.Context, pvc *core.PersistentVolumeClaim) error {
	return ac.context.UpdatePvc(ctx, pvc)
}

func (ac *actionContext) CreateEvent(evt *k8sutil.Event) {
	ac.context.CreateEvent(evt)
}

func (ac *actionContext) GetPvc(ctx context.Context, pvcName string) (*core.PersistentVolumeClaim, error) {
	return ac.context.GetPvc(ctx, pvcName)
}

// Gets the specified mode of deployment
func (ac *actionContext) GetMode() api.DeploymentMode {
	return ac.context.GetSpec().GetMode()
}

func (ac *actionContext) GetSpec() api.DeploymentSpec {
	return ac.context.GetSpec()
}

// GetDeploymentHealth returns a copy of the latest known state of cluster health
func (ac *actionContext) GetDeploymentHealth() (driver.ClusterHealth, error) {
	return ac.context.GetDeploymentHealth()
}

// GetDatabaseClient returns a cached client for the entire database (cluster coordinators or single server),
// creating one if needed.
func (ac *actionContext) GetDatabaseClient(ctx context.Context) (driver.Client, error) {
	c, err := ac.context.GetDatabaseClient(ctx)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return c, nil
}

// GetServerClient returns a cached client for a specific server.
func (ac *actionContext) GetServerClient(ctx context.Context, group api.ServerGroup, id string) (driver.Client, error) {
	c, err := ac.context.GetServerClient(ctx, group, id)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return c, nil
}

// GetAgencyClients returns a client connection for every agency member.
func (ac *actionContext) GetAgencyClients(ctx context.Context) ([]driver.Connection, error) {
	c, err := ac.context.GetAgencyClients(ctx, nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return c, nil
}

// GetAgency returns a connection to the entire agency.
func (ac *actionContext) GetAgency(ctx context.Context) (agency.Agency, error) {
	a, err := ac.context.GetAgency(ctx)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return a, nil
}

// GetSyncServerClient returns a cached client for a specific arangosync server.
func (ac *actionContext) GetSyncServerClient(ctx context.Context, group api.ServerGroup, id string) (client.API, error) {
	c, err := ac.context.GetSyncServerClient(ctx, group, id)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return c, nil
}

// GetMemberStatusByID returns the current member status
// for the member with given id.
// Returns member status, true when found, or false
// when no such member is found.
func (ac *actionContext) GetMemberStatusByID(id string) (api.MemberStatus, bool) {
	status, _ := ac.context.GetStatus()
	m, _, ok := status.Members.ElementByID(id)
	return m, ok
}

// GetMemberStatusAndGroupByID returns the current member status and group
// for the member with given id.
// Returns member status, true when found, or false
// when no such member is found.
func (ac *actionContext) GetMemberStatusAndGroupByID(id string) (api.MemberStatus, api.ServerGroup, bool) {
	status, _ := ac.context.GetStatus()
	return status.Members.ElementByID(id)
}

// CreateMember adds a new member to the given group.
// If ID is non-empty, it will be used, otherwise a new ID is created.
func (ac *actionContext) CreateMember(ctx context.Context, group api.ServerGroup, id string) (string, error) {
	result, err := ac.context.CreateMember(ctx, group, id)
	if err != nil {
		return "", errors.WithStack(err)
	}
	return result, nil
}

// UpdateMember updates the deployment status wrt the given member.
func (ac *actionContext) UpdateMember(ctx context.Context, member api.MemberStatus) error {
	status, lastVersion := ac.context.GetStatus()
	_, group, found := status.Members.ElementByID(member.ID)
	if !found {
		return errors.WithStack(errors.Newf("Member %s not found", member.ID))
	}
	if err := status.Members.Update(member, group); err != nil {
		return errors.WithStack(err)
	}
	if err := ac.context.UpdateStatus(ctx, status, lastVersion); err != nil {
		log.Debug().Err(err).Msg("Updating CR status failed")
		return errors.WithStack(err)
	}
	return nil
}

// RemoveMemberByID removes a member with given id.
func (ac *actionContext) RemoveMemberByID(ctx context.Context, id string) error {
	status, lastVersion := ac.context.GetStatus()
	_, group, found := status.Members.ElementByID(id)
	if !found {
		return nil
	}
	if err := status.Members.RemoveByID(id, group); err != nil {
		log.Debug().Err(err).Str("group", group.AsRole()).Msg("Failed to remove member")
		return errors.WithStack(err)
	}
	// Save removed member
	if err := ac.context.UpdateStatus(ctx, status, lastVersion); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// GetPod returns pod.
func (ac *actionContext) GetPod(ctx context.Context, podName string) (*core.Pod, error) {
	if pod, err := ac.context.GetPod(ctx, podName); err != nil {
		return nil, errors.WithStack(err)
	} else {
		return pod, nil
	}
}

// DeletePod deletes a pod with given name in the namespace
// of the deployment. If the pod does not exist, the error is ignored.
func (ac *actionContext) DeletePod(ctx context.Context, podName string) error {
	if err := ac.context.DeletePod(ctx, podName); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// DeletePvc deletes a persistent volume claim with given name in the namespace
// of the deployment. If the pvc does not exist, the error is ignored.
func (ac *actionContext) DeletePvc(ctx context.Context, pvcName string) error {
	if err := ac.context.DeletePvc(ctx, pvcName); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// RemovePodFinalizers removes all the finalizers from the Pod with given name in the namespace
// of the deployment. If the pod does not exist, the error is ignored.
func (ac *actionContext) RemovePodFinalizers(ctx context.Context, podName string) error {
	if err := ac.context.RemovePodFinalizers(ctx, podName); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// DeleteTLSKeyfile removes the Secret containing the TLS keyfile for the given member.
// If the secret does not exist, the error is ignored.
func (ac *actionContext) DeleteTLSKeyfile(ctx context.Context, group api.ServerGroup, member api.MemberStatus) error {
	if err := ac.context.DeleteTLSKeyfile(ctx, group, member); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// DeleteTLSCASecret removes the Secret containing the TLS CA certificate.
func (ac *actionContext) DeleteTLSCASecret(ctx context.Context) error {
	spec := ac.context.GetSpec().TLS
	if !spec.IsSecure() {
		return nil
	}
	secretName := spec.GetCASecretName()
	if secretName == "" {
		return nil
	}
	// Remove secret hash, since it is going to change
	status, lastVersion := ac.context.GetStatus()
	if status.SecretHashes != nil {
		status.SecretHashes.TLSCA = ""
		if err := ac.context.UpdateStatus(ctx, status, lastVersion); err != nil {
			return errors.WithStack(err)
		}
	}
	// Do delete the secret
	if err := ac.context.DeleteSecret(secretName); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// GetImageInfo returns the image info for an image with given name.
// Returns: (info, infoFound)
func (ac *actionContext) GetImageInfo(imageName string) (api.ImageInfo, bool) {
	status, _ := ac.context.GetStatus()
	return status.Images.GetByImage(imageName)
}

// GetImageInfo returns the image info for an current image.
// Returns: (info, infoFound)
func (ac *actionContext) GetCurrentImageInfo() (api.ImageInfo, bool) {
	status, _ := ac.context.GetStatus()

	if status.CurrentImage == nil {
		return api.ImageInfo{}, false
	}

	return *status.CurrentImage, true
}

// SetCurrentImage changes the CurrentImage field in the deployment
// status to the given image.
func (ac *actionContext) SetCurrentImage(ctx context.Context, imageInfo api.ImageInfo) error {
	return ac.context.WithStatusUpdate(ctx, func(s *api.DeploymentStatus) bool {
		if s.CurrentImage == nil || s.CurrentImage.Image != imageInfo.Image {
			s.CurrentImage = &imageInfo
			return true
		}
		return false
	}, true)
}

// InvalidateSyncStatus resets the sync state to false and triggers an inspection
func (ac *actionContext) InvalidateSyncStatus() {
	ac.context.InvalidateSyncStatus()
}

// DisableScalingCluster disables scaling DBservers and coordinators
func (ac *actionContext) DisableScalingCluster(ctx context.Context) error {
	return ac.context.DisableScalingCluster(ctx)
}

// EnableScalingCluster enables scaling DBservers and coordinators
func (ac *actionContext) EnableScalingCluster(ctx context.Context) error {
	return ac.context.EnableScalingCluster(ctx)
}
