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

package v1

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"

	pbEventsV1 "github.com/arangodb/kube-arangodb/integrations/events/v1/definition"
	"github.com/arangodb/kube-arangodb/pkg/util/shutdown"
)

func Test_EventsStream(t *testing.T) {
	ctx, c := context.WithCancel(shutdown.Context())
	defer c()

	client, cache := Client(t, ctx)

	em, err := client.Emit(ctx)
	require.NoError(t, err)

	require.NoError(t, em.Send(&pbEventsV1.EventsV1Request{Events: []*pbEventsV1.Event{{
		Type: "TYPE",
	}}}))

	require.NoError(t, em.Send(&pbEventsV1.EventsV1Request{Events: []*pbEventsV1.Event{{
		Type: "TYPE1",
	}}}))

	require.NoError(t, em.Send(&pbEventsV1.EventsV1Request{Events: []*pbEventsV1.Event{{
		Type: "TYPE2",
	}}}))

	resp, err := em.CloseAndRecv()
	require.NoError(t, err)

	require.EqualValues(t, resp.Processed, 3)

	ret := cache.Events(t)

	require.Len(t, ret, 3)
	require.Equal(t, "TYPE", ret[0].Type)
	require.Equal(t, "TYPE1", ret[1].Type)
	require.Equal(t, "TYPE2", ret[2].Type)
}

func Test_EventsStream_Exceed(t *testing.T) {
	ctx, c := context.WithCancel(shutdown.Context())
	defer c()

	client, cache := Client(t, ctx)

	em, err := client.Emit(ctx)
	require.NoError(t, err)

	all := make([]*pbEventsV1.Event, MaxEventCount)

	for id := range all {
		all[id] = &pbEventsV1.Event{
			Type: fmt.Sprintf("TYPE%d", id),
			Dimensions: map[string]string{
				"ID": fmt.Sprintf("%d", id),
			},
			ServiceId: fmt.Sprintf("TYPE%d", id),
			Body: map[string]float32{
				"ID": float32(id),
			},
		}
	}

	require.NoError(t, em.Send(&pbEventsV1.EventsV1Request{Events: all}))

	require.NoError(t, em.Send(&pbEventsV1.EventsV1Request{Events: []*pbEventsV1.Event{all[0]}}))

	_, err = em.CloseAndRecv()
	require.Error(t, err)

	require.Len(t, cache.Events(t), 0)
}

func Test_EventsStream_Time(t *testing.T) {
	ctx, c := context.WithCancel(shutdown.Context())
	defer c()

	start := time.Now()

	client, cache := Client(t, ctx)

	em, err := client.Emit(ctx)
	require.NoError(t, err)

	require.NoError(t, em.Send(&pbEventsV1.EventsV1Request{Events: []*pbEventsV1.Event{{
		Type: "TYPE",
	}}}))

	require.NoError(t, em.Send(&pbEventsV1.EventsV1Request{Events: []*pbEventsV1.Event{{
		Type:    "TYPE1",
		Created: timestamppb.New(start.Add(-time.Hour)),
	}}}))

	time.Sleep(2 * time.Second)

	require.NoError(t, em.Send(&pbEventsV1.EventsV1Request{Events: []*pbEventsV1.Event{{
		Type: "TYPE2",
	}}}))

	resp, err := em.CloseAndRecv()
	require.NoError(t, err)

	require.EqualValues(t, resp.Processed, 3)

	ret := cache.Events(t)

	require.Len(t, ret, 3)
	require.Equal(t, "TYPE", ret[0].Type)
	require.Equal(t, resp.Created.AsTime().Unix(), ret[0].GetCreated().AsTime().Unix())
	require.Equal(t, "TYPE1", ret[1].Type)
	require.Equal(t, start.Add(-time.Hour).Unix(), ret[1].GetCreated().AsTime().Unix())
	require.Equal(t, "TYPE2", ret[2].Type)
	require.Equal(t, resp.Created.AsTime().Unix(), ret[2].GetCreated().AsTime().Unix())

	require.Equal(t, ret[2].GetCreated().AsTime().Unix(), ret[0].GetCreated().AsTime().Unix())
}
