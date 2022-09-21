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

package v1

import (
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arangodb/kube-arangodb/pkg/apis/deployment"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/tools"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ArangoDeploymentList is a list of ArangoDB clusters.
type ArangoDeploymentList struct {
	meta.TypeMeta `json:",inline"`
	// Standard list metadata
	// More info: http://releases.k8s.io/HEAD/docs/devel/api-conventions.md#metadata
	meta.ListMeta `json:"metadata,omitempty"`
	Items         []ArangoDeployment `json:"items"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ArangoDeployment contains the entire Kubernetes info for an ArangoDB database deployment.
type ArangoDeployment struct {
	meta.TypeMeta   `json:",inline"`
	meta.ObjectMeta `json:"metadata,omitempty"`
	Spec            DeploymentSpec   `json:"spec,omitempty"`
	Status          DeploymentStatus `json:"status,omitempty"`
}

func (d *ArangoDeployment) OwnerOf(in meta.Object) bool {
	return tools.IsOwner(d.AsOwner(), in)
}

type ServerGroupFunc func(ServerGroup, ServerGroupSpec, *MemberStatusList) error

// AsOwner creates an OwnerReference for the given deployment
func (d *ArangoDeployment) AsOwner() meta.OwnerReference {
	trueVar := true
	return meta.OwnerReference{
		APIVersion: SchemeGroupVersion.String(),
		Kind:       deployment.ArangoDeploymentResourceKind,
		Name:       d.Name,
		UID:        d.UID,
		Controller: &trueVar,
		// For now BlockOwnerDeletion does not work on OpenShift, so we leave it out.
		//BlockOwnerDeletion: &trueVar,
	}
}

// ForeachServerGroup calls the given callback for all server groups.
// If the callback returns an error, this error is returned and no other server
// groups are processed.
// Groups are processed in this order: agents, single, dbservers, coordinators, syncmasters, syncworkers
func (d *ArangoDeployment) ForeachServerGroup(cb ServerGroupFunc, status *DeploymentStatus) error {
	if status == nil {
		status = &d.Status
	}
	return d.foreachServerGroup(cb, d.Spec, status)
}

// ForeachServerGroupAccepted calls the given callback for all accepted server groups.
// If the callback returns an error, this error is returned and no other server
// groups are processed.
// Groups are processed in this order: agents, single, dbservers, coordinators, syncmasters, syncworkers
func (d *ArangoDeployment) ForeachServerGroupAccepted(cb ServerGroupFunc, status *DeploymentStatus) error {
	if status == nil {
		status = &d.Status
	}
	spec := d.Spec
	if a := status.AcceptedSpec; a != nil {
		spec = *a
	}
	return d.foreachServerGroup(cb, spec, status)
}

func (d *ArangoDeployment) foreachServerGroup(cb ServerGroupFunc, spec DeploymentSpec, status *DeploymentStatus) error {
	if err := cb(ServerGroupAgents, spec.Agents, &status.Members.Agents); err != nil {
		return errors.WithStack(err)
	}
	if err := cb(ServerGroupSingle, spec.Single, &status.Members.Single); err != nil {
		return errors.WithStack(err)
	}
	if err := cb(ServerGroupDBServers, spec.DBServers, &status.Members.DBServers); err != nil {
		return errors.WithStack(err)
	}
	if err := cb(ServerGroupCoordinators, spec.Coordinators, &status.Members.Coordinators); err != nil {
		return errors.WithStack(err)
	}
	if err := cb(ServerGroupSyncMasters, spec.SyncMasters, &status.Members.SyncMasters); err != nil {
		return errors.WithStack(err)
	}
	if err := cb(ServerGroupSyncWorkers, spec.SyncWorkers, &status.Members.SyncWorkers); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// IsAccepted checks if accepted version match current version in spec
func (d ArangoDeployment) IsAccepted() (bool, error) {
	if as := d.Status.AcceptedSpecVersion; as != nil {
		sha, err := d.Spec.Checksum()
		if err != nil {
			return false, err
		}

		return *as == sha, nil
	}

	return false, nil
}

func (d ArangoDeployment) GetAcceptedSpec() DeploymentSpec {
	if a := d.Status.AcceptedSpec; a != nil {
		return *a
	} else {
		return d.Spec
	}
}

// IsUpToDate checks if applied version match current version in spec
func (d ArangoDeployment) IsUpToDate() (bool, error) {
	sha, err := d.Spec.Checksum()
	if err != nil {
		return false, err
	}

	return sha == d.Status.AppliedVersion && d.Status.Conditions.IsTrue(ConditionTypeUpToDate), nil
}
