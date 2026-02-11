//
// DISCLAIMER
//
// Copyright 2025-2026 ArangoDB GmbH, Cologne, Germany
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

func New(cfg Configuration) (svc.Handler, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	col := cfg.WithDatabase(cfg.Endpoint).
		CreateCollection("_meta_store", cfg.SourceCollectionProps()).
		WithTTLIndex("system_meta_store_object_ttl", 0, "ttl").
		Get()

	return newInternal(cfg, cache.NewRemoteCacheWithTTL[*Object](col, cfg.TTL)), nil
}

func newInternal(cfg Configuration, c cache.RemoteCache[*Object]) *implementation {
	return &implementation{
		cfg:   cfg,
		cache: c,
	}
}

var _ pbMetaV1.MetaV1Server = &implementation{}
var _ svc.Handler = &implementation{}

type implementation struct {
	pbMetaV1.UnimplementedMetaV1Server

	lock sync.RWMutex

	cfg Configuration

	cache cache.RemoteCache[*Object]
}

func (i *implementation) Name() string {
	return pbMetaV1.Name
}

func (i *implementation) Health(ctx context.Context) svc.HealthState {
	return svc.Healthy
}

func (i *implementation) Register(registrar *grpc.Server) {
	pbMetaV1.RegisterMetaV1Server(registrar, i)
}

func (i *implementation) Background(ctx context.Context) {
	i.init(ctx)
}

func (i *implementation) init(ctx context.Context) {
	time.Sleep(time.Second)

	timerT := time.NewTicker(time.Second)
	defer timerT.Stop()

	for {
		err := i.cache.Init(ctx)
		if err == nil {
			return
		}

		logger.Err(err).Warn("Unable to init collection")

		select {
		case <-timerT.C:
			continue
		case <-ctx.Done():
			return
		}
	}
}

func (i *implementation) Gateway(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
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

	obj := object.AsResponse()

	nobj, err := req.GetSecret().Decrypt(obj.GetObject())
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "Decryption failed: %v", err)
	}

	obj.Object = nobj

	return obj, nil
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

	if obj.Meta.Expires != nil {
		obj.TTL = obj.Meta.Expires
	}

	obj.Object.Object = req.GetObject()

	nobj, err := req.GetSecret().Encrypt(obj.Object.Object)
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "Encryption failed: %v", err)
	}

	obj.Object.Object = nobj

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
