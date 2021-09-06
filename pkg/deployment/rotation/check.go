//
// DISCLAIMER
//
// Copyright 2021 ArangoDB GmbH, Cologne, Germany
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

package rotation

import (
	"github.com/arangodb/kube-arangodb/pkg/apis/deployment"
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	inspectorInterface "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector"
	"github.com/rs/zerolog"
	core "k8s.io/api/core/v1"
)

type Mode int

const (
	EnforcedRotation Mode = iota
	GracefulRotation
	InPlaceRotation
	SilentRotation
	SkippedRotation
)

func (m Mode) And(b Mode) Mode {
	if m < b {
		return m
	}

	return b
}

// CheckPossible returns true if rotation is possible
func CheckPossible(member api.MemberStatus) bool {
	if !member.Phase.IsReady() {
		// Skip rotation when we are not ready
		return false
	}

	if member.Conditions.IsTrue(api.ConditionTypeTerminated) || member.Conditions.IsTrue(api.ConditionTypeTerminating) {
		// Termination in progress, nothing to do
		return false
	}

	return true
}

func IsRotationRequired(log zerolog.Logger, cachedStatus inspectorInterface.Inspector, spec api.DeploymentSpec, member api.MemberStatus, pod *core.Pod, specTemplate, statusTemplate *api.ArangoMemberPodTemplate) (mode Mode, plan api.Plan, reason string, err error) {
	// Determine if rotation is required based on plan and actions

	// Set default mode for return value
	mode = SkippedRotation

	if !CheckPossible(member) {
		// Check is not possible due to improper state of member
		return
	}

	if spec.MemberPropagationMode.Get() == api.DeploymentMemberPropagationModeAlways && member.Conditions.IsTrue(api.ConditionTypePendingRestart) {
		reason = "Restart is pending"
		mode = EnforcedRotation
		return
	}

	// Check if pod details are propagated
	if pod != nil {
		if member.PodUID != pod.UID {
			reason = "Pod UID does not match, this pod is not managed by Operator. Recreating"
			mode = EnforcedRotation
			return
		}

		if _, ok := pod.Annotations[deployment.ArangoDeploymentPodRotateAnnotation]; ok {
			reason = "Recreation enforced by annotation"
			mode = EnforcedRotation
			return
		}
	}

	if member.PodSpecVersion == "" {
		reason = "Pod Spec Version is nil - recreating pod"
		mode = EnforcedRotation
		return
	}

	if specTemplate == nil || statusTemplate == nil {
		// If spec or status is nil rotation is not needed
		return
	}

	// Check if any of resize events are in place
	if member.Conditions.IsTrue(api.ConditionTypePendingTLSRotation) {
		reason = "TLS Rotation pending"
		mode = EnforcedRotation
		return
	}

	pvc, exists := cachedStatus.PersistentVolumeClaim(member.PersistentVolumeClaimName)
	if exists {
		if k8sutil.IsPersistentVolumeClaimFileSystemResizePending(pvc) {
			reason = "PVC Resize pending"
			mode = EnforcedRotation
			return
		}
	}

	if statusTemplate.RotationNeeded(specTemplate) {
		reason = "Pod needs rotation - templates does not match"
		mode = GracefulRotation
		log.Info().Str("id", member.ID).Str("Before", member.PodSpecVersion).Msgf(reason)
		return
	}

	return
}
