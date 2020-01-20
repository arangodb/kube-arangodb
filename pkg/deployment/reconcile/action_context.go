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
// Author Ewout Prangsma
//

package reconcile

import (
	"context"
	"fmt"
	v1 "k8s.io/api/core/v1"

	"github.com/arangodb/go-driver/agency"

	"github.com/arangodb/arangosync-client/client"
	driver "github.com/arangodb/go-driver"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
)

// ActionContext provides methods to the Action implementations
// to control their context.
type ActionContext interface {
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
	// GetMemberStatusByID returns the current member status
	// for the member with given id.
	// Returns member status, true when found, or false
	// when no such member is found.
	GetMemberStatusByID(id string) (api.MemberStatus, bool)
	// CreateMember adds a new member to the given group.
	// If ID is non-empty, it will be used, otherwise a new ID is created.
	CreateMember(group api.ServerGroup, id string) (string, error)
	// UpdateMember updates the deployment status wrt the given member.
	UpdateMember(member api.MemberStatus) error
	// RemoveMemberByID removes a member with given id.
	RemoveMemberByID(id string) error
	// DeletePod deletes a pod with given name in the namespace
	// of the deployment. If the pod does not exist, the error is ignored.
	DeletePod(podName string) error
	// DeletePvc deletes a persistent volume claim with given name in the namespace
	// of the deployment. If the pvc does not exist, the error is ignored.
	DeletePvc(pvcName string) error
	// GetPvc returns PVC info about PVC with given name in the namespace
	// of the deployment.
	GetPvc(pvcName string) (*v1.PersistentVolumeClaim, error)
	// RemovePodFinalizers removes all the finalizers from the Pod with given name in the namespace
	// of the deployment. If the pod does not exist, the error is ignored.
	RemovePodFinalizers(podName string) error
	// DeleteTLSKeyfile removes the Secret containing the TLS keyfile for the given member.
	// If the secret does not exist, the error is ignored.
	DeleteTLSKeyfile(group api.ServerGroup, member api.MemberStatus) error
	// DeleteTLSCASecret removes the Secret containing the TLS CA certificate.
	DeleteTLSCASecret() error
	// GetImageInfo returns the image info for an image with given name.
	// Returns: (info, infoFound)
	GetImageInfo(imageName string) (api.ImageInfo, bool)
	// SetCurrentImage changes the CurrentImage field in the deployment
	// status to the given image.
	SetCurrentImage(imageInfo api.ImageInfo) error
	// GetDeploymentHealth returns a copy of the latest known state of cluster health
	GetDeploymentHealth() (driver.ClusterHealth, error)
	// InvalidateSyncStatus resets the sync state to false and triggers an inspection
	InvalidateSyncStatus()
	// GetSpec returns a copy of the spec
	GetSpec() api.DeploymentSpec
	// DisableScalingCluster disables scaling DBservers and coordinators
	DisableScalingCluster() error
	// EnableScalingCluster enables scaling DBservers and coordinators
	EnableScalingCluster() error
}

// newActionContext creates a new ActionContext implementation.
func newActionContext(log zerolog.Logger, context Context) ActionContext {
	return &actionContext{
		log:     log,
		context: context,
	}
}

// actionContext implements ActionContext
type actionContext struct {
	log     zerolog.Logger
	context Context
}

func (ac *actionContext) GetPvc(pvcName string) (*v1.PersistentVolumeClaim, error) {
	return ac.context.GetPvc(pvcName)
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
		return nil, maskAny(err)
	}
	return c, nil
}

// GetServerClient returns a cached client for a specific server.
func (ac *actionContext) GetServerClient(ctx context.Context, group api.ServerGroup, id string) (driver.Client, error) {
	c, err := ac.context.GetServerClient(ctx, group, id)
	if err != nil {
		return nil, maskAny(err)
	}
	return c, nil
}

// GetAgencyClients returns a client connection for every agency member.
func (ac *actionContext) GetAgencyClients(ctx context.Context) ([]driver.Connection, error) {
	c, err := ac.context.GetAgencyClients(ctx, nil)
	if err != nil {
		return nil, maskAny(err)
	}
	return c, nil
}

// GetAgency returns a connection to the entire agency.
func (ac *actionContext) GetAgency(ctx context.Context) (agency.Agency, error) {
	a, err := ac.context.GetAgency(ctx)
	if err != nil {
		return nil, maskAny(err)
	}
	return a, nil
}

