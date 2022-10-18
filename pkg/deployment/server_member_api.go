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

package deployment

import (
	"context"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/server"
)

type member struct {
	d     *Deployment
	group api.ServerGroup
	id    string
}

func (m member) status() (api.MemberStatus, bool) {
	status := m.d.GetStatus()
	result, _, found := status.Members.ElementByID(m.id)
	return result, found
}

func (m member) ID() string {
	return m.id
}

func (m member) PodName() string {
	if status, found := m.status(); found {
		return status.Pod.GetName()
	}
	return ""
}

func (m member) PVCName() string {
	if status, found := m.status(); found {
		return status.PersistentVolumeClaim.GetName()
	}
	return ""
}

func (m member) PVName() string {
	if status, found := m.status(); found && status.PersistentVolumeClaim != nil {
		pvcs := m.d.deps.Client.Kubernetes().CoreV1().PersistentVolumeClaims(m.d.Namespace())
		if pvc, err := pvcs.Get(context.Background(), status.PersistentVolumeClaim.GetName(), meta.GetOptions{}); err == nil {
			return pvc.Spec.VolumeName
		}
	}
	return ""
}

func (m member) MemberOfCluster() server.MemberOfCluster {
	switch m.group {
	case api.ServerGroupDBServers, api.ServerGroupCoordinators:
		if status, found := m.status(); found {
			if status.Conditions.IsTrue(api.ConditionTypeMemberOfCluster) {
				return server.IsMemberOfCluster
			} else {
				return server.IsNotMemberOfCluster
			}
		}
	}
	return server.NeverMemberOfCluster
}

func (m member) Ready() bool {
	if status, found := m.status(); found {
		return status.Conditions.IsTrue(api.ConditionTypeReady)
	}
	return false
}
