//
// DISCLAIMER
//
// Copyright 2016-2022 ArangoDB GmbH, Cologne, Germany
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