// GetSyncServerClient returns a cached client for a specific arangosync server.
func (ac *actionContext) GetSyncServerClient(ctx context.Context, group api.ServerGroup, id string) (client.API, error) {
	c, err := ac.context.GetSyncServerClient(ctx, group, id)
	if err != nil {
		return nil, maskAny(err)
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

// CreateMember adds a new member to the given group.
// If ID is non-empty, it will be used, otherwise a new ID is created.
func (ac *actionContext) CreateMember(group api.ServerGroup, id string) (string, error) {
	result, err := ac.context.CreateMember(group, id)
	if err != nil {
		return "", maskAny(err)
	}
	return result, nil
}

// UpdateMember updates the deployment status wrt the given member.
func (ac *actionContext) UpdateMember(member api.MemberStatus) error {
	status, lastVersion := ac.context.GetStatus()
	_, group, found := status.Members.ElementByID(member.ID)
	if !found {
		return maskAny(fmt.Errorf("Member %s not found", member.ID))
	}
	if err := status.Members.Update(member, group); err != nil {
		return maskAny(err)
	}
	if err := ac.context.UpdateStatus(status, lastVersion); err != nil {
		log.Debug().Err(err).Msg("Updating CR status failed")
		return maskAny(err)
	}
	return nil
}

// RemoveMemberByID removes a member with given id.
func (ac *actionContext) RemoveMemberByID(id string) error {
	status, lastVersion := ac.context.GetStatus()
	_, group, found := status.Members.ElementByID(id)
	if !found {
		return nil
	}
	if err := status.Members.RemoveByID(id, group); err != nil {
		log.Debug().Err(err).Str("group", group.AsRole()).Msg("Failed to remove member")
		return maskAny(err)
	}
	// Save removed member
	if err := ac.context.UpdateStatus(status, lastVersion); err != nil {
		return maskAny(err)
	}
	return nil
}

// DeletePod deletes a pod with given name in the namespace
// of the deployment. If the pod does not exist, the error is ignored.
func (ac *actionContext) DeletePod(podName string) error {
	if err := ac.context.DeletePod(podName); err != nil {
		return maskAny(err)
	}
	return nil
}

// DeletePvc deletes a persistent volume claim with given name in the namespace
// of the deployment. If the pvc does not exist, the error is ignored.
func (ac *actionContext) DeletePvc(pvcName string) error {
	if err := ac.context.DeletePvc(pvcName); err != nil {
		return maskAny(err)
	}
	return nil
}

// RemovePodFinalizers removes all the finalizers from the Pod with given name in the namespace
// of the deployment. If the pod does not exist, the error is ignored.
func (ac *actionContext) RemovePodFinalizers(podName string) error {
	if err := ac.context.RemovePodFinalizers(podName); err != nil {
		return maskAny(err)
	}
	return nil
}

// DeleteTLSKeyfile removes the Secret containing the TLS keyfile for the given member.
// If the secret does not exist, the error is ignored.
func (ac *actionContext) DeleteTLSKeyfile(group api.ServerGroup, member api.MemberStatus) error {
	if err := ac.context.DeleteTLSKeyfile(group, member); err != nil {
		return maskAny(err)
	}
	return nil
}

// DeleteTLSCASecret removes the Secret containing the TLS CA certificate.
func (ac *actionContext) DeleteTLSCASecret() error {
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
		if err := ac.context.UpdateStatus(status, lastVersion); err != nil {
			return maskAny(err)
		}
	}
	// Do delete the secret
	if err := ac.context.DeleteSecret(secretName); err != nil {
		return maskAny(err)
	}
	return nil
}

// GetImageInfo returns the image info for an image with given name.
// Returns: (info, infoFound)
func (ac *actionContext) GetImageInfo(imageName string) (api.ImageInfo, bool) {
	status, _ := ac.context.GetStatus()
	return status.Images.GetByImage(imageName)
}

// SetCurrentImage changes the CurrentImage field in the deployment
// status to the given image.
func (ac *actionContext) SetCurrentImage(imageInfo api.ImageInfo) error {
	status, lastVersion := ac.context.GetStatus()
	status.CurrentImage = &imageInfo
	if err := ac.context.UpdateStatus(status, lastVersion); err != nil {
		return maskAny(err)
	}
	return nil
}

// InvalidateSyncStatus resets the sync state to false and triggers an inspection
func (ac *actionContext) InvalidateSyncStatus() {
	ac.context.InvalidateSyncStatus()
}

// DisableScalingCluster disables scaling DBservers and coordinators
func (ac *actionContext) DisableScalingCluster() error {
	return ac.context.DisableScalingCluster()
}

// EnableScalingCluster enables scaling DBservers and coordinators
func (ac *actionContext) EnableScalingCluster() error {
	return ac.context.EnableScalingCluster()
}
