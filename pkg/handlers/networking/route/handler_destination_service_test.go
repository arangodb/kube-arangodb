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

package route

import (
	"testing"

	"github.com/stretchr/testify/require"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	networkingApi "github.com/arangodb/kube-arangodb/pkg/apis/networking/v1alpha1"
	sharedApi "github.com/arangodb/kube-arangodb/pkg/apis/shared/v1"
	"github.com/arangodb/kube-arangodb/pkg/operatorV2/operation"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/tests"
)

func Test_Handler_Destination_Service_Missing(t *testing.T) {
	// Setup
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
	require.False(t, extension.Status.Conditions.IsTrue(networkingApi.DestinationValidCondition))
	require.False(t, extension.Status.Conditions.IsTrue(networkingApi.ReadyCondition))

	c, ok := extension.Status.Conditions.Get(networkingApi.DestinationValidCondition)
	require.True(t, ok)
	require.EqualValues(t, c.Reason, "Destination Not Found")
	require.EqualValues(t, c.Message, "Unknown error for service `fake/deployment`: services \"deployment\" not found")
}

func Test_Handler_Destination_Service_Valid(t *testing.T) {
	// Setup
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
	svc := tests.NewMetaObject[*core.Service](t, tests.FakeNamespace, "deployment", func(t *testing.T, obj *core.Service) {
		obj.Spec.Ports = []core.ServicePort{
			{
				Port: 10244,
			},
		}
	})

	refresh := tests.CreateObjects(t, handler.kubeClient, handler.client, &deployment, &extension, &svc)

	// Test
	require.NoError(t, tests.Handle(handler, tests.NewItem(t, operation.Update, extension)))

	// Refresh
	refresh(t)

	// Assert
	require.True(t, extension.Status.Conditions.IsTrue(networkingApi.SpecValidCondition))
	require.True(t, extension.Status.Conditions.IsTrue(networkingApi.DestinationValidCondition))
	require.True(t, extension.Status.Conditions.IsTrue(networkingApi.ReadyCondition))
	require.Equal(t, networkingApi.ArangoRouteStatusTargetServiceType, extension.Status.Target.Type)

	require.Len(t, extension.Status.Target.RenderURLs(), 1)
	require.EqualValues(t, "http://deployment.fake.svc:10244/", extension.Status.Target.RenderURLs()[0])
	require.EqualValues(t, "http1", extension.Status.Target.Protocol)

	c, ok := extension.Status.Conditions.Get(networkingApi.DestinationValidCondition)
	require.True(t, ok)
	require.EqualValues(t, c.Reason, "Destination Found")
	require.EqualValues(t, c.Reason, "Destination Found")
	require.EqualValues(t, c.Hash, extension.Status.Target.Hash())
}

func Test_Handler_Destination_Service_Valid_HTTP1(t *testing.T) {
	// Setup
	handler := newFakeHandler()

	// Arrange
	extension := tests.NewMetaObject[*networkingApi.ArangoRoute](t, tests.FakeNamespace, "test",
		func(t *testing.T, obj *networkingApi.ArangoRoute) {
			obj.Spec.Deployment = util.NewType("deployment")
		},
		func(t *testing.T, obj *networkingApi.ArangoRoute) {
			obj.Spec.Destination = &networkingApi.ArangoRouteSpecDestination{
				Protocol: util.NewType(networkingApi.ArangoRouteDestinationProtocolHTTP1),
				Service: &networkingApi.ArangoRouteSpecDestinationService{
					Object: &sharedApi.Object{
						Name: "deployment",
					},
					Port: util.NewType(intstr.FromInt32(10244)),
				},
			}
		})
	deployment := tests.NewMetaObject[*api.ArangoDeployment](t, tests.FakeNamespace, "deployment")
	svc := tests.NewMetaObject[*core.Service](t, tests.FakeNamespace, "deployment", func(t *testing.T, obj *core.Service) {
		obj.Spec.Ports = []core.ServicePort{
			{
				Port: 10244,
			},
		}
	})

	refresh := tests.CreateObjects(t, handler.kubeClient, handler.client, &deployment, &extension, &svc)

	// Test
	require.NoError(t, tests.Handle(handler, tests.NewItem(t, operation.Update, extension)))

	// Refresh
	refresh(t)

	// Assert
	require.True(t, extension.Status.Conditions.IsTrue(networkingApi.SpecValidCondition))
	require.True(t, extension.Status.Conditions.IsTrue(networkingApi.DestinationValidCondition))
	require.True(t, extension.Status.Conditions.IsTrue(networkingApi.ReadyCondition))
	require.Equal(t, networkingApi.ArangoRouteStatusTargetServiceType, extension.Status.Target.Type)

	require.Len(t, extension.Status.Target.RenderURLs(), 1)
	require.EqualValues(t, "http://deployment.fake.svc:10244/", extension.Status.Target.RenderURLs()[0])
	require.EqualValues(t, "http1", extension.Status.Target.Protocol)

	c, ok := extension.Status.Conditions.Get(networkingApi.DestinationValidCondition)
	require.True(t, ok)
	require.EqualValues(t, c.Reason, "Destination Found")
	require.EqualValues(t, c.Reason, "Destination Found")
	require.EqualValues(t, c.Hash, extension.Status.Target.Hash())
}

