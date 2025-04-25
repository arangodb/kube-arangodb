//
// DISCLAIMER
//
// Copyright 2025 ArangoDB GmbH, Cologne, Germany
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

	batch "k8s.io/api/batch/v1"

	"github.com/arangodb/kube-arangodb/pkg/debug_package/shared"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/list"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
)

func KubernetesBatch(f shared.FactoryGen) {
	f.AddSection("batch").
		Register("job", true, shared.WithKubernetesItems[*batch.Job](kubernetesBatchJobsList, shared.WithDefinitions[*batch.Job])).
		Register("cronjob", true, shared.WithKubernetesItems[*batch.CronJob](kubernetesBatchCronJobsList, shared.WithDefinitions[*batch.CronJob]))
}

func kubernetesBatchJobsList(ctx context.Context, client kclient.Client, namespace string) ([]*batch.Job, error) {
	return list.ListObjects[*batch.JobList, *batch.Job](ctx, client.Kubernetes().BatchV1().Jobs(namespace), func(result *batch.JobList) []*batch.Job {
		q := make([]*batch.Job, len(result.Items))

		for id, e := range result.Items {
			q[id] = e.DeepCopy()
		}

		return q
	})
}

func kubernetesBatchCronJobsList(ctx context.Context, client kclient.Client, namespace string) ([]*batch.CronJob, error) {
	return list.ListObjects[*batch.CronJobList, *batch.CronJob](ctx, client.Kubernetes().BatchV1().CronJobs(namespace), func(result *batch.CronJobList) []*batch.CronJob {
		q := make([]*batch.CronJob, len(result.Items))

		for id, e := range result.Items {
			q[id] = e.DeepCopy()
		}

		return q
	})
}
