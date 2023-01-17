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

package kubernetes

import (
	"context"
	"fmt"
	"io"

	"github.com/rs/zerolog"
	core "k8s.io/api/core/v1"

	"github.com/arangodb/kube-arangodb/pkg/debug_package/cli"
	"github.com/arangodb/kube-arangodb/pkg/debug_package/shared"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
)

func Pods() shared.Factory {
	return shared.NewFactory("kubernetes-pods", true, pods)
}

func pods(logger zerolog.Logger, files chan<- shared.File) error {
	k, ok := kclient.GetDefaultFactory().Client()
	if !ok {
		return errors.Newf("Client is not initialised")
	}

	pods, err := ListPods(k)
	if err != nil {
		return err
	}

	podsList := pods.AsList()

	files <- shared.NewYAMLFile("kubernetes/pods.yaml", func() ([]interface{}, error) {
		q := make([]interface{}, len(podsList))

		for id := range podsList {
			q[id] = podsList[id]
		}

		return q, nil
	})

	if cli.GetInput().PodLogs {
		if err := podsLogs(k, files, podsList...); err != nil {
			logger.Err(err).Msgf("Error while collecting pod logs")
		}
	}

	return nil
}

func podsLogs(client kclient.Client, files chan<- shared.File, pods ...*core.Pod) error {
	errs := make([]error, len(pods))

	for id := range pods {
		errs[id] = podLogs(client, files, pods[id])
	}

	return errors.Errors(errs...)
}

func podLogs(client kclient.Client, files chan<- shared.File, pod *core.Pod) error {
	podYaml(files, pod)

	errs := make([]error, 0, len(pod.Status.ContainerStatuses)+len(pod.Status.InitContainerStatuses)+len(pod.Status.EphemeralContainerStatuses))

	if s := pod.Status.ContainerStatuses; len(s) > 0 {
		for id := range s {
			if s[id].State.Waiting != nil {
				continue
			}

			errs = append(errs, errors.Wrapf(podContainerLogs(client, files, pod, s[id].Name), "Unable to read %s Container logs", s[id].Name))
		}
	}

	if s := pod.Status.EphemeralContainerStatuses; len(s) > 0 {
		for id := range s {
			if s[id].State.Waiting != nil {
				continue
			}

			errs = append(errs, errors.Wrapf(podContainerLogs(client, files, pod, s[id].Name), "Unable to read %s EphemeralContainer logs", s[id].Name))
		}
	}

	if s := pod.Status.InitContainerStatuses; len(s) > 0 {
		for id := range s {
			if s[id].State.Waiting != nil {
				continue
			}

			errs = append(errs, errors.Wrapf(podContainerLogs(client, files, pod, s[id].Name), "Unable to read %s InitContainer logs", s[id].Name))
		}
	}

	return errors.Errors(errs...)
}

func podYaml(files chan<- shared.File, pod *core.Pod) {
	files <- shared.NewYAMLFile(fmt.Sprintf("kubernetes/pods/%s/pod.yaml", pod.GetName()), func() ([]interface{}, error) {
		return []interface{}{pod}, nil
	})
}

func podContainerLogs(client kclient.Client, files chan<- shared.File, pod *core.Pod, container string) error {
	res := client.Kubernetes().CoreV1().Pods(pod.GetNamespace()).GetLogs(pod.GetName(), &core.PodLogOptions{
		Container:  container,
		Timestamps: true,
	})

	q, err := res.Stream(context.Background())
	if err != nil {
		return err
	}

	defer q.Close()

	d, err := io.ReadAll(q)
	if err != nil {
		return err
	}

	files <- shared.NewFile(fmt.Sprintf("kubernetes/pods/%s/logs/container/%s", pod.GetName(), container), func() ([]byte, error) {
		return d, nil
	})

	return nil
}