func Test_Handler_Destination_Service_Valid_HTTP2(t *testing.T) {
	// Setup
	handler := newFakeHandler()

	// Arrange
	extension := tests.NewMetaObject[*networkingApi.ArangoRoute](t, tests.FakeNamespace, "test",
		func(t *testing.T, obj *networkingApi.ArangoRoute) {
			obj.Spec.Deployment = util.NewType("deployment")
		},
		func(t *testing.T, obj *networkingApi.ArangoRoute) {
			obj.Spec.Destination = &networkingApi.ArangoRouteSpecDestination{
				Protocol: util.NewType(networkingApi.ArangoRouteDestinationProtocolHTTP2),
				Service: &networkingApi.ArangoRouteSpecDestinationService{
					Object: &sharedApi.Object{
						Name: "deployment",
					},
					Port: util.NewType(intstr.FromInt32(10244)),
				},
			}
		})
	deployment := tests.NewMetaObject[*api.ArangoDeployment](t, tests.FakeNamespace, "deployment")
	svc := tests.NewMetaObject[*core.Service](t, tests.FakeNamespace, "deployment", func(t *testing.T, obj *core.Service) {
		obj.Spec.Ports = []core.ServicePort{
			{
				Port: 10244,
			},
		}
	})

	refresh := tests.CreateObjects(t, handler.kubeClient, handler.client, &deployment, &extension, &svc)

	// Test
	require.NoError(t, tests.Handle(handler, tests.NewItem(t, operation.Update, extension)))

	// Refresh
	refresh(t)

	// Assert
	require.True(t, extension.Status.Conditions.IsTrue(networkingApi.SpecValidCondition))
	require.True(t, extension.Status.Conditions.IsTrue(networkingApi.DestinationValidCondition))
	require.True(t, extension.Status.Conditions.IsTrue(networkingApi.ReadyCondition))
	require.Equal(t, networkingApi.ArangoRouteStatusTargetServiceType, extension.Status.Target.Type)

	require.Len(t, extension.Status.Target.RenderURLs(), 1)
	require.EqualValues(t, "http://deployment.fake.svc:10244/", extension.Status.Target.RenderURLs()[0])
	require.EqualValues(t, "http2", extension.Status.Target.Protocol)

	c, ok := extension.Status.Conditions.Get(networkingApi.DestinationValidCondition)
	require.True(t, ok)
	require.EqualValues(t, c.Reason, "Destination Found")
	require.EqualValues(t, c.Reason, "Destination Found")
	require.EqualValues(t, c.Hash, extension.Status.Target.Hash())
}

func Test_Handler_Destination_Service_Valid_WithIP(t *testing.T) {
	// Setup
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
	svc := tests.NewMetaObject[*core.Service](t, tests.FakeNamespace, "deployment", func(t *testing.T, obj *core.Service) {
		obj.Spec.Ports = []core.ServicePort{
			{
				Port: 10244,
			},
		}
		obj.Spec.ClusterIP = "127.0.0.2"
	})

	refresh := tests.CreateObjects(t, handler.kubeClient, handler.client, &deployment, &extension, &svc)

	// Test
	require.NoError(t, tests.Handle(handler, tests.NewItem(t, operation.Update, extension)))

	// Refresh
	refresh(t)

	// Assert
	require.True(t, extension.Status.Conditions.IsTrue(networkingApi.SpecValidCondition))
	require.True(t, extension.Status.Conditions.IsTrue(networkingApi.DestinationValidCondition))
	require.True(t, extension.Status.Conditions.IsTrue(networkingApi.ReadyCondition))
	require.Equal(t, networkingApi.ArangoRouteStatusTargetServiceType, extension.Status.Target.Type)

	require.Len(t, extension.Status.Target.RenderURLs(), 1)
	require.EqualValues(t, "http://127.0.0.2:10244/", extension.Status.Target.RenderURLs()[0])

	c, ok := extension.Status.Conditions.Get(networkingApi.DestinationValidCondition)
	require.True(t, ok)
	require.EqualValues(t, c.Reason, "Destination Found")
	require.EqualValues(t, c.Reason, "Destination Found")
	require.EqualValues(t, c.Hash, extension.Status.Target.Hash())
}

