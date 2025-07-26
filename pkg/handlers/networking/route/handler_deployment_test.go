//
// DISCLAIMER
//
// Copyright 2024-2025 ArangoDB GmbH, Cologne, Germany
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

package route

import (
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/uuid"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	networkingApi "github.com/arangodb/kube-arangodb/pkg/apis/networking/v1beta1"
	sharedApi "github.com/arangodb/kube-arangodb/pkg/apis/shared/v1"
	"github.com/arangodb/kube-arangodb/pkg/operatorV2/operation"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/tests"
)

func Test_Handler_Deployment(t *testing.T) {
	handler := newFakeHandler()

	// Arrange
	extension := tests.NewMetaObject[*networkingApi.ArangoRoute](t, tests.FakeNamespace, "test",
		func(t *testing.T, obj *networkingApi.ArangoRoute) {
			obj.Spec.Deployment = util.NewType("deployment")
		},
		func(t *testing.T, obj *networkingApi.ArangoRoute) {
			obj.Spec.Destination = &networkingApi.ArangoRouteSpecDestination{
				Service: &networkingApi.ArangoRouteSpecDestinationService{
					Object: &sharedApi.Object{
						Name: "deployment",
					},
					Port: util.NewType(intstr.FromInt32(10244)),
				},
			}
		})
	deployment := tests.NewMetaObject[*api.ArangoDeployment](t, tests.FakeNamespace, "deployment")

	refresh := tests.CreateObjects(t, handler.kubeClient, handler.client, &deployment, &extension)

	// Test
	require.NoError(t, tests.Handle(handler, tests.NewItem(t, operation.Update, extension)))

	// Refresh
	refresh(t)

	// Assert
	require.True(t, extension.Status.Conditions.IsTrue(networkingApi.SpecValidCondition))
	require.True(t, extension.Status.Conditions.IsTrue(networkingApi.DeploymentFoundCondition))
}

func Test_Handler_MissingDeployment(t *testing.T) {
	handler := newFakeHandler()

	// Arrange
	extension := tests.NewMetaObject[*networkingApi.ArangoRoute](t, tests.FakeNamespace, "test", func(t *testing.T, obj *networkingApi.ArangoRoute) {
		obj.Spec.Deployment = util.NewType("deployment-missing")
	},
		func(t *testing.T, obj *networkingApi.ArangoRoute) {
			obj.Spec.Destination = &networkingApi.ArangoRouteSpecDestination{
				Service: &networkingApi.ArangoRouteSpecDestinationService{
					Object: &sharedApi.Object{
						Name: "deployment",
					},
					Port: util.NewType(intstr.FromInt32(10244)),
				},
			}
		})
	deployment := tests.NewMetaObject[*api.ArangoDeployment](t, tests.FakeNamespace, "deployment")

	refresh := tests.CreateObjects(t, handler.kubeClient, handler.client, &deployment, &extension)

	// Test
	require.NoError(t, tests.Handle(handler, tests.NewItem(t, operation.Update, extension)))

	// Refresh
	refresh(t)

	// Assert
	require.True(t, extension.Status.Conditions.IsTrue(networkingApi.SpecValidCondition))
	require.False(t, extension.Status.Conditions.IsTrue(networkingApi.DeploymentFoundCondition))
}

func Test_Handler_Deployment_Changed(t *testing.T) {
	handler := newFakeHandler()

	// Arrange
	extension := tests.NewMetaObject[*networkingApi.ArangoRoute](t, tests.FakeNamespace, "test", func(t *testing.T, obj *networkingApi.ArangoRoute) {
		obj.Spec.Deployment = util.NewType("deployment")
	},
		func(t *testing.T, obj *networkingApi.ArangoRoute) {
			obj.Spec.Destination = &networkingApi.ArangoRouteSpecDestination{
				Service: &networkingApi.ArangoRouteSpecDestinationService{
					Object: &sharedApi.Object{
						Name: "deployment",
					},
					Port: util.NewType(intstr.FromInt32(10244)),
				},
			}
		})
	deployment := tests.NewMetaObject[*api.ArangoDeployment](t, tests.FakeNamespace, "deployment")
	refresh := tests.CreateObjects(t, handler.kubeClient, handler.client, &deployment, &extension)

	// Test
	require.NoError(t, tests.Handle(handler, tests.NewItem(t, operation.Update, extension)))

	// Refresh
	refresh(t)

	// Assert
	require.True(t, extension.Status.Conditions.IsTrue(networkingApi.DeploymentFoundCondition))

	deployment.UID = uuid.NewUUID()

	tests.UpdateObjects(t, handler.kubeClient, handler.client, &deployment)

	// Test
	require.NoError(t, tests.Handle(handler, tests.NewItem(t, operation.Update, extension)))

	// Refresh
	refresh(t)

	// Assert
	require.True(t, extension.Status.Conditions.IsTrue(networkingApi.SpecValidCondition))
	require.False(t, extension.Status.Conditions.IsTrue(networkingApi.DeploymentFoundCondition))
}
