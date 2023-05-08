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

package resources

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_MetricsInc_Container(t *testing.T) {
	var m Metrics

	m.IncMemberContainerRestarts("ID", "server", "OOMKill", 137)

	require.EqualValues(t, 1, m.Members["ID"].ContainerRestarts["server"][137]["OOMKill"])

	m.IncMemberContainerRestarts("ID", "server", "OOMKill", 137)
	require.EqualValues(t, 2, m.Members["ID"].ContainerRestarts["server"][137]["OOMKill"])

	m.IncMemberContainerRestarts("ID", "server2", "OOMKill", 137)
	require.EqualValues(t, 2, m.Members["ID"].ContainerRestarts["server"][137]["OOMKill"])
	require.EqualValues(t, 1, m.Members["ID"].ContainerRestarts["server2"][137]["OOMKill"])

	m.IncMemberContainerRestarts("ID2", "server", "OOMKill", 137)
	m.IncMemberContainerRestarts("ID", "server", "OOMKill", 138)
	require.EqualValues(t, 2, m.Members["ID"].ContainerRestarts["server"][137]["OOMKill"])
	require.EqualValues(t, 1, m.Members["ID"].ContainerRestarts["server2"][137]["OOMKill"])
	require.EqualValues(t, 1, m.Members["ID2"].ContainerRestarts["server"][137]["OOMKill"])
	require.EqualValues(t, 1, m.Members["ID"].ContainerRestarts["server"][138]["OOMKill"])
}

func Test_MetricsInc_InitContainer(t *testing.T) {
	var m Metrics

	m.IncMemberInitContainerRestarts("ID", "server", "OOMKill", 137)

	require.EqualValues(t, 1, m.Members["ID"].InitContainerRestarts["server"][137]["OOMKill"])

	m.IncMemberInitContainerRestarts("ID", "server", "OOMKill", 137)
	require.EqualValues(t, 2, m.Members["ID"].InitContainerRestarts["server"][137]["OOMKill"])

	m.IncMemberInitContainerRestarts("ID", "server2", "OOMKill", 137)
	require.EqualValues(t, 2, m.Members["ID"].InitContainerRestarts["server"][137]["OOMKill"])
	require.EqualValues(t, 1, m.Members["ID"].InitContainerRestarts["server2"][137]["OOMKill"])

	m.IncMemberInitContainerRestarts("ID2", "server", "OOMKill", 137)
	m.IncMemberInitContainerRestarts("ID", "server", "OOMKill", 138)
	require.EqualValues(t, 2, m.Members["ID"].InitContainerRestarts["server"][137]["OOMKill"])
	require.EqualValues(t, 1, m.Members["ID"].InitContainerRestarts["server2"][137]["OOMKill"])
	require.EqualValues(t, 1, m.Members["ID2"].InitContainerRestarts["server"][137]["OOMKill"])
	require.EqualValues(t, 1, m.Members["ID"].InitContainerRestarts["server"][138]["OOMKill"])
}

func Test_MetricsInc_EphemeralContainer(t *testing.T) {
	var m Metrics

	m.IncMemberEphemeralContainerRestarts("ID", "server", "OOMKill", 137)

	require.EqualValues(t, 1, m.Members["ID"].EphemeralContainerRestarts["server"][137]["OOMKill"])

	m.IncMemberEphemeralContainerRestarts("ID", "server", "OOMKill", 137)
	require.EqualValues(t, 2, m.Members["ID"].EphemeralContainerRestarts["server"][137]["OOMKill"])

	m.IncMemberEphemeralContainerRestarts("ID", "server2", "OOMKill", 137)
	require.EqualValues(t, 2, m.Members["ID"].EphemeralContainerRestarts["server"][137]["OOMKill"])
	require.EqualValues(t, 1, m.Members["ID"].EphemeralContainerRestarts["server2"][137]["OOMKill"])

	m.IncMemberEphemeralContainerRestarts("ID2", "server", "OOMKill", 137)
	m.IncMemberEphemeralContainerRestarts("ID", "server", "OOMKill", 138)
	require.EqualValues(t, 2, m.Members["ID"].EphemeralContainerRestarts["server"][137]["OOMKill"])
	require.EqualValues(t, 1, m.Members["ID"].EphemeralContainerRestarts["server2"][137]["OOMKill"])
	require.EqualValues(t, 1, m.Members["ID2"].EphemeralContainerRestarts["server"][137]["OOMKill"])
	require.EqualValues(t, 1, m.Members["ID"].EphemeralContainerRestarts["server"][138]["OOMKill"])
}
