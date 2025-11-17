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

package route

import (
	goHttp "net/http"
	"testing"

	"github.com/stretchr/testify/require"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	networkingApi "github.com/arangodb/kube-arangodb/pkg/apis/networking/v1beta1"
	"github.com/arangodb/kube-arangodb/pkg/operatorV2/operation"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/tests"
)

func Test_Handler_Redirect_Valid(t *testing.T) {
	// Setup
	handler := newFakeHandler()

	// Arrange
	extension := tests.NewMetaObject[*networkingApi.ArangoRoute](t, tests.FakeNamespace, "test",
		func(t *testing.T, obj *networkingApi.ArangoRoute) {
			obj.Spec.Deployment = util.NewType("deployment")
		},
		func(t *testing.T, obj *networkingApi.ArangoRoute) {
			obj.Spec.Destination = &networkingApi.ArangoRouteSpecDestination{
				Redirect: &networkingApi.ArangoRouteSpecDestinationRedirect{},
				Path:     util.NewType("/ui/"),
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
	require.True(t, extension.Status.Conditions.IsTrue(networkingApi.DestinationValidCondition))
	require.True(t, extension.Status.Conditions.IsTrue(networkingApi.ReadyCondition))
	require.Equal(t, networkingApi.ArangoRouteStatusTargetRedirectType, extension.Status.Target.Type)

	require.EqualValues(t, extension.Status.Target.Path, "/ui/")
	require.EqualValues(t, extension.Status.Target.Redirect.Code, goHttp.StatusTemporaryRedirect)

	c, ok := extension.Status.Conditions.Get(networkingApi.DestinationValidCondition)
	require.True(t, ok)
	require.EqualValues(t, c.Reason, "Destination Found")
	require.EqualValues(t, c.Message, "Destination Found")
	require.EqualValues(t, c.Hash, extension.Status.Target.Hash())
}

func Test_Handler_Redirect_Invalid_Code(t *testing.T) {
	// Setup
	handler := newFakeHandler()

	// Arrange
	extension := tests.NewMetaObject[*networkingApi.ArangoRoute](t, tests.FakeNamespace, "test",
		func(t *testing.T, obj *networkingApi.ArangoRoute) {
			obj.Spec.Deployment = util.NewType("deployment")
		},
		func(t *testing.T, obj *networkingApi.ArangoRoute) {
			obj.Spec.Destination = &networkingApi.ArangoRouteSpecDestination{
				Redirect: &networkingApi.ArangoRouteSpecDestinationRedirect{
					Code: util.NewType(goHttp.StatusNotFound),
				},
				Path: util.NewType("/ui/"),
			}
		})
	deployment := tests.NewMetaObject[*api.ArangoDeployment](t, tests.FakeNamespace, "deployment")

	refresh := tests.CreateObjects(t, handler.kubeClient, handler.client, &deployment, &extension)

	// Test
	require.NoError(t, tests.Handle(handler, tests.NewItem(t, operation.Update, extension)))

	// Refresh
	refresh(t)

	// Assert
	require.False(t, extension.Status.Conditions.IsTrue(networkingApi.SpecValidCondition))
	require.False(t, extension.Status.Conditions.IsTrue(networkingApi.ReadyCondition))
	require.Nil(t, extension.Status.Target)

	c, ok := extension.Status.Conditions.Get(networkingApi.SpecValidCondition)
	require.True(t, ok)
	require.EqualValues(t, c.Reason, "Spec is invalid")
	require.EqualValues(t, c.Message, "Received 1 errors: spec.destination.redirect.code: Invalid code. Got 404, allowed 301 & 307")
}

func Test_Handler_Redirect_Invalid_Target(t *testing.T) {
	// Setup
	handler := newFakeHandler()

	// Arrange
	extension := tests.NewMetaObject[*networkingApi.ArangoRoute](t, tests.FakeNamespace, "test",
		func(t *testing.T, obj *networkingApi.ArangoRoute) {
			obj.Spec.Deployment = util.NewType("deployment")
		},
		func(t *testing.T, obj *networkingApi.ArangoRoute) {
			obj.Spec.Destination = &networkingApi.ArangoRouteSpecDestination{
				Redirect: &networkingApi.ArangoRouteSpecDestinationRedirect{},
				Path:     util.NewType("&643/ui/"),
			}
		})
	deployment := tests.NewMetaObject[*api.ArangoDeployment](t, tests.FakeNamespace, "deployment")

	refresh := tests.CreateObjects(t, handler.kubeClient, handler.client, &deployment, &extension)

	// Test
	require.NoError(t, tests.Handle(handler, tests.NewItem(t, operation.Update, extension)))

	// Refresh
	refresh(t)

	// Assert
	require.False(t, extension.Status.Conditions.IsTrue(networkingApi.SpecValidCondition))
	require.False(t, extension.Status.Conditions.IsTrue(networkingApi.ReadyCondition))
	require.Nil(t, extension.Status.Target)

	c, ok := extension.Status.Conditions.Get(networkingApi.SpecValidCondition)
	require.True(t, ok)
	require.EqualValues(t, c.Reason, "Spec is invalid")
	require.EqualValues(t, c.Message, "Received 1 errors: spec.destination.path: String '&643/ui/' is not a valid api path")
}

func Test_Handler_Redirect_Invalid_Missing(t *testing.T) {
	// Setup
	handler := newFakeHandler()

	// Arrange
	extension := tests.NewMetaObject[*networkingApi.ArangoRoute](t, tests.FakeNamespace, "test",
		func(t *testing.T, obj *networkingApi.ArangoRoute) {
			obj.Spec.Deployment = util.NewType("deployment")
		},
		func(t *testing.T, obj *networkingApi.ArangoRoute) {
			obj.Spec.Destination = &networkingApi.ArangoRouteSpecDestination{}
		})
	deployment := tests.NewMetaObject[*api.ArangoDeployment](t, tests.FakeNamespace, "deployment")

	refresh := tests.CreateObjects(t, handler.kubeClient, handler.client, &deployment, &extension)

	// Test
	require.NoError(t, tests.Handle(handler, tests.NewItem(t, operation.Update, extension)))

	// Refresh
	refresh(t)

	// Assert
	require.False(t, extension.Status.Conditions.IsTrue(networkingApi.SpecValidCondition))
	require.False(t, extension.Status.Conditions.IsTrue(networkingApi.ReadyCondition))
	require.Nil(t, extension.Status.Target)

	c, ok := extension.Status.Conditions.Get(networkingApi.SpecValidCondition)
	require.True(t, ok)
	require.EqualValues(t, c.Reason, "Spec is invalid")
	require.EqualValues(t, c.Message, "Received 1 errors: spec.destination: Elements not provided. Expected 1. Possible: endpoints, redirect, service")
}

func Test_Handler_Redirect_Empty(t *testing.T) {
	// Setup
	handler := newFakeHandler()

	// Arrange
	extension := tests.NewMetaObject[*networkingApi.ArangoRoute](t, tests.FakeNamespace, "test",
		func(t *testing.T, obj *networkingApi.ArangoRoute) {
			obj.Spec.Deployment = util.NewType("deployment")
		},
		func(t *testing.T, obj *networkingApi.ArangoRoute) {
			obj.Spec.Destination = &networkingApi.ArangoRouteSpecDestination{
				Redirect: &networkingApi.ArangoRouteSpecDestinationRedirect{},
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
	require.True(t, extension.Status.Conditions.IsTrue(networkingApi.DestinationValidCondition))
	require.True(t, extension.Status.Conditions.IsTrue(networkingApi.ReadyCondition))
	require.Equal(t, networkingApi.ArangoRouteStatusTargetRedirectType, extension.Status.Target.Type)

	require.EqualValues(t, extension.Status.Target.Path, "/")
	require.EqualValues(t, extension.Status.Target.Redirect.Code, goHttp.StatusTemporaryRedirect)

	c, ok := extension.Status.Conditions.Get(networkingApi.DestinationValidCondition)
	require.True(t, ok)
	require.EqualValues(t, c.Reason, "Destination Found")
	require.EqualValues(t, c.Message, "Destination Found")
	require.EqualValues(t, c.Hash, extension.Status.Target.Hash())
}

func Test_Handler_Redirect_Valid_CustomCode(t *testing.T) {
	// Setup
	handler := newFakeHandler()

	// Arrange
	extension := tests.NewMetaObject[*networkingApi.ArangoRoute](t, tests.FakeNamespace, "test",
		func(t *testing.T, obj *networkingApi.ArangoRoute) {
			obj.Spec.Deployment = util.NewType("deployment")
		},
		func(t *testing.T, obj *networkingApi.ArangoRoute) {
			obj.Spec.Destination = &networkingApi.ArangoRouteSpecDestination{
				Redirect: &networkingApi.ArangoRouteSpecDestinationRedirect{
					Code: util.NewType(goHttp.StatusMovedPermanently),
				},
				Path: util.NewType("/ui/"),
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
	require.True(t, extension.Status.Conditions.IsTrue(networkingApi.DestinationValidCondition))
	require.True(t, extension.Status.Conditions.IsTrue(networkingApi.ReadyCondition))
	require.Equal(t, networkingApi.ArangoRouteStatusTargetRedirectType, extension.Status.Target.Type)

	require.EqualValues(t, extension.Status.Target.Path, "/ui/")
	require.EqualValues(t, extension.Status.Target.Redirect.Code, goHttp.StatusMovedPermanently)

	c, ok := extension.Status.Conditions.Get(networkingApi.DestinationValidCondition)
	require.True(t, ok)
	require.EqualValues(t, c.Reason, "Destination Found")
	require.EqualValues(t, c.Message, "Destination Found")
	require.EqualValues(t, c.Hash, extension.Status.Target.Hash())
}
