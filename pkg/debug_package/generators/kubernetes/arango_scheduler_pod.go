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

package kubernetes

import (
	"context"
	"fmt"

	"github.com/rs/zerolog"

	schedulerApi "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1beta1"
	"github.com/arangodb/kube-arangodb/pkg/debug_package/cli"
	"github.com/arangodb/kube-arangodb/pkg/debug_package/shared"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/kerrors"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
)

func schedulerPods(logger zerolog.Logger, files chan<- shared.File, client kclient.Client) error {
	pods, err := listSchedulerPods(client)
	if err != nil {
		if kerrors.IsForbiddenOrNotFound(err) {
			return nil
		}

		return err
	}

	if err := errors.ExecuteWithErrorArrayP2(schedulerPod, client, files, pods...); err != nil {
		logger.Err(err).Msgf("Error while collecting arango scheduler pods")
		return err
	}

	return nil
}

func schedulerPod(client kclient.Client, files chan<- shared.File, ext *schedulerApi.ArangoSchedulerPod) error {
	files <- shared.NewYAMLFile(fmt.Sprintf("kubernetes/arango/scheduler/arangoschedulerpods/%s.yaml", ext.GetName()), func() ([]interface{}, error) {
		return []interface{}{ext}, nil
	})

	return nil
}

func listSchedulerPods(client kclient.Client) ([]*schedulerApi.ArangoSchedulerPod, error) {
	return ListObjects[*schedulerApi.ArangoSchedulerPodList, *schedulerApi.ArangoSchedulerPod](context.Background(), client.Arango().SchedulerV1beta1().ArangoSchedulerPods(cli.GetInput().Namespace), func(result *schedulerApi.ArangoSchedulerPodList) []*schedulerApi.ArangoSchedulerPod {
		q := make([]*schedulerApi.ArangoSchedulerPod, len(result.Items))

		for id, e := range result.Items {
			q[id] = e.DeepCopy()
		}

		return q
	})
}
