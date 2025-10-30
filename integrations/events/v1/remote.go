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
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/grpc"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"

	"github.com/arangodb/go-driver/v2/arangodb"

	"github.com/arangodb/kube-arangodb/pkg/util/cache"
)

type RemoteStore[IN proto.Message] interface {
	Init(ctx context.Context) error

	Emit(ctx context.Context, events ...IN) error
}

func NewArangoRemoteStore[IN proto.Message](client cache.Object[arangodb.Collection]) RemoteStore[IN] {
	return &arangoRemoteStore[IN]{
		client: client,
	}
}

type arangoRemoteStore[IN proto.Message] struct {
	client cache.Object[arangodb.Collection]
}

func (a *arangoRemoteStore[IN]) Init(ctx context.Context) error {
	return a.client.Init(ctx)
}

func (a *arangoRemoteStore[IN]) Emit(ctx context.Context, events ...IN) error {
	if len(events) == 0 {
		return nil
	}

	col, err := a.client.Get(ctx)
	if err != nil {
		return err
	}

	_, err = col.CreateDocuments(ctx, util.FormatList(events, func(a IN) grpc.Object[IN] {
		return grpc.NewObject(a)
	}))
	if err != nil {
		return errors.Wrapf(err, "Unable to save events")
	}
	return err
}
