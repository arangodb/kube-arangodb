//
// DISCLAIMER
//
// Copyright 2026 ArangoDB GmbH, Cologne, Germany
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

package shared

import (
	"context"
	"fmt"
	"testing"

	pbEnvoyAuthV3 "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
	"github.com/stretchr/testify/require"
)

// recordingHandler records the order in which it is invoked, may mutate the response and may fail.
type recordingHandler struct {
	id     string
	called *[]string
	err    error
	mutate func(*Response)
}

func (r recordingHandler) Handle(ctx context.Context, request *pbEnvoyAuthV3.CheckRequest, current *Response) error {
	*r.called = append(*r.called, r.id)
	if r.mutate != nil {
		r.mutate(current)
	}
	return r.err
}

func gen(h AuthHandler, ok bool, err error) FactoryGen {
	return func(ctx context.Context, configuration Configuration) (AuthHandler, bool, error) {
		return h, ok, err
	}
}

func Test_Factory_Render(t *testing.T) {
	ctx := context.Background()

	t.Run("Propagates factory error", func(t *testing.T) {
		f := NewFactory(gen(nil, false, fmt.Errorf("boom")))
		_, err := f.Render(ctx, Configuration{})
		require.EqualError(t, err, "boom")
	})

	t.Run("Skips handlers reporting not-ok", func(t *testing.T) {
		var order []string
		f := NewFactory(
			gen(recordingHandler{id: "a", called: &order}, true, nil),
			gen(nil, false, nil),
			gen(recordingHandler{id: "b", called: &order}, true, nil),
		)

		h, err := f.Render(ctx, Configuration{})
		require.NoError(t, err)

		require.NoError(t, h.Handle(ctx, &pbEnvoyAuthV3.CheckRequest{}, &Response{}))
		require.Equal(t, []string{"a", "b"}, order)
	})
}

func Test_Handlers_Handle(t *testing.T) {
	ctx := context.Background()

	t.Run("Invokes handlers in order and applies mutations", func(t *testing.T) {
		var order []string
		f := NewFactory(
			gen(recordingHandler{id: "a", called: &order, mutate: func(r *Response) {
				r.User = &ResponseAuth{User: "root"}
			}}, true, nil),
			gen(recordingHandler{id: "b", called: &order}, true, nil),
		)

		h, err := f.Render(ctx, Configuration{})
		require.NoError(t, err)

		var current Response
		require.NoError(t, h.Handle(ctx, &pbEnvoyAuthV3.CheckRequest{}, &current))
		require.Equal(t, []string{"a", "b"}, order)
		require.True(t, current.Authenticated())
	})

	t.Run("Stops on first error", func(t *testing.T) {
		var order []string
		f := NewFactory(
			gen(recordingHandler{id: "a", called: &order, err: fmt.Errorf("denied")}, true, nil),
			gen(recordingHandler{id: "b", called: &order}, true, nil),
		)

		h, err := f.Render(ctx, Configuration{})
		require.NoError(t, err)

		require.EqualError(t, h.Handle(ctx, &pbEnvoyAuthV3.CheckRequest{}, &Response{}), "denied")
		require.Equal(t, []string{"a"}, order)
	})
}
