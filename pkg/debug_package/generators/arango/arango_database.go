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

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/debug_package/shared"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/list"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
)

func Database(f shared.FactoryGen) {
	f.AddSection("database").
		Register("deployment", true, shared.WithKubernetesItems[*api.ArangoDeployment](arangoDatabaseV1ArangoDeploymentList,
			shared.WithDefinitions[*api.ArangoDeployment],
			arangoDatabaseDeploymentMembers,
			arangoDatabaseDeploymentAgencyDump,
			arangoDatabaseDeploymentPlatform)).
		Register("member", true, shared.WithKubernetesItems[*api.ArangoMember](arangoDatabaseV1ArangoMemberList, shared.WithDefinitions[*api.ArangoMember])).
		Register("task", true, shared.WithKubernetesItems[*api.ArangoTask](arangoDatabaseV1ArangoTaskList, shared.WithDefinitions[*api.ArangoTask])).
		Register("acs", true, shared.WithKubernetesItems[*api.ArangoClusterSynchronization](arangoDatabaseV1ArangoClusterSynchronizationList, shared.WithDefinitions[*api.ArangoClusterSynchronization]))
}

func arangoDatabaseV1ArangoDeploymentList(ctx context.Context, client kclient.Client, namespace string) ([]*api.ArangoDeployment, error) {
	return list.ListObjects[*api.ArangoDeploymentList, *api.ArangoDeployment](ctx, client.Arango().DatabaseV1().ArangoDeployments(namespace), func(result *api.ArangoDeploymentList) []*api.ArangoDeployment {
		q := make([]*api.ArangoDeployment, len(result.Items))

		for id, e := range result.Items {
			q[id] = e.DeepCopy()
		}

		return q
	})
}

func arangoDatabaseV1ArangoTaskList(ctx context.Context, client kclient.Client, namespace string) ([]*api.ArangoTask, error) {
	return list.ListObjects[*api.ArangoTaskList, *api.ArangoTask](ctx, client.Arango().DatabaseV1().ArangoTasks(namespace), func(result *api.ArangoTaskList) []*api.ArangoTask {
		q := make([]*api.ArangoTask, len(result.Items))

		for id, e := range result.Items {
			q[id] = e.DeepCopy()
		}

		return q
	})
}

func arangoDatabaseV1ArangoMemberList(ctx context.Context, client kclient.Client, namespace string) ([]*api.ArangoMember, error) {
	return list.ListObjects[*api.ArangoMemberList, *api.ArangoMember](ctx, client.Arango().DatabaseV1().ArangoMembers(namespace), func(result *api.ArangoMemberList) []*api.ArangoMember {
		q := make([]*api.ArangoMember, len(result.Items))

		for id, e := range result.Items {
			q[id] = e.DeepCopy()
		}

		return q
	})
}

func arangoDatabaseV1ArangoClusterSynchronizationList(ctx context.Context, client kclient.Client, namespace string) ([]*api.ArangoClusterSynchronization, error) {
	return list.ListObjects[*api.ArangoClusterSynchronizationList, *api.ArangoClusterSynchronization](ctx, client.Arango().DatabaseV1().ArangoClusterSynchronizations(namespace), func(result *api.ArangoClusterSynchronizationList) []*api.ArangoClusterSynchronization {
		q := make([]*api.ArangoClusterSynchronization, len(result.Items))

		for id, e := range result.Items {
			q[id] = e.DeepCopy()
		}

		return q
	})
}
