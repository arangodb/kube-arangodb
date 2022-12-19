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

package reconcile

import (
	"context"
	"time"

	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/rotation"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector"
)

func newRuntimeContainerImageUpdateAction(action api.Action, actionCtx ActionContext) Action {
	a := &actionRuntimeContainerImageUpdate{}

	a.actionImpl = newBaseActionImplDefRef(action, actionCtx)

	return a
}

var _ ActionPost = &actionRuntimeContainerImageUpdate{}
var _ ActionPre = &actionRuntimeContainerImageUpdate{}

type actionRuntimeContainerImageUpdate struct {
	// actionImpl implement timeout and member id functions
	actionImpl
}

func (a actionRuntimeContainerImageUpdate) Pre(ctx context.Context) error {
	a.log.Info("Updating member condition")
	m, ok := a.actionCtx.GetMemberStatusByID(a.action.MemberID)
	if !ok {
		a.log.Info("member is gone already")
		return nil
	}

	cname, _, ok := a.getContainerDetails()
	if !ok {
		a.log.Info("Unable to find container details")
		return nil
	}

	if c, ok := m.Conditions.Get(api.ConditionTypeUpdating); ok {
		if c.Params == nil {
			c.Params = api.ConditionParams{}
		}

		if c.Params[api.ConditionParamContainerUpdatingName] != cname {
			c.Params[api.ConditionParamContainerUpdatingName] = cname

			if err := a.actionCtx.UpdateMember(ctx, m); err != nil {
				return err
			}
		}
	}

	return nil
}

func (a actionRuntimeContainerImageUpdate) Post(ctx context.Context) error {
	a.log.Info("Updating container image")
	m, ok := a.actionCtx.GetMemberStatusByID(a.action.MemberID)
	if !ok {
		a.log.Info("member is gone already")
		return nil
	}

	if c, ok := m.Conditions.Get(api.ConditionTypeUpdating); ok {
		if c.Params != nil {
			if _, ok := c.Params[api.ConditionParamContainerUpdatingName]; ok {
				delete(c.Params, api.ConditionParamContainerUpdatingName)

				if len(c.Params) == 0 {
					c.Params = nil
				}

				if err := a.actionCtx.UpdateMember(ctx, m); err != nil {
					return err
				}
			}
		}
	}

	cname, image, ok := a.getContainerDetails()
	if !ok {
		a.log.Info("Unable to find container details")
		return nil
	}

	cache := a.actionCtx.ACS().CurrentClusterCache()
	name := m.ArangoMemberName(a.actionCtx.GetName(), a.action.Group)

	_, ok = cache.ArangoMember().V1().GetSimple(name)
	if !ok {
		err := errors.Newf("ArangoMember not found")
		a.log.Err(err).Error("ArangoMember not found")
		return err
	}

	return inspector.WithArangoMemberStatusUpdate(ctx, cache, name, func(in *api.ArangoMember) (bool, error) {
		if in.Spec.Template == nil || in.Status.Template == nil ||
			in.Spec.Template.PodSpec == nil || in.Status.Template.PodSpec == nil {
			a.log.Info("Nil Member definition")
			return false, nil
		}

		if len(in.Spec.Template.PodSpec.Spec.Containers) != len(in.Status.Template.PodSpec.Spec.Containers) {
			a.log.Info("Invalid size of containers")
			return false, nil
		}

		for id := range in.Spec.Template.PodSpec.Spec.Containers {
			if in.Spec.Template.PodSpec.Spec.Containers[id].Name == cname {
				if in.Status.Template.PodSpec.Spec.Containers[id].Name != cname {
					a.log.Info("Invalid order of containers")
					return false, nil
				}

				if in.Spec.Template.PodSpec.Spec.Containers[id].Image != image {
					a.log.Str("got", in.Spec.Template.PodSpec.Spec.Containers[id].Image).Str("expected", image).Info("Invalid spec image of container")
					return false, nil
				}

				if in.Status.Template.PodSpec.Spec.Containers[id].Image != image {
					in.Status.Template.PodSpec.Spec.Containers[id].Image = image
					return true, nil
				}
				return false, nil
			}
		}
		return false, nil
	})
}

func (a actionRuntimeContainerImageUpdate) getContainerDetails() (string, string, bool) {
	container, ok := a.action.GetParam(rotation.ContainerName)
	if !ok {
		return "", "", false
	}

	image, ok := a.action.GetParam(rotation.ContainerImage)
	if !ok {
		return "", "", false
	}

	return container, image, true
}

