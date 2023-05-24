//
// DISCLAIMER
//
// Copyright 2016-2023 ArangoDB GmbH, Cologne, Germany
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

package patcher

import (
	"testing"

	"github.com/stretchr/testify/require"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/arangodb/kube-arangodb/pkg/util"
)

func Test_Service_Ports(t *testing.T) {
	t.Run("Equal", func(t *testing.T) {
		q := PatchServicePorts([]core.ServicePort{
			{
				Name: "test",
			},
		})(&core.Service{
			Spec: core.ServiceSpec{
				Ports: []core.ServicePort{
					{
						Name: "test",
					},
				},
			},
		})

		require.Len(t, q, 0)
	})

	t.Run("Missing", func(t *testing.T) {
		q := PatchServicePorts([]core.ServicePort{
			{
				Name: "test",
			},
			{
				Name: "exporter",
			},
		})(&core.Service{
			Spec: core.ServiceSpec{
				Ports: []core.ServicePort{
					{
						Name: "test",
					},
				},
			},
		})

		require.Len(t, q, 1)
	})
}

func Test_Service_OnlyPorts(t *testing.T) {
	t.Run("Equal", func(t *testing.T) {
		q := PatchServiceOnlyPorts(core.ServicePort{
			Name: "test",
		})(&core.Service{
			Spec: core.ServiceSpec{
				Ports: []core.ServicePort{
					{
						Name: "test",
					},
				},
			},
		})

		require.Len(t, q, 0)
	})

	t.Run("Missing", func(t *testing.T) {
		q := PatchServiceOnlyPorts(core.ServicePort{
			Name: "test",
		})(&core.Service{
			Spec: core.ServiceSpec{
				Ports: []core.ServicePort{
					{
						Name: "test2",
					},
				},
			},
		})

		require.Len(t, q, 1)
	})

	t.Run("Different", func(t *testing.T) {
		q := PatchServiceOnlyPorts(core.ServicePort{
			Name: "test",
			Port: 8529,
		})(&core.Service{
			Spec: core.ServiceSpec{
				Ports: []core.ServicePort{
					{
						Name: "test1",
					},
				},
			},
		})

		require.Len(t, q, 1)
	})

	t.Run("Different Port", func(t *testing.T) {
		q := PatchServiceOnlyPorts(core.ServicePort{
			Name: "test",
			Port: 8529,
		})(&core.Service{
			Spec: core.ServiceSpec{
				Ports: []core.ServicePort{
					{
						Name: "test",
					},
				},
			},
		})

		require.Len(t, q, 1)
	})

	t.Run("Changed NodePort", func(t *testing.T) {
		q := PatchServiceOnlyPorts(core.ServicePort{
			Name: "test",
			Port: 8529,
		})(&core.Service{
			Spec: core.ServiceSpec{
				Ports: []core.ServicePort{
					{
						Name:     "test",
						Port:     8529,
						NodePort: 12345,
					},
				},
			},
		})

		require.Len(t, q, 0)
	})

	t.Run("Changed Port", func(t *testing.T) {
		q := PatchServiceOnlyPorts(core.ServicePort{
			Name: "test",
			Port: 8528,
		})(&core.Service{
			Spec: core.ServiceSpec{
				Ports: []core.ServicePort{
					{
						Name:     "test",
						Port:     8529,
						NodePort: 12345,
					},
				},
			},
		})

		require.Len(t, q, 1)
	})

	t.Run("Ignore fields", func(t *testing.T) {
		q := PatchServiceOnlyPorts(core.ServicePort{
			Name: "test",
			Port: 8528,
		})(&core.Service{
			Spec: core.ServiceSpec{
				Ports: []core.ServicePort{
					{
						Name:        "test",
						Protocol:    core.ProtocolTCP,
						AppProtocol: util.NewType[string]("test"),
						Port:        8528,
						TargetPort: intstr.IntOrString{
							StrVal: "TEST",
							IntVal: 0,
						},
						NodePort: 6543,
					},
				},
			},
		})

		require.Len(t, q, 0)
	})
}
