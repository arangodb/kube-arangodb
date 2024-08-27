//
// DISCLAIMER
//
// Copyright 2023-2024 ArangoDB GmbH, Cologne, Germany
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
	"time"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"

	pbPongV1 "github.com/arangodb/kube-arangodb/integrations/pong/v1/definition"
	pbSharedV1 "github.com/arangodb/kube-arangodb/integrations/shared/v1/definition"
	"github.com/arangodb/kube-arangodb/pkg/util/svc"
)

func New() svc.Handler {
	return &impl{}
}

var _ pbPongV1.PongV1Server = &impl{}
var _ svc.Handler = &impl{}

type impl struct {
	pbPongV1.UnimplementedPongV1Server
}

func (i *impl) Name() string {
	return pbPongV1.Name
}

func (i *impl) Health() svc.HealthState {
	return svc.Healthy
}

func (i *impl) Register(registrar *grpc.Server) {
	pbPongV1.RegisterPongV1Server(registrar, i)
}
func (i *impl) Ping(context.Context, *pbSharedV1.Empty) (*pbPongV1.PongV1PingResponse, error) {
	return &pbPongV1.PongV1PingResponse{Time: timestamppb.New(time.Now().UTC())}, nil
}
