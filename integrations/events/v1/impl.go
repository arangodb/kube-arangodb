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
	"io"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	pbEventsV1 "github.com/arangodb/kube-arangodb/integrations/events/v1/definition"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/svc"
)

func New(ctx context.Context, cfg Configuration) (svc.Handler, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	col := cfg.KVCollection(cfg.Endpoint, "_system", "_events")

	col = withTTLIndex(col)

	if _, err := col.Get(ctx); err != nil {
		return nil, err
	}

	return newInternal(cfg, NewArangoRemoteStore[*pbEventsV1.Event](col)), nil
}

func newInternal(cfg Configuration, c RemoteStore[*pbEventsV1.Event]) *implementation {
	q := c

	if cfg.Async.Enabled {
		q = WithAsync(q, cfg.Async.Size, cfg.Async.Retry.Timeout, cfg.Async.Retry.Delay)
	}

	obj := &implementation{
		cfg:    cfg,
		remote: q,
	}

	return obj
}

var _ pbEventsV1.EventsV1Server = &implementation{}
var _ svc.Handler = &implementation{}

type implementation struct {
	pbEventsV1.UnimplementedEventsV1Server

	cfg    Configuration
	remote RemoteStore[*pbEventsV1.Event]
}

func (i *implementation) Name() string {
	return pbEventsV1.Name
}

func (i *implementation) Health() svc.HealthState {
	return svc.Healthy
}

func (i *implementation) Register(registrar *grpc.Server) {
	pbEventsV1.RegisterEventsV1Server(registrar, i)
}

func (i *implementation) Gateway(ctx context.Context, mux *runtime.ServeMux) error {
	return nil
}

func (a *implementation) Background() context.CancelFunc {
	return svc.RunBackground(a.remote)
}

func (i *implementation) Emit(server pbEventsV1.EventsV1_EmitServer) error {
	var events = make([]*pbEventsV1.Event, 0, MaxEventCount)

	start := time.Now().Truncate(time.Second)

	for {
		msg, err := server.Recv()

		if errors.IsGRPCCode(err, codes.Canceled) {
			return io.ErrUnexpectedEOF
		}

		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return err
		}

		if len(events)+len(msg.GetEvents()) > MaxEventCount {
			return status.Error(codes.ResourceExhausted, "exceeded limit of the events per request")
		}

		for _, ev := range msg.GetEvents() {
			// Trim to seconds to keep cross-platform compatibility
			if q := ev.Created; q == nil {
				ev.Created = timestamppb.New(start)
			} else {
				ev.Created = timestamppb.New(q.AsTime().Truncate(time.Second))
			}

			events = append(events, ev)
		}
	}

	if len(events) == 0 {
		return server.SendAndClose(&pbEventsV1.EventsV1Response{
			Created:   timestamppb.New(start),
			Processed: 0,
			Completed: true,
		})
	}

	if err := i.remote.Emit(server.Context(), events...); err != nil {
		logger.Err(err).Int("events", len(events)).Warn("Failed to emit events")
		return status.Error(codes.Internal, "Unable to emit events")
	}

	logger.Int("events", len(events)).Info("Emitted events")

	return server.SendAndClose(&pbEventsV1.EventsV1Response{
		Created:   timestamppb.New(start),
		Processed: int32(len(events)),
		Completed: !i.cfg.Async.Enabled,
	})
}
