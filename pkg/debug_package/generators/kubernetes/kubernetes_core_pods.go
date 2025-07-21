//
// DISCLAIMER
//
// Copyright 2016-2025 ArangoDB GmbH, Cologne, Germany
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
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
)

func kubernetesCorePodLogs(ctx context.Context, logger zerolog.Logger, client kclient.Client, files chan<- shared.File, item *core.Pod) error {
	if !cli.GetInput().PodLogs {
		return nil
	}

	if s := item.Status.ContainerStatuses; len(s) > 0 {
		for id := range s {
			files <- kubernetesCorePodLogsExtract(ctx, client, item, s[id].Name)

			if s[id].RestartCount > 0 {
				files <- kubernetesPreviousCorePodLogsExtract(ctx, client, item, s[id].Name)
			}
		}
	}

	if s := item.Status.EphemeralContainerStatuses; len(s) > 0 {
		for id := range s {
			files <- kubernetesCorePodLogsExtract(ctx, client, item, s[id].Name)

			if s[id].RestartCount > 0 {
				files <- kubernetesPreviousCorePodLogsExtract(ctx, client, item, s[id].Name)
			}
		}
	}

	if s := item.Status.InitContainerStatuses; len(s) > 0 {
		for id := range s {
			files <- kubernetesCorePodLogsExtract(ctx, client, item, s[id].Name)

			if s[id].RestartCount > 0 {
				files <- kubernetesPreviousCorePodLogsExtract(ctx, client, item, s[id].Name)
			}
		}
	}

	return nil
}

func kubernetesPreviousCorePodLogsExtract(ctx context.Context, client kclient.Client, item *core.Pod, container string) shared.File {
	return shared.NewFile(fmt.Sprintf("logs/container/%s.previous", container), func() ([]byte, error) {
		res := client.Kubernetes().CoreV1().Pods(item.GetNamespace()).GetLogs(item.GetName(), &core.PodLogOptions{
			Container:  container,
			Timestamps: true,
			Previous:   true,
		})

		q, err := res.Stream(ctx)
		if err != nil {
			return nil, err
		}

		defer q.Close()

		d, err := io.ReadAll(q)
		if err != nil {
			return nil, err
		}

		return d, nil
	})
}

func kubernetesCorePodLogsExtract(ctx context.Context, client kclient.Client, item *core.Pod, container string) shared.File {
	return shared.NewFile(fmt.Sprintf("logs/container/%s", container), func() ([]byte, error) {
		res := client.Kubernetes().CoreV1().Pods(item.GetNamespace()).GetLogs(item.GetName(), &core.PodLogOptions{
			Container:  container,
			Timestamps: true,
		})

		q, err := res.Stream(ctx)
		if err != nil {
			return nil, err
		}

		defer q.Close()

		d, err := io.ReadAll(q)
		if err != nil {
			return nil, err
		}

		return d, nil
	})
}
