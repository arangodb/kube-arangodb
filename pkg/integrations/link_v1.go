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

package integrations

import (
	"context"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pbImplLinkV1 "github.com/arangodb/kube-arangodb/integrations/link/v1"
	pbMetaV1 "github.com/arangodb/kube-arangodb/integrations/meta/v1/definition"
	pbStorageV2 "github.com/arangodb/kube-arangodb/integrations/storage/v2/definition"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/svc"
)

func init() {
	registerer.Register(pbImplLinkV1.Name, func() Integration {
		return &linkV1{}
	})
}

type linkV1 struct {
	linkID          string
	internalAddress string
}

func (b *linkV1) Name() string {
	return pbImplLinkV1.Name
}

func (b *linkV1) Description() string {
	return "LinkV1 Integration"
}

func (b *linkV1) Register(cmd *cobra.Command, fs FlagEnvHandler) error {
	return errors.Errors(
		fs.StringVar(&b.linkID, "connector-id", "", "Link UUID"),
		fs.StringVar(&b.internalAddress, "internal-address", "127.0.0.1:9092", "Internal gRPC address for MetaV1 and StorageV2 clients"),
	)
}

func (b *linkV1) Handler(ctx context.Context, cmd *cobra.Command) (svc.Handler, error) {
	if b.linkID == "" {
		return nil, errors.Errorf("Connector ID is required")
	}

	conn, err := grpc.NewClient(b.internalAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to connect to internal gRPC")
	}

	metaClient := pbMetaV1.NewMetaV1Client(conn)
	storageClient := pbStorageV2.NewStorageV2Client(conn)

	return pbImplLinkV1.New(metaClient, storageClient, b.linkID), nil
}

func (*linkV1) Init(ctx context.Context, cmd *cobra.Command) error {
	return nil
}

// EnabledTypes returns that link is available on both internal and external listeners.
// External listener exposes LinkV1External (AI tool facing REST API).
// Internal listener exposes LinkV1Internal (link process facing gRPC + HTTP API).
func (*linkV1) EnabledTypes() (internal, external bool) {
	return true, true
}
