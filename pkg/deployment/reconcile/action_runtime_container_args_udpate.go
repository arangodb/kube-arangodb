//
// DISCLAIMER
//
// Copyright 2016-2021 ArangoDB GmbH, Cologne, Germany
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
// Author Tomasz Mielech
//

package reconcile

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	core "k8s.io/api/core/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/rotation"
	"github.com/arangodb/kube-arangodb/pkg/util/arangod"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

func init() {
	registerAction(api.ActionTypeRuntimeContainerArgsLogLevelUpdate, runtimeContainerArgsUpdate)
}

func runtimeContainerArgsUpdate(log zerolog.Logger, action api.Action, actionCtx ActionContext) Action {
	a := &actionRuntimeContainerArgsUpdate{}

	a.actionImpl = newActionImplDefRef(log, action, actionCtx, defaultTimeout)

	return a
}

var _ ActionReloadCachedStatus = &actionRuntimeContainerArgsUpdate{}
var _ ActionPost = &actionRuntimeContainerArgsUpdate{}

type actionRuntimeContainerArgsUpdate struct {
	// actionImpl implement timeout and member id functions
	actionImpl
}

// Post updates arguments for the specific Arango member.
func (a actionRuntimeContainerArgsUpdate) Post(ctx context.Context) error {

	m, ok := a.actionCtx.GetMemberStatusByID(a.action.MemberID)
	if !ok {
		a.log.Info().Msg("member is gone already")
		return nil
	}

	memberName := m.ArangoMemberName(a.actionCtx.GetName(), a.action.Group)
	member, ok := a.actionCtx.GetCachedStatus().ArangoMember(memberName)
	if !ok {
		return errors.Errorf("ArangoMember %s not found", memberName)
	}

	containerName, ok := a.action.GetParam(rotation.ContainerName)
	if !ok {
		a.log.Warn().Msgf("Unable to find action's param %s", rotation.ContainerName)
		return nil
	}

	log := a.log.With().Str("containerName", containerName).Logger()
	updateMemberStatusArgs := func(obj *api.ArangoMember, s *api.ArangoMemberStatus) bool {
		if obj.Spec.Template == nil || s.Template == nil ||
			obj.Spec.Template.PodSpec == nil || s.Template.PodSpec == nil {
			log.Info().Msgf("Nil Member definition")
			return false
		}

		if len(obj.Spec.Template.PodSpec.Spec.Containers) != len(s.Template.PodSpec.Spec.Containers) {
			log.Info().Msgf("Invalid size of containers")
			return false
		}

		for id := range obj.Spec.Template.PodSpec.Spec.Containers {
			if obj.Spec.Template.PodSpec.Spec.Containers[id].Name == containerName {
				if s.Template.PodSpec.Spec.Containers[id].Name != containerName {
					log.Info().Msgf("Invalid order of containers")
					return false
				}

				s.Template.PodSpec.Spec.Containers[id].Command = obj.Spec.Template.PodSpec.Spec.Containers[id].Command
				log.Info().Msgf("Updating container args")
				return true
			}
		}

		log.Info().Msgf("can not find the container")

		return false
	}

	err := a.actionCtx.WithArangoMemberStatusUpdate(ctx, member.GetNamespace(), member.GetName(), updateMemberStatusArgs)
	if err != nil {
		return errors.WithMessage(err, "Error while updating member status")
	}

	return nil
}

// ReloadCachedStatus reloads the inspector cache when the action is done.
func (a actionRuntimeContainerArgsUpdate) ReloadCachedStatus() bool {
	return true
}

// Start starts the action for changing conditions on the provided member.
func (a actionRuntimeContainerArgsUpdate) Start(ctx context.Context) (bool, error) {

	m, ok := a.actionCtx.GetMemberStatusByID(a.action.MemberID)
	if !ok {
		a.log.Info().Msg("member is gone already")
		return true, nil
	}

	if !m.Phase.IsReady() {
		a.log.Info().Msg("Member is not ready, unable to run update operation")
		return true, nil
	}

	containerName, ok := a.action.GetParam(rotation.ContainerName)
	if !ok {
		return true, nil
	}

	memberName := m.ArangoMemberName(a.actionCtx.GetName(), a.action.Group)
	member, ok := a.actionCtx.GetCachedStatus().ArangoMember(memberName)
	if !ok {
		return false, errors.Errorf("ArangoMember %s not found", memberName)
	}

	pod, ok := a.actionCtx.GetCachedStatus().Pod(m.PodName)
	if !ok {
		a.log.Info().Str("podName", m.PodName).Msg("pod is not present")
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

		a.log.Info().Interface("topics", topicsLogLevel).Msg("send log level to the ArangoDB")
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
func (a actionRuntimeContainerArgsUpdate) CheckProgress(_ context.Context) (bool, bool, error) {
	return true, false, nil
}

// setLogLevel sets the log's levels for the specific server.
func (a actionRuntimeContainerArgsUpdate) setLogLevel(ctx context.Context, logLevels map[string]string) error {
	if len(logLevels) == 0 {
		return nil
	}

	ctxChild, cancel := context.WithTimeout(ctx, arangod.GetRequestTimeout())
	defer cancel()
	cli, err := a.actionCtx.GetServerClient(ctxChild, a.action.Group, a.action.MemberID)
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

	ctxChild, cancel = context.WithTimeout(ctx, arangod.GetRequestTimeout())
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
