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
	"sync"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arangodb/go-driver/v2/arangodb/shared"

	pbMetaV1 "github.com/arangodb/kube-arangodb/integrations/meta/v1/definition"
	pbSharedV1 "github.com/arangodb/kube-arangodb/integrations/shared/v1/definition"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/cache"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/svc"
)

func New(ctx context.Context, cfg Configuration) (svc.Handler, error) {
	return newInternal(ctx, cfg)
}

func newInternal(ctx context.Context, cfg Configuration) (*implementation, error) {
	return newInternalWithRemoteCache(ctx, cfg, cache.NewRemoteCacheWithTTL[*Object](cfg.KVCollection(cfg.Endpoint, "_system", "_meta_store"), cfg.TTL))
}

func newInternalWithRemoteCache(ctx context.Context, cfg Configuration, c cache.RemoteCache[*Object]) (*implementation, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	obj := &implementation{
		cfg:   cfg,
		ctx:   ctx,
		cache: c,
	}

	return obj, nil
}

var _ pbMetaV1.MetaV1Server = &implementation{}
var _ svc.Handler = &implementation{}

type implementation struct {
	pbMetaV1.UnimplementedMetaV1Server

	lock sync.RWMutex

	ctx context.Context
	cfg Configuration

	cache cache.RemoteCache[*Object]
}

func (i *implementation) Name() string {
	return pbMetaV1.Name
}

func (i *implementation) Health() svc.HealthState {
	return svc.Healthy
}

func (i *implementation) Register(registrar *grpc.Server) {
	pbMetaV1.RegisterMetaV1Server(registrar, i)
}

func (i *implementation) Gateway(ctx context.Context, mux *runtime.ServeMux) error {
	return nil
}

func (i *implementation) Get(ctx context.Context, req *pbMetaV1.ObjectRequest) (*pbMetaV1.ObjectResponse, error) {
	i.lock.RLock()
	defer i.lock.RUnlock()

	key := i.cfg.Key(req.GetKey())
	object, exists, err := i.cache.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, status.Errorf(codes.NotFound, "Key %s not found", key)
	}

	return object.AsResponse(), nil

}

func (i *implementation) Set(ctx context.Context, req *pbMetaV1.SetRequest) (*pbMetaV1.ObjectResponse, error) {
	i.lock.Lock()
	defer i.lock.Unlock()

	key := i.cfg.Key(req.GetKey())

	var objMeta ObjectMeta

	objMeta.Updated = meta.Now()

	if v := req.GetTtl(); v != nil {
		objMeta.Expires = util.NewType(meta.NewTime(time.Now().Add(v.AsDuration())))
	}

	var obj Object

	obj.Meta = &objMeta
	obj.Key = key
	obj.Rev = req.Revision

	obj.Object.Object = req.GetObject()

	if err := i.cache.Put(ctx, key, &obj); err != nil {
		if shared.IsPreconditionFailed(err) {
			logger.Err(err).Str("key", key).Warn("Precondition failed")
			return nil, status.Errorf(codes.FailedPrecondition, "Key %s cannot be updated with revision %s", key, req.GetRevision())
		}

		return nil, err
	}

	nObj, exists, err := i.cache.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, status.Errorf(codes.NotFound, "Key %s not found", key)
	}

	return nObj.AsResponse(), nil
}

func (i *implementation) Delete(ctx context.Context, req *pbMetaV1.ObjectRequest) (*pbSharedV1.Empty, error) {
	i.lock.Lock()
	defer i.lock.Unlock()

	key := i.cfg.Key(req.GetKey())

	removed, err := i.cache.Remove(ctx, key)
	if err != nil {
		return nil, err
	}

	if !removed {
		return nil, status.Errorf(codes.NotFound, "Key %s not found", key)
	}

	return &pbSharedV1.Empty{}, nil
}

func (i *implementation) List(req *pbMetaV1.ListRequest, server pbMetaV1.MetaV1_ListServer) error {
	log := logger.Str("func", "List")

	size := int(util.OptionalType(req.Batch, 128))

	if size <= 0 {
		return status.Errorf(codes.InvalidArgument, "batch cannot be smaller than 0")
	}

	resp, err := i.cache.List(server.Context(), size, util.OptionalType(req.Prefix, ""))
	if err != nil {
		log.Err(err).Debug("Failed to list objects")
		return err
	}

	for {
		keys, err := resp.Next(server.Context())
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}

		if err := server.Send(&pbMetaV1.ListResponseChunk{Keys: keys}); err != nil {
			log.Err(err).Debug("Failed to send ListResponseChunk")
			return err
		}
	}
}
