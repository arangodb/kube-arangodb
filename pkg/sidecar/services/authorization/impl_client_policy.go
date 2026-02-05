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

package authorization

import (
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pbSharedV1 "github.com/arangodb/kube-arangodb/integrations/shared/v1/definition"
	sidecarSvcAuthz "github.com/arangodb/kube-arangodb/pkg/sidecar/services/authorization/definition"
	"github.com/arangodb/kube-arangodb/pkg/sidecar/services/authorization/pool"
	sidecarSvcAuthzTypes "github.com/arangodb/kube-arangodb/pkg/sidecar/services/authorization/types"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/svc/authenticator"
)

func (a *implementation) PoolPolicyChanges(request *sidecarSvcAuthz.AuthorizationPoolRequest, g grpc.ServerStreamingServer[sidecarSvcAuthz.AuthorizationPoolPolicyResponse]) error {
	if authenticator.GetIdentity(g.Context()) == nil {
		return status.Error(codes.Unauthenticated, "Unauthenticated")
	}

	index := request.GetStart()

	tickerT := time.NewTicker(250 * time.Millisecond)
	defer tickerT.Stop()

	last := time.Now()

	for {
		select {
		case <-tickerT.C:
			// Process
			items, err := a.policies.Pool(index)
			if err != nil {
				if pool.IsPoolOutOfBoundsError(err) {
					return status.Error(codes.OutOfRange, "out of bounds")
				}

				return status.Error(codes.Internal, err.Error())
			}

			if len(items) == 0 {
				if time.Since(last) > request.GetTimeout().AsDuration() {
					// Send empty response
					if err := g.Send(&sidecarSvcAuthz.AuthorizationPoolPolicyResponse{
						Items: nil,
					}); err != nil {
						return status.Error(codes.Internal, err.Error())
					}
					last = time.Now()
				}

				continue
			}

			for _, item := range util.BatchList(128, items) {
				if err := g.Send(&sidecarSvcAuthz.AuthorizationPoolPolicyResponse{
					Items: util.FormatList(item, func(a pool.OffsetItem[*sidecarSvcAuthzTypes.Policy]) *sidecarSvcAuthz.AuthorizationPoolPolicyResponseItem {
						return &sidecarSvcAuthz.AuthorizationPoolPolicyResponseItem{
							Name:  a.Name,
							Index: a.Sequence,
							Item:  a.Item,
						}
					}),
				}); err != nil {
					return status.Error(codes.Internal, err.Error())
				}
			}

			last = time.Now()
			index = items[len(items)-1].Sequence

		case <-g.Context().Done():
			return g.Context().Err()
		}
	}
}

func (a *implementation) GetPolicy(empty *pbSharedV1.Empty, g grpc.ServerStreamingServer[sidecarSvcAuthz.AuthorizationPoolPolicyResponse]) error {
	if authenticator.GetIdentity(g.Context()) == nil {
		return status.Error(codes.Unauthenticated, "Unauthenticated")
	}

	items := a.policies.Get()

	for _, item := range util.BatchList(128, items) {
		if err := g.Send(&sidecarSvcAuthz.AuthorizationPoolPolicyResponse{
			Items: util.FormatList(item, func(a pool.OffsetItem[*sidecarSvcAuthzTypes.Policy]) *sidecarSvcAuthz.AuthorizationPoolPolicyResponseItem {
				return &sidecarSvcAuthz.AuthorizationPoolPolicyResponseItem{
					Name:  a.Name,
					Index: a.Sequence,
					Item:  a.Item,
				}
			}),
		}); err != nil {
			return err
		}
	}

	return nil
}
