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
	"fmt"
	"strings"

	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/rotation"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/definitions"
)

func newRuntimeContainerArgsLogLevelUpdateAction(action api.Action, actionCtx ActionContext) Action {
	a := &actionRuntimeContainerArgsLogLevelUpdate{}

	a.actionImpl = newActionImplDefRef(action, actionCtx)

	return a
}

var _ ActionPost = &actionRuntimeContainerArgsLogLevelUpdate{}

type actionRuntimeContainerArgsLogLevelUpdate struct {
	// actionImpl implement timeout and member id functions
	actionImpl
}

// Post updates arguments for the specific Arango member.
func (a actionRuntimeContainerArgsLogLevelUpdate) Post(ctx context.Context) error {
	m, ok := a.actionCtx.GetMemberStatusByID(a.action.MemberID)
	if !ok {
		a.log.Info("member is gone already")
		return nil
	}

	cache, ok := a.actionCtx.ACS().ClusterCache(m.ClusterID)
	if !ok {
		return errors.Errorf("Client is not ready")
	}

	memberName := m.ArangoMemberName(a.actionCtx.GetName(), a.action.Group)
	_, ok = cache.ArangoMember().V1().GetSimple(memberName)
	if !ok {
		return errors.Errorf("ArangoMember %s not found", memberName)
	}

	containerName, ok := a.action.GetParam(rotation.ContainerName)
	if !ok {
		a.log.Warn("Unable to find action's param %s", rotation.ContainerName)
		return nil
	}

	log := a.log.Str("containerName", containerName)
	updateMemberStatusArgs := func(in *api.ArangoMember) (bool, error) {
		if in.Spec.Template == nil || in.Status.Template == nil ||
			in.Spec.Template.PodSpec == nil || in.Status.Template.PodSpec == nil {
			log.Info("Nil Member definition")
			return false, nil
		}

		if len(in.Spec.Template.PodSpec.Spec.Containers) != len(in.Status.Template.PodSpec.Spec.Containers) {
			log.Info("Invalid size of containers")
			return false, nil
		}

		for id := range in.Spec.Template.PodSpec.Spec.Containers {
			if in.Spec.Template.PodSpec.Spec.Containers[id].Name == containerName {
				if in.Status.Template.PodSpec.Spec.Containers[id].Name != containerName {
					log.Info("Invalid order of containers")
					return false, nil
				}

				in.Status.Template.PodSpec.Spec.Containers[id].Command = in.Spec.Template.PodSpec.Spec.Containers[id].Command
				log.Info("Updating container args")
				return true, nil
			}
		}

		log.Info("can not find the container")

		return false, nil
	}

	err := inspector.WithArangoMemberStatusUpdate(ctx, cache, memberName, updateMemberStatusArgs)
	if err != nil {
		return errors.WithMessage(err, "Error while updating member status")
	}

	return nil
}

func (a *actionRuntimeContainerArgsLogLevelUpdate) ReloadComponents() []definitions.Component {
	return []definitions.Component{
		definitions.Pod,
	}
}

