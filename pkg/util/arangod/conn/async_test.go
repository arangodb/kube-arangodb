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

package conn

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/arangodb/go-driver"

	"github.com/arangodb/kube-arangodb/pkg/util/tests"
)

func Test_Async(t *testing.T) {
	s := tests.NewServer(t)

	c := s.NewConnection()

	c = NewAsyncConnection(c)

	client, err := driver.NewClient(driver.ClientConfig{
		Connection: c,
	})
	require.NoError(t, err)

	a := tests.NewAsyncHandler(t, s, http.MethodGet, "/_api/version", http.StatusOK, driver.VersionInfo{
		Server:  "foo",
		Version: "",
		License: "",
		Details: nil,
	})

	a.Start()

	_, err = client.Version(context.Background())
	require.Error(t, err)
	id, ok := IsAsyncJobInProgress(err)
	require.True(t, ok)
	require.Equal(t, a.ID(), id)

	a.InProgress()

	ctx := WithAsyncID(context.TODO(), a.ID())

	_, err = client.Version(ctx)
	require.Error(t, err)
	id, ok = IsAsyncJobInProgress(err)
	require.True(t, ok)
	require.Equal(t, a.ID(), id)

	a.Done()

	v, err := client.Version(ctx)
	require.NoError(t, err)

	require.Equal(t, v.Server, "foo")

	defer s.Stop()
}
