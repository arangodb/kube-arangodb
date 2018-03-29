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

package resources

import (
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"

	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

// EnsurePVCs creates all PVC's listed in member status
func (r *Resources) EnsurePVCs() error {
	kubecli := r.context.GetKubeCli()
	apiObject := r.context.GetAPIObject()
	deploymentName := apiObject.GetName()
	ns := apiObject.GetNamespace()
	owner := apiObject.AsOwner()
	iterator := r.context.GetServerGroupIterator()
	status := r.context.GetStatus()

	if err := iterator.ForeachServerGroup(func(group api.ServerGroup, spec api.ServerGroupSpec, status *api.MemberStatusList) error {
		for _, m := range *status {
			if m.PersistentVolumeClaimName != "" {
				storageClassName := spec.GetStorageClassName()
				role := group.AsRole()
				resources := spec.Resources
				if err := k8sutil.CreatePersistentVolumeClaim(kubecli, m.PersistentVolumeClaimName, deploymentName, ns, storageClassName, role, resources, owner); err != nil {
					return maskAny(err)
				}
			}
		}
		return nil
	}, &status); err != nil {
		return maskAny(err)
	}
	return nil
}
