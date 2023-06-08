//
// DISCLAIMER
//
// Copyright 2016-2023 ArangoDB GmbH, Cologne, Germany
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
	"time"

	core "k8s.io/api/core/v1"

	"github.com/arangodb/go-driver"

	backupApi "github.com/arangodb/kube-arangodb/pkg/apis/backup/v1"
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/acs/sutil"
	agencyCache "github.com/arangodb/kube-arangodb/pkg/deployment/agency"
	"github.com/arangodb/kube-arangodb/pkg/deployment/agency/state"
	"github.com/arangodb/kube-arangodb/pkg/deployment/member"
	"github.com/arangodb/kube-arangodb/pkg/deployment/reconciler"
	"github.com/arangodb/kube-arangodb/pkg/logging"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	inspectorInterface "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector"
)

// ActionContext provides methods to the Action implementations
// to control their context.
type ActionContext interface {
	reconciler.DeploymentStatusUpdate
	reconciler.DeploymentAgencyMaintenance
	reconciler.DeploymentPodRenderer
	reconciler.ArangoAgencyGet
	reconciler.DeploymentInfoGetter
	reconciler.DeploymentDatabaseClient
	reconciler.KubernetesEventGenerator

	member.StateInspectorGetter

	sutil.ACSGetter

	Metrics() *Metrics

	ActionLocalsContext
	ActionProgressor

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
	CreateMember(ctx context.Context, group api.ServerGroup, id string, mods ...CreateMemberMod) (string, error)
	// UpdateMember updates the deployment status wrt the given member.
	UpdateMember(ctx context.Context, member api.MemberStatus) error
	// RemoveMemberByID removes a member with given id.
	RemoveMemberByID(ctx context.Context, id string) error
	// GetImageInfo returns the image info for an image with given name.
	// Returns: (info, infoFound)
	GetImageInfo(imageName string) (api.ImageInfo, bool)
	// GetCurrentImageInfo returns the image info for an current image.
	// Returns: (info, infoFound)
	GetCurrentImageInfo() (api.ImageInfo, bool)
	// SetCurrentImage changes the CurrentImage field in the deployment
	// status to the given image.
	SetCurrentImage(ctx context.Context, imageInfo api.ImageInfo) error
	// DisableScalingCluster disables scaling DBservers and coordinators
	DisableScalingCluster(ctx context.Context) error
	// EnableScalingCluster enables scaling DBservers and coordinators
	EnableScalingCluster(ctx context.Context) error
	// UpdateClusterCondition update status of ArangoDeployment with defined modifier. If action returns True action is taken
	UpdateClusterCondition(ctx context.Context, conditionType api.ConditionType, status bool, reason, message string) error
	// GetBackup receives information about a backup resource
	GetBackup(ctx context.Context, backup string) (*backupApi.ArangoBackup, error)
	// GetName receives information about a deployment name
	GetName() string
	// SelectImage select currently used image by pod
	SelectImage(spec api.DeploymentSpec, status api.DeploymentStatus) (api.ImageInfo, bool)
}

type ActionLocalsContext interface {
	CurrentLocals() api.PlanLocals

	Get(action api.Action, key api.PlanLocalKey) (string, bool)
	Add(key api.PlanLocalKey, value string, override bool) bool

	SetTime(key api.PlanLocalKey, t time.Time) bool
	GetTime(action api.Action, key api.PlanLocalKey) (time.Time, bool)

	BackoffExecution(action api.Action, key api.PlanLocalKey, duration time.Duration) bool
}

// ActionProgressor describe functions to follow a progress of an action.
type ActionProgressor interface {
	// GetProgress returns progress of an action.
	GetProgress() string
	// SetProgress sets progress of an action.
	SetProgress(progress string)
}

// newActionContext creates a new ActionContext implementation.
func newActionContext(log logging.Logger, context Context, metrics *Metrics) ActionContext {
	return &actionContext{
		log:     log,
		context: context,
		metrics: metrics,
	}
}

// actionContext implements ActionContext
type actionContext struct {
	context      Context
	log          logging.Logger
	cachedStatus inspectorInterface.Inspector
	locals       api.PlanLocals
	Progress     string
	metrics      *Metrics
}

func (ac *actionContext) IsSyncEnabled() bool {
	return ac.context.IsSyncEnabled()
}

func (ac *actionContext) WithMemberStatusUpdateErr(ctx context.Context, id string, group api.ServerGroup, action reconciler.DeploymentMemberStatusUpdateErrFunc) error {
	return ac.context.WithMemberStatusUpdateErr(ctx, id, group, action)
}

