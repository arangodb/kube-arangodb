//
// DISCLAIMER
//
// Copyright 2023-2025 ArangoDB GmbH, Cologne, Germany
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

package arango

import (
	"context"

	schedulerApi "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1beta1"
	"github.com/arangodb/kube-arangodb/pkg/debug_package/shared"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/list"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
)

func Scheduler(f shared.FactoryGen) {
	f.AddSection("scheduler").
		Register("profile", true, shared.WithKubernetesItems[*schedulerApi.ArangoProfile](arangoSchedulerV1Beta1ArangoProfileList, shared.WithDefinitions[*schedulerApi.ArangoProfile])).
		Register("batchjob", true, shared.WithKubernetesItems[*schedulerApi.ArangoSchedulerBatchJob](arangoSchedulerV1Beta1ArangoSchedulerBatchJobList, shared.WithDefinitions[*schedulerApi.ArangoSchedulerBatchJob])).
		Register("cronjob", true, shared.WithKubernetesItems[*schedulerApi.ArangoSchedulerCronJob](arangoSchedulerV1Beta1ArangoSchedulerCronJobList, shared.WithDefinitions[*schedulerApi.ArangoSchedulerCronJob])).
		Register("deployment", true, shared.WithKubernetesItems[*schedulerApi.ArangoSchedulerDeployment](arangoSchedulerV1Beta1ArangoSchedulerDeploymentList, shared.WithDefinitions[*schedulerApi.ArangoSchedulerDeployment])).
		Register("pod", true, shared.WithKubernetesItems[*schedulerApi.ArangoSchedulerPod](arangoSchedulerV1Beta1ArangoSchedulerPodList, shared.WithDefinitions[*schedulerApi.ArangoSchedulerPod]))
}

func arangoSchedulerV1Beta1ArangoProfileList(ctx context.Context, client kclient.Client, namespace string) ([]*schedulerApi.ArangoProfile, error) {
	return list.ListObjects[*schedulerApi.ArangoProfileList, *schedulerApi.ArangoProfile](ctx, client.Arango().SchedulerV1beta1().ArangoProfiles(namespace), func(result *schedulerApi.ArangoProfileList) []*schedulerApi.ArangoProfile {
		q := make([]*schedulerApi.ArangoProfile, len(result.Items))

		for id, e := range result.Items {
			q[id] = e.DeepCopy()
		}

		return q
	})
}

func arangoSchedulerV1Beta1ArangoSchedulerBatchJobList(ctx context.Context, client kclient.Client, namespace string) ([]*schedulerApi.ArangoSchedulerBatchJob, error) {
	return list.ListObjects[*schedulerApi.ArangoSchedulerBatchJobList, *schedulerApi.ArangoSchedulerBatchJob](ctx, client.Arango().SchedulerV1beta1().ArangoSchedulerBatchJobs(namespace), func(result *schedulerApi.ArangoSchedulerBatchJobList) []*schedulerApi.ArangoSchedulerBatchJob {
		q := make([]*schedulerApi.ArangoSchedulerBatchJob, len(result.Items))

		for id, e := range result.Items {
			q[id] = e.DeepCopy()
		}

		return q
	})
}

func arangoSchedulerV1Beta1ArangoSchedulerCronJobList(ctx context.Context, client kclient.Client, namespace string) ([]*schedulerApi.ArangoSchedulerCronJob, error) {
	return list.ListObjects[*schedulerApi.ArangoSchedulerCronJobList, *schedulerApi.ArangoSchedulerCronJob](ctx, client.Arango().SchedulerV1beta1().ArangoSchedulerCronJobs(namespace), func(result *schedulerApi.ArangoSchedulerCronJobList) []*schedulerApi.ArangoSchedulerCronJob {
		q := make([]*schedulerApi.ArangoSchedulerCronJob, len(result.Items))

		for id, e := range result.Items {
			q[id] = e.DeepCopy()
		}

		return q
	})
}

func arangoSchedulerV1Beta1ArangoSchedulerDeploymentList(ctx context.Context, client kclient.Client, namespace string) ([]*schedulerApi.ArangoSchedulerDeployment, error) {
	return list.ListObjects[*schedulerApi.ArangoSchedulerDeploymentList, *schedulerApi.ArangoSchedulerDeployment](ctx, client.Arango().SchedulerV1beta1().ArangoSchedulerDeployments(namespace), func(result *schedulerApi.ArangoSchedulerDeploymentList) []*schedulerApi.ArangoSchedulerDeployment {
		q := make([]*schedulerApi.ArangoSchedulerDeployment, len(result.Items))

		for id, e := range result.Items {
			q[id] = e.DeepCopy()
		}

		return q
	})
}

func arangoSchedulerV1Beta1ArangoSchedulerPodList(ctx context.Context, client kclient.Client, namespace string) ([]*schedulerApi.ArangoSchedulerPod, error) {
	return list.ListObjects[*schedulerApi.ArangoSchedulerPodList, *schedulerApi.ArangoSchedulerPod](ctx, client.Arango().SchedulerV1beta1().ArangoSchedulerPods(namespace), func(result *schedulerApi.ArangoSchedulerPodList) []*schedulerApi.ArangoSchedulerPod {
		q := make([]*schedulerApi.ArangoSchedulerPod, len(result.Items))

		for id, e := range result.Items {
			q[id] = e.DeepCopy()
		}

		return q
	})
}
