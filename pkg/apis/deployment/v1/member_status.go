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

package v1

import (
	"time"

	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	driver "github.com/arangodb/go-driver"

	"github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/strings"
)

// MemberStatus holds the current status of a single member (server)
type MemberStatus struct {
	// ID holds the unique ID of the member.
	// This id is also used within the ArangoDB cluster to identify this server.
	ID string `json:"id"`
	// UID holds the unique UID of the member.
	// UID is created once member run in AddMember action.
	UID types.UID `json:"uid,omitempty"`
	// RID holds the ID of the member run.
	// Value is updated in Pending Phase.
	RID types.UID `json:"rid,omitempty"`
	// ClusterID holds the ID of the Arango deployment.
	ClusterID types.UID `json:"cid,omitempty"`
	// Phase holds the current lifetime phase of this member
	Phase MemberPhase `json:"phase"`
	// CreatedAt holds the creation timestamp of this member.
	CreatedAt meta.Time `json:"created-at"`
	// Conditions specific to this member
	Conditions ConditionList `json:"conditions,omitempty"`
	// RecentTerminations holds the times when this member was recently terminated.
	// First entry is the oldest. (do not add omitempty, since we want to be able to switch from a list to an empty list)
	RecentTerminations []meta.Time `json:"recent-terminations"`
	// IsInitialized is set after the very first time a pod was created for this member.
	// After that, DBServers must have a UUID field or fail.
	IsInitialized bool `json:"initialized"`
	// CleanoutJobID holds the ID of the agency job for cleaning out this server
	CleanoutJobID string `json:"cleanout-job-id,omitempty"`
	// ArangoVersion holds the ArangoDB version in member
	ArangoVersion driver.Version `json:"arango-version,omitempty"`
	// ImageId holds the members ArangoDB image ID
	ImageID string `json:"image-id,omitempty"`
	// Image holds image details
	Image *ImageInfo `json:"image,omitempty"`
	// OldImage holds old image defails
	OldImage *ImageInfo `json:"old-image,omitempty"`
	// Architecture defines Image architecture type
	Architecture *ArangoDeploymentArchitectureType `json:"architecture,omitempty"`
	// Upgrade define if upgrade should be enforced during next execution
	Upgrade bool `json:"upgrade,omitempty"`

	// Topology define topology member status assignment
	Topology     *TopologyMemberStatus `json:"topology,omitempty"`
	Pod          *MemberPodStatus      `json:"pod,omitempty"`
	SecondaryPod *MemberPodStatus      `json:"secondaryPod,omitempty"`

	// PersistentVolumeClaim keeps information about PVC for Primary Pod
	PersistentVolumeClaim *MemberPersistentVolumeClaimStatus `json:"persistentVolumeClaim,omitempty"`

	// SecondaryPersistentVolumeClaim keeps information about PVC for SecondaryPod
	SecondaryPersistentVolumeClaim *MemberPersistentVolumeClaimStatus `json:"secondaryPersistentVolumeClaim,omitempty"`

	// deprecated
	// SideCarSpecs contains list of specifications specified for side cars
	SideCarSpecs map[string]core.Container `json:"sidecars-specs,omitempty"`
	// deprecated
	// PodName holds the name of the Pod that currently runs this member
	PodName string `json:"podName,omitempty"`
	// deprecated
	// PodUID holds the UID of the Pod that currently runs this member
	PodUID types.UID `json:"podUID,omitempty"`
	// deprecated
	// PodSpecVersion holds the checksum of Pod spec that currently runs this member. Used to rotate pods
	PodSpecVersion string `json:"podSpecVersion,omitempty"`
	// deprecated
	// PersistentVolumeClaimName holds the name of the persistent volume claim used for this member (if any).
	PersistentVolumeClaimName string `json:"persistentVolumeClaimName,omitempty"`
	// deprecated
	// Endpoint definition how member should be reachable
	Endpoint *string `json:"-"`
}

// Equal checks for equality
func (s MemberStatus) Equal(other MemberStatus) bool {
	return s.ID == other.ID &&
		s.UID == other.UID &&
		s.RID == other.RID &&
		s.ClusterID == other.ClusterID &&
		s.Phase == other.Phase &&
		util.TimeCompareEqual(s.CreatedAt, other.CreatedAt) &&
		s.PersistentVolumeClaim.Equal(other.PersistentVolumeClaim) &&
		s.SecondaryPersistentVolumeClaim.Equal(other.SecondaryPersistentVolumeClaim) &&
		s.Pod.Equal(other.Pod) &&
		s.SecondaryPod.Equal(other.SecondaryPod) &&
		s.Conditions.Equal(other.Conditions) &&
		s.IsInitialized == other.IsInitialized &&
		s.CleanoutJobID == other.CleanoutJobID &&
		s.ArangoVersion == other.ArangoVersion &&
		s.ImageID == other.ImageID &&
		s.Image.Equal(other.Image) &&
		s.OldImage.Equal(other.OldImage) &&
		s.Architecture.Equal(other.Architecture) &&
		s.Upgrade == other.Upgrade &&
		strings.CompareStringPointers(s.Endpoint, other.Endpoint)
}

// Age returns the duration since the creation timestamp of this member.
func (s MemberStatus) Age() time.Duration {
	return time.Since(s.CreatedAt.Time)
}

// RemoveTerminationsBefore removes all recent terminations before the given timestamp.
// It returns the number of terminations that have been removed.
func (s *MemberStatus) RemoveTerminationsBefore(timestamp time.Time) int {
	removed := 0
	for {
		if len(s.RecentTerminations) == 0 {
			// Nothing left
			return removed
		}
		if s.RecentTerminations[0].Time.Before(timestamp) {
			// Let's remove it
			s.RecentTerminations = s.RecentTerminations[1:]
			removed++
		} else {
			// First (oldest) is not before given timestamp, we're done
			return removed
		}
	}
}

// deprecated
func (s *MemberStatus) GetEndpoint(defaultEndpoint string) string {
	if s == nil || s.Endpoint == nil {
		return defaultEndpoint
	}

	return *s.Endpoint
}

// RecentTerminationsSince returns the number of terminations since the given timestamp.
func (s MemberStatus) RecentTerminationsSince(timestamp time.Time) int {
	count := 0
	for idx := len(s.RecentTerminations) - 1; idx >= 0; idx-- {
		if s.RecentTerminations[idx].Time.Before(timestamp) {
			// This termination is before the timestamp, so we're done
			return count
		}
		count++
	}
	return count
}

// IsNotReadySince returns true when the given member has not been ready since the given timestamp.
// That means it:
// - A) Was created before timestamp and never reached a ready state or
// - B) The Ready condition is set to false, and last transision is before timestamp
func (s MemberStatus) IsNotReadySince(timestamp time.Time) bool {
	cond, found := s.Conditions.Get(ConditionTypeReady)
	if found {
		// B
		return cond.Status != core.ConditionTrue && cond.LastTransitionTime.Time.Before(timestamp)
	}
	// A
	return s.CreatedAt.Time.Before(timestamp)
}

// ArangoMemberName create member name from given member
func (s MemberStatus) ArangoMemberName(deploymentName string, group ServerGroup) string {
	return shared.CreatePodHostName(deploymentName, group.AsRole(), s.ID)
}