// Start starts the action for changing conditions on the provided member.
func (a actionRuntimeContainerArgsLogLevelUpdate) Start(ctx context.Context) (bool, error) {
	m, ok := a.actionCtx.GetMemberStatusByID(a.action.MemberID)
	if !ok {
		a.log.Info("member is gone already")
		return true, nil
	}

	cache, ok := a.actionCtx.ACS().ClusterCache(m.ClusterID)
	if !ok {
		return true, errors.Errorf("Client is not ready")
	}

	if !m.Phase.IsReady() {
		a.log.Info("Member is not ready, unable to run update operation")
		return true, nil
	}

	containerName, ok := a.action.GetParam(rotation.ContainerName)
	if !ok {
		return true, nil
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

	var op cmpContainer = func(containerSpec core.Container, containerStatus core.Container) error {
		topicsLogLevel := map[string]string{}

		// Set log levels from the provided spec.
		for _, arg := range containerSpec.Command {
			if ok, topic, value := getTopicAndLevel(arg); ok {
				topicsLogLevel[topic] = value
			}
		}

		if err := a.setLogLevel(ctx, topicsLogLevel); err != nil {
			return errors.WithMessage(err, "can not set log level")
		}

		a.log.Interface("topics", topicsLogLevel).Info("send log level to the ArangoDB")
		return nil
	}

	if err := checkContainer(member, pod, containerName, op); err != nil && err != api.NotFoundError {
		return false, errors.WithMessagef(err, "can not check the container %s", containerName)
	}

	return true, nil
}

type cmpContainer func(spec core.Container, status core.Container) error

func checkContainer(member *api.ArangoMember, pod *core.Pod, containerName string, action cmpContainer) error {
	spec, status, err := validateMemberAndPod(member, pod)
	if err != nil {
		return err
	}

	id := getIndexContainer(pod, spec, status, containerName)
	if id < 0 {
		return api.NotFoundError
	}

	return action(spec.Spec.Containers[id], status.Spec.Containers[id])
}

// getIndexContainer returns the index of the container from the list of containers.
func getIndexContainer(pod *core.Pod, spec *core.PodTemplateSpec, status *core.PodTemplateSpec,
	containerName string) int {

	for id := range pod.Spec.Containers {
		if pod.Spec.Containers[id].Name == spec.Spec.Containers[id].Name ||
			pod.Spec.Containers[id].Name == status.Spec.Containers[id].Name ||
			pod.Spec.Containers[id].Name == containerName {

			return id
		}
	}

	return -1
}

func validateMemberAndPod(member *api.ArangoMember, pod *core.Pod) (*core.PodTemplateSpec, *core.PodTemplateSpec, error) {

	if member.Spec.Template == nil || member.Spec.Template.PodSpec == nil {
		return nil, nil, fmt.Errorf("member spec is not present")
	}

	if member.Status.Template == nil || member.Status.Template.PodSpec == nil {
		return nil, nil, fmt.Errorf("member status is not present")
	}

	if len(pod.Spec.Containers) != len(member.Spec.Template.PodSpec.Spec.Containers) {
		return nil, nil, fmt.Errorf("spec container count is not equal")
	}

	if len(pod.Spec.Containers) != len(member.Status.Template.PodSpec.Spec.Containers) {
		return nil, nil, fmt.Errorf("status container count is not equal")
	}

	return member.Spec.Template.PodSpec, member.Status.Template.PodSpec, nil
}

// CheckProgress returns always true because it does not have to wait for any result.
func (a actionRuntimeContainerArgsLogLevelUpdate) CheckProgress(_ context.Context) (bool, bool, error) {
	return true, false, nil
}

// setLogLevel sets the log's levels for the specific server.
func (a actionRuntimeContainerArgsLogLevelUpdate) setLogLevel(ctx context.Context, logLevels map[string]string) error {
	if len(logLevels) == 0 {
		return nil
	}

	cli, err := a.actionCtx.GetMembersState().GetMemberClient(a.action.MemberID)
	if err != nil {
		return err
	}
	conn := cli.Connection()

	req, err := conn.NewRequest("PUT", "_admin/log/level")
	if err != nil {
		return err
	}

	if _, err := req.SetBody(logLevels); err != nil {
		return err
	}

	ctxChild, cancel := globals.GetGlobalTimeouts().ArangoD().WithTimeout(ctx)
	defer cancel()
	resp, err := conn.Do(ctxChild, req)
	if err != nil {
		return err
	}

	return resp.CheckStatus(200)
}

// getTopicAndLevel returns topics and log level from the argument.
func getTopicAndLevel(arg string) (bool, string, string) {
	if !strings.HasPrefix(strings.TrimLeft(arg, " "), "--log.level") {
		return false, "", ""
	}

	logLevelOption := k8sutil.ExtractStringToOptionPair(arg)
	if len(logLevelOption.Value) > 0 {
		logValueOption := k8sutil.ExtractStringToOptionPair(logLevelOption.Value)
		if len(logValueOption.Value) > 0 {
			// It is the topic log, e.g.: --log.level=request=INFO.
			return true, logValueOption.Key, logValueOption.Value
		} else {
			// It is the general log, e.g.: --log.level=INFO.
			return true, "general", logLevelOption.Value
		}
	}

	return false, "", ""
}