func (ac *actionContext) WithMemberStatusUpdate(ctx context.Context, id string, group api.ServerGroup, action reconciler.DeploymentMemberStatusUpdateFunc) error {
	return ac.context.WithMemberStatusUpdate(ctx, id, group, action)
}

func (ac *actionContext) CreateOperatorEngineOpsAlertEvent(message string, args ...interface{}) {
	ac.context.CreateOperatorEngineOpsAlertEvent(message, args...)
}

func (ac *actionContext) Metrics() *Metrics {
	return ac.metrics
}

func (ac *actionContext) ACS() sutil.ACS {
	return ac.context.ACS()
}

func (ac *actionContext) GetDatabaseAsyncClient(ctx context.Context) (driver.Client, error) {
	return ac.context.GetDatabaseAsyncClient(ctx)
}

func (ac *actionContext) GetServerAsyncClient(id string) (driver.Client, error) {
	return ac.context.GetServerAsyncClient(id)
}

func (ac *actionContext) CurrentLocals() api.PlanLocals {
	return ac.locals
}

func (ac *actionContext) Get(action api.Action, key api.PlanLocalKey) (string, bool) {
	return ac.locals.GetWithParent(action.Locals, key)
}

func (ac *actionContext) BackoffExecution(action api.Action, key api.PlanLocalKey, duration time.Duration) bool {
	t, ok := ac.GetTime(action, key)
	if !ok {
		// Reset as zero time
		t = time.Time{}
	}

	if t.IsZero() || time.Since(t) > duration {
		// Execution is needed
		ac.SetTime(key, time.Now())
		return true
	}

	return false
}

func (ac *actionContext) SetTime(key api.PlanLocalKey, t time.Time) bool {
	return ac.Add(key, t.Format(util.TimeLayout), true)
}

// SetProgress sets progress to an action.
func (ac *actionContext) SetProgress(progress string) {
	ac.Progress = progress
}

// GetProgress gets progress of an action.
func (ac *actionContext) GetProgress() string {
	return ac.Progress
}

func (ac *actionContext) GetTime(action api.Action, key api.PlanLocalKey) (time.Time, bool) {
	s, ok := ac.locals.GetWithParent(action.Locals, key)
	if !ok {
		return time.Time{}, false
	}

	if t, err := time.Parse(util.TimeLayout, s); err != nil {
		return time.Time{}, false
	} else {
		return t, true
	}
}

func (ac *actionContext) Add(key api.PlanLocalKey, value string, override bool) bool {
	return ac.locals.Add(key, value, override)
}

func (ac *actionContext) GetMembersState() member.StateInspector {
	return ac.context.GetMembersState()
}

func (ac *actionContext) UpdateStatus(ctx context.Context, status api.DeploymentStatus) error {
	return ac.context.UpdateStatus(ctx, status)
}

func (ac *actionContext) GetNamespace() string {
	return ac.context.GetNamespace()
}

func (ac *actionContext) GetStatus() api.DeploymentStatus {
	return ac.context.GetStatus()
}

func (ac *actionContext) GetStatusSnapshot() api.DeploymentStatus {
	return ac.context.GetStatus()
}

func (ac *actionContext) GenerateMemberEndpoint(group api.ServerGroup, member api.MemberStatus) (string, error) {
	return ac.context.GenerateMemberEndpoint(group, member)
}

func (ac *actionContext) GetAgencyHealth() (agencyCache.Health, bool) {
	return ac.context.GetAgencyHealth()
}

func (ac *actionContext) ShardsInSyncMap() (state.ShardsSyncStatus, bool) {
	return ac.context.ShardsInSyncMap()
}

func (ac *actionContext) GetAgencyCache() (state.State, bool) {
	return ac.context.GetAgencyCache()
}

func (ac *actionContext) GetAgencyArangoDBCache() (state.DB, bool) {
	return ac.context.GetAgencyArangoDBCache()
}

func (ac *actionContext) SetAgencyMaintenanceMode(ctx context.Context, enabled bool) error {
	return ac.context.SetAgencyMaintenanceMode(ctx, enabled)
}

func (ac *actionContext) RenderPodForMember(ctx context.Context, acs sutil.ACS, spec api.DeploymentSpec, status api.DeploymentStatus, memberID string, imageInfo api.ImageInfo) (*core.Pod, error) {
	return ac.context.RenderPodForMember(ctx, acs, spec, status, memberID, imageInfo)
}

