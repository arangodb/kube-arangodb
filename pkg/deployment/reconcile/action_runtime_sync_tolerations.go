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
	"reflect"

	"github.com/pkg/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/patch"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/tolerations"
)

func newRuntimeContainerSyncTolerationsAction(action api.Action, actionCtx ActionContext) Action {
	a := &actionRuntimeContainerSyncTolerations{}

	a.actionImpl = newActionImplDefRef(action, actionCtx)

	return a
}

type actionRuntimeContainerSyncTolerations struct {
	// actionImpl implement timeout and member id functions
	actionImpl

	actionEmptyCheckProgress
}

// Start starts the action for changing conditions on the provided member.
func (a actionRuntimeContainerSyncTolerations) Start(ctx context.Context) (bool, error) {
	m, ok := a.actionCtx.GetMemberStatusByID(a.action.MemberID)
	if !ok {
		a.log.Info("member is gone already")
		return true, nil
	}

	cache, ok := a.actionCtx.ACS().ClusterCache(m.ClusterID)
	if !ok {
		return true, errors.Errorf("Client is not ready")
	}

	memberName := m.ArangoMemberName(a.actionCtx.GetName(), a.action.Group)
	member, ok := cache.ArangoMember().V1().GetSimple(memberName)
	if !ok {
		return false, errors.Errorf("ArangoMember %s not found", memberName)
	}

	pod, ok := cache.Pod().V1().GetSimple(m.Pod.GetName())
	if !ok {
		a.log.Str("podName", m.Pod.GetName()).Info("pod is not present")
		return true, nil
	}

	currentTolerations := pod.Spec.Tolerations

	expectedTolerations := member.Spec.Template.PodSpec.Spec.Tolerations

	origTolerations := tolerations.CreatePodTolerations(a.actionCtx.GetMode(), a.action.Group)

	calculatedTolerations := tolerations.MergeTolerationsIfNotFound(currentTolerations, origTolerations, expectedTolerations)

	if reflect.DeepEqual(currentTolerations, calculatedTolerations) {
		return true, nil
	}

	p, err := patch.NewPatch(patch.ItemReplace(patch.NewPath("spec", "tolerations"), calculatedTolerations)).Marshal()
	if err != nil {
		return false, errors.Wrapf(err, "Unable to create patch")
	}

	nctx, c := globals.GetGlobalTimeouts().Kubernetes().WithTimeout(ctx)
	defer c()

	if _, err := a.actionCtx.ACS().CurrentClusterCache().PodsModInterface().V1().Patch(nctx, pod.GetName(), types.JSONPatchType, p, meta.PatchOptions{}); err != nil {
		return false, errors.Wrapf(err, "Unable to apply patch")
	}

	return true, nil
}
