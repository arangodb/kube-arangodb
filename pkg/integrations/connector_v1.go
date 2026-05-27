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

	pbImplConnectorV1 "github.com/arangodb/kube-arangodb/integrations/connector/v1"
	pbMetaV1 "github.com/arangodb/kube-arangodb/integrations/meta/v1/definition"
	pbStorageV2 "github.com/arangodb/kube-arangodb/integrations/storage/v2/definition"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/svc"
)

func init() {
	registerer.Register(pbImplConnectorV1.Name, func() Integration {
		return &connectorV1{}
	})
}

type connectorV1 struct {
	connectorID     string
	internalAddress string
}

func (b *connectorV1) Name() string {
	return pbImplConnectorV1.Name
}

func (b *connectorV1) Description() string {
	return "ConnectorV1 Integration"
}

func (b *connectorV1) Register(cmd *cobra.Command, fs FlagEnvHandler) error {
	return errors.Errors(
		fs.StringVar(&b.connectorID, "connector-id", "", "Connector UUID"),
		fs.StringVar(&b.internalAddress, "internal-address", "127.0.0.1:9092", "Internal gRPC address for MetaV1 and StorageV2 clients"),
	)
}

func (b *connectorV1) Handler(ctx context.Context, cmd *cobra.Command) (svc.Handler, error) {
	if b.connectorID == "" {
		return nil, errors.Errorf("Connector ID is required")
	}

	conn, err := grpc.NewClient(b.internalAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to connect to internal gRPC")
	}

	metaClient := pbMetaV1.NewMetaV1Client(conn)
	storageClient := pbStorageV2.NewStorageV2Client(conn)

	return pbImplConnectorV1.New(metaClient, storageClient, b.connectorID), nil
}

func (*connectorV1) Init(ctx context.Context, cmd *cobra.Command) error {
	return nil
}

// EnabledTypes returns that connector is available on both internal and external listeners.
// External listener exposes ConnectorV1External (AI tool facing REST API).
// Internal listener exposes ConnectorV1Internal (connector process facing gRPC + HTTP API).
func (*connectorV1) EnabledTypes() (internal, external bool) {
	return true, true
}