func Test_Handler_Destination_Service_Valid_WithPath(t *testing.T) {
	// Setup
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
				Path: util.NewType("/test/path/"),
			}
		})
	deployment := tests.NewMetaObject[*api.ArangoDeployment](t, tests.FakeNamespace, "deployment")
	svc := tests.NewMetaObject[*core.Service](t, tests.FakeNamespace, "deployment", func(t *testing.T, obj *core.Service) {
		obj.Spec.Ports = []core.ServicePort{
			{
				Port: 10244,
			},
		}
		obj.Spec.ClusterIP = "127.0.0.2"
	})

	refresh := tests.CreateObjects(t, handler.kubeClient, handler.client, &deployment, &extension, &svc)

	// Test
	require.NoError(t, tests.Handle(handler, tests.NewItem(t, operation.Update, extension)))

	// Refresh
	refresh(t)

	// Assert
	require.True(t, extension.Status.Conditions.IsTrue(networkingApi.SpecValidCondition))
	require.True(t, extension.Status.Conditions.IsTrue(networkingApi.DestinationValidCondition))
	require.True(t, extension.Status.Conditions.IsTrue(networkingApi.ReadyCondition))
	require.Equal(t, networkingApi.ArangoRouteStatusTargetServiceType, extension.Status.Target.Type)

	require.Len(t, extension.Status.Target.RenderURLs(), 1)
	require.EqualValues(t, "http://127.0.0.2:10244/test/path/", extension.Status.Target.RenderURLs()[0])

	c, ok := extension.Status.Conditions.Get(networkingApi.DestinationValidCondition)
	require.True(t, ok)
	require.EqualValues(t, c.Reason, "Destination Found")
	require.EqualValues(t, c.Reason, "Destination Found")
	require.EqualValues(t, c.Hash, extension.Status.Target.Hash())
}

func Test_Handler_Destination_Service_ValidName(t *testing.T) {
	// Setup
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
					Port: util.NewType(intstr.FromString("test")),
				},
			}
		})
	deployment := tests.NewMetaObject[*api.ArangoDeployment](t, tests.FakeNamespace, "deployment")
	svc := tests.NewMetaObject[*core.Service](t, tests.FakeNamespace, "deployment", func(t *testing.T, obj *core.Service) {
		obj.Spec.Ports = []core.ServicePort{
			{
				Port: 10241,
				Name: "test1",
			},
			{
				Port: 10244,
				Name: "test",
			},
		}
	})

	refresh := tests.CreateObjects(t, handler.kubeClient, handler.client, &deployment, &extension, &svc)

	// Test
	require.NoError(t, tests.Handle(handler, tests.NewItem(t, operation.Update, extension)))

	// Refresh
	refresh(t)

	// Assert
	require.True(t, extension.Status.Conditions.IsTrue(networkingApi.SpecValidCondition))
	require.True(t, extension.Status.Conditions.IsTrue(networkingApi.DestinationValidCondition))
	require.True(t, extension.Status.Conditions.IsTrue(networkingApi.ReadyCondition))
	require.Equal(t, networkingApi.ArangoRouteStatusTargetServiceType, extension.Status.Target.Type)

	c, ok := extension.Status.Conditions.Get(networkingApi.DestinationValidCondition)
	require.True(t, ok)
	require.EqualValues(t, c.Reason, "Destination Found")
	require.EqualValues(t, c.Reason, "Destination Found")
	require.EqualValues(t, c.Hash, extension.Status.Target.Hash())
}

