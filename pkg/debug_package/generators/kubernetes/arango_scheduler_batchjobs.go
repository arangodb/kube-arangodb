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

func schedulerBatchJobs(logger zerolog.Logger, files chan<- shared.File, client kclient.Client) error {
	batchjobs, err := listSchedulerBatchJobs(client)
	if err != nil {
		if kerrors.IsForbiddenOrNotFound(err) {
			return nil
		}

		return err
	}

	if err := errors.ExecuteWithErrorArrayP2(schedulerBatchJob, client, files, batchjobs...); err != nil {
		logger.Err(err).Msgf("Error while collecting arango scheduler batchjobs")
		return err
	}

	return nil
}

func schedulerBatchJob(client kclient.Client, files chan<- shared.File, ext *schedulerApi.ArangoSchedulerBatchJob) error {
	files <- shared.NewYAMLFile(fmt.Sprintf("kubernetes/arango/scheduler/arangoschedulerbatchjobs/%s.yaml", ext.GetName()), func() ([]interface{}, error) {
		return []interface{}{ext}, nil
	})

	return nil
}

func listSchedulerBatchJobs(client kclient.Client) ([]*schedulerApi.ArangoSchedulerBatchJob, error) {
	return ListObjects[*schedulerApi.ArangoSchedulerBatchJobList, *schedulerApi.ArangoSchedulerBatchJob](context.Background(), client.Arango().SchedulerV1beta1().ArangoSchedulerBatchJobs(cli.GetInput().Namespace), func(result *schedulerApi.ArangoSchedulerBatchJobList) []*schedulerApi.ArangoSchedulerBatchJob {
		q := make([]*schedulerApi.ArangoSchedulerBatchJob, len(result.Items))

		for id, e := range result.Items {
			q[id] = e.DeepCopy()
		}

		return q
	})
}
