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

	mlApiv1alpha1 "github.com/arangodb/kube-arangodb/pkg/apis/ml/v1alpha1"
	mlApi "github.com/arangodb/kube-arangodb/pkg/apis/ml/v1beta1"
	"github.com/arangodb/kube-arangodb/pkg/debug_package/shared"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/list"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
)

func ML(f shared.FactoryGen) {
	f.AddSection("ml").
		Register("batchjob", true, shared.WithKubernetesItems[*mlApiv1alpha1.ArangoMLBatchJob](arangoMLV1Alpha1ArangoMLBatchJobList, shared.WithDefinitions[*mlApiv1alpha1.ArangoMLBatchJob])).
		Register("cronjob", true, shared.WithKubernetesItems[*mlApiv1alpha1.ArangoMLCronJob](arangoMLV1Alpha1ArangoMLCronJobList, shared.WithDefinitions[*mlApiv1alpha1.ArangoMLCronJob])).
		Register("extension", true, shared.WithKubernetesItems[*mlApi.ArangoMLExtension](arangoMLV1Beta1ArangoMLExtensionList, shared.WithDefinitions[*mlApi.ArangoMLExtension])).
		Register("storage", true, shared.WithKubernetesItems[*mlApi.ArangoMLStorage](arangoMLV1Beta1ArangoMLStorageList, shared.WithDefinitions[*mlApi.ArangoMLStorage]))
}

func arangoMLV1Alpha1ArangoMLBatchJobList(ctx context.Context, client kclient.Client, namespace string) ([]*mlApiv1alpha1.ArangoMLBatchJob, error) {
	return list.ListObjects[*mlApiv1alpha1.ArangoMLBatchJobList, *mlApiv1alpha1.ArangoMLBatchJob](ctx, client.Arango().MlV1alpha1().ArangoMLBatchJobs(namespace), func(result *mlApiv1alpha1.ArangoMLBatchJobList) []*mlApiv1alpha1.ArangoMLBatchJob {
		q := make([]*mlApiv1alpha1.ArangoMLBatchJob, len(result.Items))

		for id, e := range result.Items {
			q[id] = e.DeepCopy()
		}

		return q
	})
}

func arangoMLV1Alpha1ArangoMLCronJobList(ctx context.Context, client kclient.Client, namespace string) ([]*mlApiv1alpha1.ArangoMLCronJob, error) {
	return list.ListObjects[*mlApiv1alpha1.ArangoMLCronJobList, *mlApiv1alpha1.ArangoMLCronJob](ctx, client.Arango().MlV1alpha1().ArangoMLCronJobs(namespace), func(result *mlApiv1alpha1.ArangoMLCronJobList) []*mlApiv1alpha1.ArangoMLCronJob {
		q := make([]*mlApiv1alpha1.ArangoMLCronJob, len(result.Items))

		for id, e := range result.Items {
			q[id] = e.DeepCopy()
		}

		return q
	})
}

func arangoMLV1Beta1ArangoMLExtensionList(ctx context.Context, client kclient.Client, namespace string) ([]*mlApi.ArangoMLExtension, error) {
	return list.ListObjects[*mlApi.ArangoMLExtensionList, *mlApi.ArangoMLExtension](ctx, client.Arango().MlV1beta1().ArangoMLExtensions(namespace), func(result *mlApi.ArangoMLExtensionList) []*mlApi.ArangoMLExtension {
		q := make([]*mlApi.ArangoMLExtension, len(result.Items))

		for id, e := range result.Items {
			q[id] = e.DeepCopy()
		}

		return q
	})
}

func arangoMLV1Beta1ArangoMLStorageList(ctx context.Context, client kclient.Client, namespace string) ([]*mlApi.ArangoMLStorage, error) {
	return list.ListObjects[*mlApi.ArangoMLStorageList, *mlApi.ArangoMLStorage](ctx, client.Arango().MlV1beta1().ArangoMLStorages(namespace), func(result *mlApi.ArangoMLStorageList) []*mlApi.ArangoMLStorage {
		q := make([]*mlApi.ArangoMLStorage, len(result.Items))

		for id, e := range result.Items {
			q[id] = e.DeepCopy()
		}

		return q
	})
}
