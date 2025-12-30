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

package impl

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/arangodb/kube-arangodb/integrations/shared/v1/definition"
	"github.com/arangodb/kube-arangodb/pkg/api/server"
)

func (i *implementation) OperatorLiveness(ctx context.Context, empty *definition.Empty) (*definition.Empty, error) {
	if v := i.cfg.LivenessProbe; v != nil {
		if !v.IsAlive() {
			return nil, status.Error(codes.Unavailable, "NotReady")
		}
	}

	return &definition.Empty{}, nil
}

func (i *implementation) OperatorReadiness(ctx context.Context, empty *definition.Empty) (*definition.Empty, error) {
	for _, v := range i.cfg.ReadinessProbes {
		if !v.IsReady() {
			return nil, status.Error(codes.Unavailable, "NotReady")
		}
	}

	return &definition.Empty{}, nil
}

func (i *implementation) OperatorServiceReadiness(ctx context.Context, health *server.OperatorService) (*definition.Empty, error) {
	if v, ok := i.cfg.ReadinessProbes[health.GetName()]; !ok {
		return nil, status.Error(codes.NotFound, "Not Found")
	} else if !v.IsReady() {
		return nil, status.Error(codes.Unavailable, "NotReady")
	}

	return &definition.Empty{}, nil
}
