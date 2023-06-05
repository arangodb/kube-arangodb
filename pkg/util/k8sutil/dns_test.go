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

package k8sutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arangodb/kube-arangodb/pkg/util"
)

// TestCreatePodDNSName tests CreatePodDNSName.
func TestCreatePodDNSName(t *testing.T) {
	depl := &meta.ObjectMeta{
		Name:      "test",
		Namespace: "ns",
	}
	n := CreatePodDNSName(depl, "agent", "id1")
	assert.Equal(t, "test-agent-id1.test-int.ns.svc", n)
}

// TestCreateDatabaseClientServiceDNSName tests CreateDatabaseClientServiceDNSName.
func TestCreateDatabaseClientServiceDNSName(t *testing.T) {
	depl := &meta.ObjectMeta{
		Name:      "test",
		Namespace: "ns",
	}
	n := CreateDatabaseClientServiceDNSName(depl)
	assert.Equal(t, "test.ns.svc", n)
}

func TestCreatePodDNSNameWithDomain(t *testing.T) {
	depl := &meta.ObjectMeta{
		Name:      "test",
		Namespace: "ns",
	}

	assert.Equal(t, "test-agent-id1.test-int.ns.svc", CreatePodDNSNameWithDomain(depl, nil, "agent", "id1"))
	assert.Equal(t, "test-agent-id1.test-int.ns.svc.cluster.local", CreatePodDNSNameWithDomain(depl, util.NewType[string]("cluster.local"), "agent", "id1"))
}
