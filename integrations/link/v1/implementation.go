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

package v1

import (
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc"

	pbLinkV1 "github.com/arangodb/kube-arangodb/integrations/link/v1/definition"
	pbMetaV1 "github.com/arangodb/kube-arangodb/integrations/meta/v1/definition"
	pbStorageV2 "github.com/arangodb/kube-arangodb/integrations/storage/v2/definition"
	"github.com/arangodb/kube-arangodb/pkg/util/svc"
)

var _ pbLinkV1.LinkV1InternalServer = &implementation{}
var _ pbLinkV1.LinkV1ExternalServer = &implementation{}
var _ svc.Handler = &implementation{}

// New creates a LinkV1 handler serving both internal and external APIs.
// linkID is the UUID of this link type (from configuration).
// handlerID is generated once per runtime instance.
func New(metaClient pbMetaV1.MetaV1Client, storageClient pbStorageV2.StorageV2Client, linkID string) svc.Handler {
	handlerID := uuid.New().String()
	return &implementation{
		store:     newJobStore(metaClient, linkID, handlerID),
		meta:      metaClient,
		storage:   storageClient,
		linkID:    linkID,
		handlerID: handlerID,
	}
}

type implementation struct {
	store     *jobStore
	meta      pbMetaV1.MetaV1Client
	storage   pbStorageV2.StorageV2Client
	linkID    string
	handlerID string

	info *pbLinkV1.LinkInfo

	pbLinkV1.UnimplementedLinkV1InternalServer
	pbLinkV1.UnimplementedLinkV1ExternalServer
}

func (i *implementation) Name() string {
	return Name
}

func (i *implementation) Register(registrar *grpc.Server) {
	pbLinkV1.RegisterLinkV1InternalServer(registrar, i)
	pbLinkV1.RegisterLinkV1ExternalServer(registrar, i)
}

func (i *implementation) Health(ctx context.Context) svc.HealthState {
	if i.info == nil {
		return svc.Unhealthy
	}
	return svc.Healthy
}

// Background starts the handler heartbeat loop. Blocks until ctx is cancelled.
func (i *implementation) Background(ctx context.Context) {
	startHeartbeat(ctx, i.meta, i.linkID, i.handlerID)
}