func Test_Handler_Destination_Service_WrongPort(t *testing.T) {
	// Setup
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
	svc := tests.NewMetaObject[*core.Service](t, tests.FakeNamespace, "deployment", func(t *testing.T, obj *core.Service) {
		obj.Spec.Ports = []core.ServicePort{
			{
				Port: 10245,
			},
		}
	})

	refresh := tests.CreateObjects(t, handler.kubeClient, handler.client, &deployment, &extension, &svc)

	// Test
	require.NoError(t, tests.Handle(handler, tests.NewItem(t, operation.Update, extension)))

	// Refresh
	refresh(t)

	// Assert
	require.True(t, extension.Status.Conditions.IsTrue(networkingApi.SpecValidCondition))
	require.False(t, extension.Status.Conditions.IsTrue(networkingApi.DestinationValidCondition))
	require.False(t, extension.Status.Conditions.IsTrue(networkingApi.ReadyCondition))

	c, ok := extension.Status.Conditions.Get(networkingApi.DestinationValidCondition)
	require.True(t, ok)
	require.EqualValues(t, c.Reason, "Destination Not Found")
	require.EqualValues(t, c.Message, "Port `10244` not defined on Service `fake/deployment`")
}

func Test_Handler_Destination_Service_WrongPortName(t *testing.T) {
	// Setup
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
					Port: util.NewType(intstr.FromString("test")),
				},
			}
		})
	deployment := tests.NewMetaObject[*api.ArangoDeployment](t, tests.FakeNamespace, "deployment")
	svc := tests.NewMetaObject[*core.Service](t, tests.FakeNamespace, "deployment", func(t *testing.T, obj *core.Service) {
		obj.Spec.Ports = []core.ServicePort{
			{
				Port: 10245,
			},
		}
	})

	refresh := tests.CreateObjects(t, handler.kubeClient, handler.client, &deployment, &extension, &svc)

	// Test
	require.NoError(t, tests.Handle(handler, tests.NewItem(t, operation.Update, extension)))

	// Refresh
	refresh(t)

	// Assert
	require.True(t, extension.Status.Conditions.IsTrue(networkingApi.SpecValidCondition))
	require.False(t, extension.Status.Conditions.IsTrue(networkingApi.DestinationValidCondition))
	require.False(t, extension.Status.Conditions.IsTrue(networkingApi.ReadyCondition))

	c, ok := extension.Status.Conditions.Get(networkingApi.DestinationValidCondition)
	require.True(t, ok)
	require.EqualValues(t, c.Reason, "Destination Not Found")
	require.EqualValues(t, c.Message, "Port `test` not defined on Service `fake/deployment`")
}

func Test_Handler_Destination_Service_Insecure_Default(t *testing.T) {
	// Setup
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
	svc := tests.NewMetaObject[*core.Service](t, tests.FakeNamespace, "deployment", func(t *testing.T, obj *core.Service) {
		obj.Spec.Ports = []core.ServicePort{
			{
				Port: 10244,
			},
		}
	})

	refresh := tests.CreateObjects(t, handler.kubeClient, handler.client, &deployment, &extension, &svc)

	// Testcense
	require.NoError(t, tests.Handle(handler, tests.NewItem(t, operation.Update, extension)))

	// Refresh
	refresh(t)

	// Assert
	require.True(t, extension.Status.Conditions.IsTrue(networkingApi.SpecValidCondition))
	require.True(t, extension.Status.Conditions.IsTrue(networkingApi.DestinationValidCondition))
	require.True(t, extension.Status.Conditions.IsTrue(networkingApi.ReadyCondition))
	require.Equal(t, networkingApi.ArangoRouteStatusTargetServiceType, extension.Status.Target.Type)

	c, ok := extension.Status.Conditions.Get(networkingApi.DestinationValidCondition)
	require.True(t, ok)
	require.EqualValues(t, c.Reason, "Destination Found")
	require.EqualValues(t, c.Reason, "Destination Found")
	require.EqualValues(t, c.Hash, extension.Status.Target.Hash())

	require.False(t, extension.Status.Target.TLS.IsInsecure())
}

func Test_Handler_Destination_Service_Insecure_Nil(t *testing.T) {
	// Setup
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
				TLS: &networkingApi.ArangoRouteSpecDestinationTLS{
					Insecure: nil,
				},
			}
		})
	deployment := tests.NewMetaObject[*api.ArangoDeployment](t, tests.FakeNamespace, "deployment")
	svc := tests.NewMetaObject[*core.Service](t, tests.FakeNamespace, "deployment", func(t *testing.T, obj *core.Service) {
		obj.Spec.Ports = []core.ServicePort{
			{
				Port: 10244,
			},
		}
	})

	refresh := tests.CreateObjects(t, handler.kubeClient, handler.client, &deployment, &extension, &svc)

	// Test
	require.NoError(t, tests.Handle(handler, tests.NewItem(t, operation.Update, extension)))

	// Refresh
	refresh(t)

	// Assert
	require.True(t, extension.Status.Conditions.IsTrue(networkingApi.SpecValidCondition))
	require.True(t, extension.Status.Conditions.IsTrue(networkingApi.DestinationValidCondition))
	require.True(t, extension.Status.Conditions.IsTrue(networkingApi.ReadyCondition))
	require.Equal(t, networkingApi.ArangoRouteStatusTargetServiceType, extension.Status.Target.Type)

	c, ok := extension.Status.Conditions.Get(networkingApi.DestinationValidCondition)
	require.True(t, ok)
	require.EqualValues(t, c.Reason, "Destination Found")
	require.EqualValues(t, c.Reason, "Destination Found")
	require.EqualValues(t, c.Hash, extension.Status.Target.Hash())

	require.False(t, extension.Status.Target.TLS.IsInsecure())
}