func (ac *actionContext) RenderPodTemplateForMember(ctx context.Context, acs sutil.ACS, spec api.DeploymentSpec, status api.DeploymentStatus, memberID string, imageInfo api.ImageInfo) (*core.PodTemplateSpec, error) {
	return ac.context.RenderPodTemplateForMember(ctx, acs, spec, status, memberID, imageInfo)
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

func (ac *actionContext) GetBackup(ctx context.Context, backup string) (*backupApi.ArangoBackup, error) {
	return ac.context.GetBackup(ctx, backup)
}

func (ac *actionContext) WithStatusUpdateErr(ctx context.Context, action reconciler.DeploymentStatusUpdateErrFunc) error {
	return ac.context.WithStatusUpdateErr(ctx, action)
}

func (ac *actionContext) WithStatusUpdate(ctx context.Context, action reconciler.DeploymentStatusUpdateFunc) error {
	return ac.context.WithStatusUpdate(ctx, action)
}

func (ac *actionContext) UpdateClusterCondition(ctx context.Context, conditionType api.ConditionType, status bool, reason, message string) error {
	return ac.context.WithStatusUpdate(ctx, func(s *api.DeploymentStatus) bool {
		return s.Conditions.Update(conditionType, status, reason, message)
	})
}

func (ac *actionContext) GetAPIObject() k8sutil.APIObject {
	return ac.context.GetAPIObject()
}

func (ac *actionContext) CreateEvent(evt *k8sutil.Event) {
	ac.context.CreateEvent(evt)
}

// GetMode gets the specified mode of deployment.
func (ac *actionContext) GetMode() api.DeploymentMode {
	return ac.context.GetSpec().GetMode()
}

func (ac *actionContext) GetSpec() api.DeploymentSpec {
	return ac.context.GetSpec()
}

// GetMemberStatusByID returns the current member status
// for the member with given id.
// Returns member status, true when found, or false
// when no such member is found.
func (ac *actionContext) GetMemberStatusByID(id string) (api.MemberStatus, bool) {
	status := ac.context.GetStatus()
	m, _, ok := status.Members.ElementByID(id)
	return m, ok
}

// GetMemberStatusAndGroupByID returns the current member status and group
// for the member with given id.
// Returns member status, true when found, or false
// when no such member is found.
func (ac *actionContext) GetMemberStatusAndGroupByID(id string) (api.MemberStatus, api.ServerGroup, bool) {
	status := ac.context.GetStatus()
	return status.Members.ElementByID(id)
}

// CreateMember adds a new member to the given group.
// If ID is non-empty, it will be used, otherwise a new ID is created.
func (ac *actionContext) CreateMember(ctx context.Context, group api.ServerGroup, id string, mods ...CreateMemberMod) (string, error) {
	result, err := ac.context.CreateMember(ctx, group, id, mods...)
	if err != nil {
		return "", errors.WithStack(err)
	}
	return result, nil
}

// UpdateMember updates the deployment status wrt the given member.
func (ac *actionContext) UpdateMember(ctx context.Context, member api.MemberStatus) error {
	status := ac.context.GetStatus()
	_, group, found := status.Members.ElementByID(member.ID)
	if !found {
		return errors.WithStack(errors.Newf("Member %s not found", member.ID))
	}
	if err := status.Members.Update(member, group); err != nil {
		return errors.WithStack(err)
	}
	if err := ac.context.UpdateStatus(ctx, status); err != nil {
		ac.log.Err(err).Debug("Updating CR status failed")
		return errors.WithStack(err)
	}
	return nil
}

// RemoveMemberByID removes a member with given id.
func (ac *actionContext) RemoveMemberByID(ctx context.Context, id string) error {
	status := ac.context.GetStatus()
	_, group, found := status.Members.ElementByID(id)
	if !found {
		return nil
	}
	if err := status.Members.RemoveByID(id, group); err != nil {
		ac.log.Err(err).Str("group", group.AsRole()).Debug("Failed to remove member")
		return errors.WithStack(err)
	}
	// Save removed member
	if err := ac.context.UpdateStatus(ctx, status); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// GetImageInfo returns the image info for an image with given name.
// Returns: (info, infoFound)
func (ac *actionContext) GetImageInfo(imageName string) (api.ImageInfo, bool) {
	status := ac.context.GetStatus()
	return status.Images.GetByImage(imageName)
}

// GetImageInfo returns the image info for an current image.
// Returns: (info, infoFound)
func (ac *actionContext) GetCurrentImageInfo() (api.ImageInfo, bool) {
	status := ac.context.GetStatus()

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
	})
}

// DisableScalingCluster disables scaling DBservers and coordinators
func (ac *actionContext) DisableScalingCluster(ctx context.Context) error {
	return ac.context.DisableScalingCluster(ctx)
}

// EnableScalingCluster enables scaling DBservers and coordinators
func (ac *actionContext) EnableScalingCluster(ctx context.Context) error {
	return ac.context.EnableScalingCluster(ctx)
}
