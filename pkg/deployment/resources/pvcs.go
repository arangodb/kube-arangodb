//
// DISCLAIMER
//
// Copyright 2020 ArangoDB GmbH, Cologne, Germany
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
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/resources/inspector"
	"github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

// createPVCFinalizers creates a list of finalizers for a PVC created for the given group.
func (r *Resources) createPVCFinalizers(group api.ServerGroup) []string {
	return []string{constants.FinalizerPVCMemberExists}
}

// EnsurePVCs creates all PVC's listed in member status
func (r *Resources) EnsurePVCs(cachedStatus inspector.Inspector) error {
	kubecli := r.context.GetKubeCli()
	apiObject := r.context.GetAPIObject()
	deploymentName := apiObject.GetName()
	ns := apiObject.GetNamespace()
	owner := apiObject.AsOwner()
	iterator := r.context.GetServerGroupIterator()
	status, _ := r.context.GetStatus()
	enforceAntiAffinity := r.context.GetSpec().GetEnvironment().IsProduction()
	pvcs := kubecli.CoreV1().PersistentVolumeClaims(apiObject.GetNamespace())

	if err := iterator.ForeachServerGroup(func(group api.ServerGroup, spec api.ServerGroupSpec, status *api.MemberStatusList) error {
		for _, m := range *status {
			if m.PersistentVolumeClaimName == "" {
				continue
			}

			_, exists := cachedStatus.PersistentVolumeClaim(m.PersistentVolumeClaimName)
			if exists {
				continue
			}
			storageClassName := spec.GetStorageClassName()
			role := group.AsRole()
			resources := spec.Resources
			vct := spec.VolumeClaimTemplate
			finalizers := r.createPVCFinalizers(group)
			if err := k8sutil.CreatePersistentVolumeClaim(pvcs, m.PersistentVolumeClaimName, deploymentName, ns, storageClassName, role, enforceAntiAffinity, resources, vct, finalizers, owner); err != nil {
				return maskAny(err)
			}
		}
		return nil
	}, &status); err != nil {
		return maskAny(err)
	}
	return nil
}
