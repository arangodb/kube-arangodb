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

	driver "github.com/arangodb/go-driver"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/util/arangod"
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
	GetAgencyClients(ctx context.Context) ([]arangod.Agency, error)
	// GetMemberStatusByID returns the current member status
	// for the member with given id.
	// Returns member status, true when found, or false
	// when no such member is found.
	GetMemberStatusByID(id string) (api.MemberStatus, bool)
	// CreateMember adds a new member to the given group.
	CreateMember(group api.ServerGroup) error
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
}

// NewActionContext creates a new ActionContext implementation.
func NewActionContext(log zerolog.Logger, context ReconcileContext) ActionContext {
	return &actionContext{
		log:     log,
		context: context,
	}
}

// actionContext implements ActionContext
type actionContext struct {
	log     zerolog.Logger
	context ReconcileContext
}

// Gets the specified mode of deployment
func (ac *actionContext) GetMode() api.DeploymentMode {
	return ac.context.GetSpec().GetMode()
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
func (ac *actionContext) GetAgencyClients(ctx context.Context) ([]arangod.Agency, error) {
	c, err := ac.context.GetAgencyClients(ctx)
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
	m, _, ok := ac.context.GetStatus().Members.ElementByID(id)
	return m, ok
}

// CreateMember adds a new member to the given group.
func (ac *actionContext) CreateMember(group api.ServerGroup) error {
	if err := ac.context.CreateMember(group); err != nil {
		return maskAny(err)
	}
	return nil
}

// UpdateMember updates the deployment status wrt the given member.
func (ac *actionContext) UpdateMember(member api.MemberStatus) error {
	status := ac.context.GetStatus()
	_, group, found := status.Members.ElementByID(member.ID)
	if !found {
		return maskAny(fmt.Errorf("Member %s not found", member.ID))
	}
	status.Members.UpdateMemberStatus(member, group)
	if err := ac.context.UpdateStatus(status); err != nil {
		log.Debug().Err(err).Msg("Updating CR status failed")
		return maskAny(err)
	}
	return nil
}

// RemoveMemberByID removes a member with given id.
func (ac *actionContext) RemoveMemberByID(id string) error {
	status := ac.context.GetStatus()
	_, group, found := status.Members.ElementByID(id)
	if !found {
		return nil
	}
	if err := status.Members.RemoveByID(id, group); err != nil {
		log.Debug().Err(err).Str("group", group.AsRole()).Msg("Failed to remove member")
		return maskAny(err)
	}
	// Save removed member
	if err := ac.context.UpdateStatus(status); err != nil {
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