func Test_Handler_Destination_Service_Insecure_HTTPS_Override(t *testing.T) {
	// Setup
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
				TLS: &networkingApi.ArangoRouteSpecDestinationTLS{
					Insecure: util.NewType(true),
				},
				Schema: util.NewType(networkingApi.ArangoRouteSpecDestinationSchemaHTTPS),
			}
		})
	deployment := tests.NewMetaObject[*api.ArangoDeployment](t, tests.FakeNamespace, "deployment")
	svc := tests.NewMetaObject[*core.Service](t, tests.FakeNamespace, "deployment", func(t *testing.T, obj *core.Service) {
		obj.Spec.Ports = []core.ServicePort{
			{
				Port: 10244,
			},
		}
	})

	refresh := tests.CreateObjects(t, handler.kubeClient, handler.client, &deployment, &extension, &svc)

	// Test
	require.NoError(t, tests.Handle(handler, tests.NewItem(t, operation.Update, extension)))

	// Refresh
	refresh(t)

	// Assert
	require.True(t, extension.Status.Conditions.IsTrue(networkingApi.SpecValidCondition))
	require.True(t, extension.Status.Conditions.IsTrue(networkingApi.DestinationValidCondition))
	require.True(t, extension.Status.Conditions.IsTrue(networkingApi.ReadyCondition))
	require.Equal(t, networkingApi.ArangoRouteStatusTargetServiceType, extension.Status.Target.Type)

	c, ok := extension.Status.Conditions.Get(networkingApi.DestinationValidCondition)
	require.True(t, ok)
	require.EqualValues(t, c.Reason, "Destination Found")
	require.EqualValues(t, c.Reason, "Destination Found")
	require.EqualValues(t, c.Hash, extension.Status.Target.Hash())

	require.True(t, extension.Status.Target.TLS.IsInsecure())
}

func Test_Handler_Destination_Service_Insecure_HTTP_Override(t *testing.T) {
	// Setup
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
				TLS: &networkingApi.ArangoRouteSpecDestinationTLS{
					Insecure: util.NewType(true),
				},
				Schema: util.NewType(networkingApi.ArangoRouteSpecDestinationSchemaHTTP),
			}
		})
	deployment := tests.NewMetaObject[*api.ArangoDeployment](t, tests.FakeNamespace, "deployment")
	svc := tests.NewMetaObject[*core.Service](t, tests.FakeNamespace, "deployment", func(t *testing.T, obj *core.Service) {
		obj.Spec.Ports = []core.ServicePort{
			{
				Port: 10244,
			},
		}
	})

	refresh := tests.CreateObjects(t, handler.kubeClient, handler.client, &deployment, &extension, &svc)

	// Test
	require.NoError(t, tests.Handle(handler, tests.NewItem(t, operation.Update, extension)))

	// Refresh
	refresh(t)

	// Assert
	require.True(t, extension.Status.Conditions.IsTrue(networkingApi.SpecValidCondition))
	require.True(t, extension.Status.Conditions.IsTrue(networkingApi.DestinationValidCondition))
	require.True(t, extension.Status.Conditions.IsTrue(networkingApi.ReadyCondition))
	require.Equal(t, networkingApi.ArangoRouteStatusTargetServiceType, extension.Status.Target.Type)

	c, ok := extension.Status.Conditions.Get(networkingApi.DestinationValidCondition)
	require.True(t, ok)
	require.EqualValues(t, c.Reason, "Destination Found")
	require.EqualValues(t, c.Reason, "Destination Found")
	require.EqualValues(t, c.Hash, extension.Status.Target.Hash())

	require.False(t, extension.Status.Target.TLS.IsInsecure())
}