// Start starts the action for changing conditions on the provided member.
func (a actionRuntimeContainerImageUpdate) Start(ctx context.Context) (bool, error) {
	m, ok := a.actionCtx.GetMemberStatusByID(a.action.MemberID)
	if !ok {
		a.log.Info("member is gone already")
		return true, nil
	}

	cache, ok := a.actionCtx.ACS().ClusterCache(m.ClusterID)
	if !ok {
		return true, errors.Newf("Client is not ready")
	}

	name, image, ok := a.getContainerDetails()
	if !ok {
		a.log.Info("Unable to find container details")
		return true, nil
	}

	if !m.Phase.IsReady() {
		a.log.Info("Member is not ready, unable to run update operation")
		return true, nil
	}

	member, ok := a.actionCtx.ACS().CurrentClusterCache().ArangoMember().V1().GetSimple(m.ArangoMemberName(a.actionCtx.GetName(), a.action.Group))
	if !ok {
		err := errors.Newf("ArangoMember not found")
		a.log.Err(err).Error("ArangoMember not found")
		return false, err
	}

	pod, ok := cache.Pod().V1().GetSimple(m.Pod.GetName())
	if !ok {
		a.log.Info("pod is not present")
		return true, nil
	}

	if member.Spec.Template == nil || member.Spec.Template.PodSpec == nil {
		a.log.Info("pod spec is not present")
		return true, nil
	}

	if member.Status.Template == nil || member.Status.Template.PodSpec == nil {
		a.log.Info("pod status is not present")
		return true, nil
	}

	if len(pod.Spec.Containers) != len(member.Spec.Template.PodSpec.Spec.Containers) {
		a.log.Info("spec container count is not equal")
		return true, nil
	}

	if len(pod.Spec.Containers) != len(member.Status.Template.PodSpec.Spec.Containers) {
		a.log.Info("status container count is not equal")
		return true, nil
	}

	spec := member.Spec.Template.PodSpec
	status := member.Status.Template.PodSpec

	for id := range pod.Spec.Containers {
		if pod.Spec.Containers[id].Name != spec.Spec.Containers[id].Name ||
			pod.Spec.Containers[id].Name != status.Spec.Containers[id].Name ||
			pod.Spec.Containers[id].Name != name {
			continue
		}

		if pod.Spec.Containers[id].Image != image {
			// Update pod image
			pod.Spec.Containers[id].Image = image

			if _, err := a.actionCtx.ACS().CurrentClusterCache().PodsModInterface().V1().Update(ctx, pod, meta.UpdateOptions{}); err != nil {
				return true, err
			}

			// Start wait action
			return false, nil
		}

		return true, nil
	}

	return true, nil
}

func (a actionRuntimeContainerImageUpdate) CheckProgress(ctx context.Context) (bool, bool, error) {
	a.log.Info("Update Progress")
	m, ok := a.actionCtx.GetMemberStatusByID(a.action.MemberID)
	if !ok {
		a.log.Info("member is gone already")
		return true, false, nil
	}

	groupSpec := a.actionCtx.GetSpec().GetServerGroupSpec(a.action.Group)

	cache, ok := a.actionCtx.ACS().ClusterCache(m.ClusterID)
	if !ok {
		a.log.Info("Cluster is not ready")
		return false, false, nil
	}

	pod, ok := cache.Pod().V1().GetSimple(m.Pod.GetName())
	if !ok {
		a.log.Info("pod is not present")
		return true, false, nil
	}

	if err := k8sutil.EnsureFinalizerAbsent(ctx, cache.PodsModInterface().V1(), pod, k8sutil.GetFinalizers(groupSpec, a.action.Group)...); err != nil {
		a.log.Err(err).Error("Unable to enforce finalizer")
	}

	name, image, ok := a.getContainerDetails()
	if !ok {
		a.log.Info("Unable to find container details")
		return true, false, nil
	}

	cspec, ok := k8sutil.GetContainerByName(pod, name)
	if !ok {
		a.log.Info("Unable to find container spec")
		return true, false, nil
	}

	cstatus, ok := k8sutil.GetContainerStatusByName(pod, name)
	if !ok {
		a.log.Info("Unable to find container status")
		return true, false, nil
	}

	if cspec.Image != image {
		a.log.Info("Image changed")
		return true, false, nil
	}

	if s := cstatus.State.Terminated; s != nil {
		// We are in terminated state
		// Image is changed after start
		if cspec.Image != cstatus.Image {
			// Image not yet updated, retry soon
			return false, false, nil
		}

		// Pod won't get up and running
		return true, false, errors.Newf("Container %s failed during image replacement: (%d) %s: %s", name, s.ExitCode, s.Reason, s.Message)
	} else if s := cstatus.State.Waiting; s != nil {
		if pod.Spec.RestartPolicy == core.RestartPolicyAlways {
			lastTermination := cstatus.LastTerminationState.Terminated
			if lastTermination != nil {
				allowedRestartPeriod := time.Now().Add(time.Second * -20)
				if lastTermination.FinishedAt.Time.Before(allowedRestartPeriod) {
					return true, false, errors.Newf("Container %s continuously failing during image replacement: (%d) %s: %s", name, lastTermination.ExitCode, lastTermination.Reason, lastTermination.Message)
				} else {
					a.log.Str("pod-name", pod.GetName()).Debug("pod is restarting - we are not marking it as terminated yet..")
				}
			}
		}

		// Pod is still pulling image or pending for pod start
		return false, false, nil
	} else if s := cstatus.State.Running; s != nil {
		// Image is changed after restart
		if cspec.Image != cstatus.Image {
			// Image not yet updated, retry soon
			return false, false, nil
		}

		return true, false, nil
	} else {
		// Unknown state
		return false, false, nil
	}
}
