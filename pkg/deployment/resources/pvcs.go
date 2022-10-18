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

package resources

import (
	"context"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	inspectorInterface "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector"
)

// createPVCFinalizers creates a list of finalizers for a PVC created for the given group.
func (r *Resources) createPVCFinalizers() []string {
	return []string{constants.FinalizerPVCMemberExists}
}

func (r *Resources) ensurePVCFinalizers(in meta.Object) bool {
	return k8sutil.EnsureFinalizers(in, r.createPVCFinalizers(), nil)
}

// EnsurePVCs creates all PVC's listed in member status
func (r *Resources) EnsurePVCs(ctx context.Context, cachedStatus inspectorInterface.Inspector) error {
	apiObject := r.context.GetAPIObject()
	deploymentName := apiObject.GetName()
	owner := apiObject.AsOwner()
	iterator := r.context.GetServerGroupIterator()
	status := r.context.GetStatus()
	enforceAntiAffinity := r.context.GetSpec().GetEnvironment().IsProduction()

	if err := iterator.ForeachServerGroupAccepted(func(group api.ServerGroup, spec api.ServerGroupSpec, status *api.MemberStatusList) error {
		for _, m := range *status {
			if m.PersistentVolumeClaim.GetName() == "" {
				continue
			}

			if _, exists := cachedStatus.PersistentVolumeClaim().V1().GetSimple(m.PersistentVolumeClaim.GetName()); exists {
				continue
			}

			storageClassName := spec.GetStorageClassName()
			role := group.AsRole()
			resources := spec.Resources
			vct := spec.VolumeClaimTemplate
			finalizers := r.createPVCFinalizers()
			err := globals.GetGlobalTimeouts().Kubernetes().RunWithTimeout(ctx, func(ctxChild context.Context) error {
				return k8sutil.CreatePersistentVolumeClaim(ctxChild, cachedStatus.PersistentVolumeClaimsModInterface().V1(),
					m.PersistentVolumeClaim.GetName(), deploymentName, storageClassName, role, enforceAntiAffinity,
					resources, vct, finalizers, owner)
			})
			if err != nil {
				return errors.WithStack(err)
			}
		}
		return nil
	}, &status); err != nil {
		return errors.WithStack(err)
	}
	return nil
}
