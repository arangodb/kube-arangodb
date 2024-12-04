//
// DISCLAIMER
//
// Copyright 2024 ArangoDB GmbH, Cologne, Germany
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

package shutdown

import (
	"context"
	"fmt"
	"strconv"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	core "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	pbSharedV1 "github.com/arangodb/kube-arangodb/integrations/shared/v1/definition"
	pbShutdownV1 "github.com/arangodb/kube-arangodb/integrations/shutdown/v1/definition"
	"github.com/arangodb/kube-arangodb/pkg/logging"
	operator "github.com/arangodb/kube-arangodb/pkg/operatorV2"
	"github.com/arangodb/kube-arangodb/pkg/operatorV2/event"
	"github.com/arangodb/kube-arangodb/pkg/operatorV2/operation"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/constants"
)

var logger = logging.Global().RegisterAndGetLogger("platform-pod-shutdown", logging.Info)

type handler struct {
	kubeClient kubernetes.Interface

	eventRecorder event.RecorderInstance

	operator operator.Operator
}

func (h *handler) Name() string {
	return Kind()
}

func (h *handler) Handle(ctx context.Context, item operation.Item) error {
	pod, err := util.WithKubernetesContextTimeoutP2A2(ctx, h.kubeClient.CoreV1().Pods(item.Namespace).Get, item.Name, meta.GetOptions{})
	if err != nil {
		if apiErrors.IsNotFound(err) {
			return nil
		}

		return err
	}

	// If not annotated, stop execution
	if _, ok := pod.Annotations[constants.AnnotationShutdownManagedContainer]; !ok {
		return nil
	}

	for _, container := range pod.Status.ContainerStatuses {
		v, ok := pod.Annotations[fmt.Sprintf("%s/%s", constants.AnnotationShutdownCoreContainer, container.Name)]
		if !ok {
			continue
		}

		switch v {
		case constants.AnnotationShutdownCoreContainerModeWait:
			if container.State.Terminated == nil {
				// Container is not yet stopped, skip shutdown
				return nil
			}
		}
	}

	// All containers, which are expected to shutdown, are down

	for _, container := range pod.Status.ContainerStatuses {
		v, ok := pod.Annotations[fmt.Sprintf("%s/%s", constants.AnnotationShutdownContainer, container.Name)]
		if !ok {
			continue
		}

		// We did not reach running state, nothing to do
		if container.State.Running == nil {
			continue
		}

		port, ok := h.getContainerPort(pod.Spec.Containers, container.Name, v)
		if !ok {
			// We did not find port, continue
			continue
		}

		if port.ContainerPort == 0 {
			continue
		}

		if pod.Status.PodIP == "" {
			continue
		}

		if err := util.WithKubernetesContextTimeoutP1A1(ctx, h.invokeShutdown, fmt.Sprintf("%s:%d", pod.Status.PodIP, port.ContainerPort)); err != nil {
			logger.WrapObj(item).Err(err).Str("container", container.Name).Debug("Unable to send shutdown request")
		}

		logger.WrapObj(item).Str("container", container.Name).Debug("Shutdown request sent")
	}

	// Always return nil
	return nil
}

func (h *handler) CanBeHandled(item operation.Item) bool {
	return item.Group == Group() &&
		item.Version == Version() &&
		item.Kind == Kind()
}

func (h *handler) getContainerPort(containers []core.Container, container, port string) (core.ContainerPort, bool) {
	if v, err := strconv.Atoi(port); err == nil {
		return core.ContainerPort{
			ContainerPort: int32(v),
		}, true
	}

	for _, c := range containers {
		if c.Name != container {
			continue
		}

		for _, p := range c.Ports {
			if p.Name == port {
				return p, true
			}
		}
	}

	return core.ContainerPort{}, false
}

func (h *handler) invokeShutdown(ctx context.Context, addr string) error {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}

	defer conn.Close()

	client := pbShutdownV1.NewShutdownV1Client(conn)

	if _, err := client.Shutdown(ctx, &pbSharedV1.Empty{}); err != nil {
		return err
	}

	return nil
}
