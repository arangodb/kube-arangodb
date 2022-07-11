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

package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/arangodb/kube-arangodb/pkg/deployment/resources/inspector"
	inspectorInterface "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/throttle"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
)

const FakeNamespace = "fake"

func NewInspector(t *testing.T, c kclient.Client) inspectorInterface.Inspector {
	i := inspector.NewInspector(throttle.NewAlwaysThrottleComponents(), c, FakeNamespace, FakeNamespace)
	require.NoError(t, i.Refresh(context.Background()))

	return i
}

func NewEmptyInspector(t *testing.T) inspectorInterface.Inspector {
	c := kclient.NewFakeClient()

	return NewInspector(t, c)
}
