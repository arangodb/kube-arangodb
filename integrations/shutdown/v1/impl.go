//
// DISCLAIMER
//
// Copyright 2023-2025 ArangoDB GmbH, Cologne, Germany
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
	"io/fs"
	"path"
	"path/filepath"
	goStrings "strings"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"

	pbSharedV1 "github.com/arangodb/kube-arangodb/integrations/shared/v1/definition"
	pbShutdownV1 "github.com/arangodb/kube-arangodb/integrations/shutdown/v1/definition"
	pbStorageV2 "github.com/arangodb/kube-arangodb/integrations/storage/v2/definition"
	"github.com/arangodb/kube-arangodb/pkg/util/closer"
	utilConstants "github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	ugrpc "github.com/arangodb/kube-arangodb/pkg/util/grpc"
	"github.com/arangodb/kube-arangodb/pkg/util/shutdown"
	"github.com/arangodb/kube-arangodb/pkg/util/svc"
)

func New(cfg Configuration, c context.CancelFunc) svc.Handler {
	var z = &impl{
		closer: c,
		cfg:    cfg,
	}

	z.close = closer.CloseOnce(z)

	return z
}

var _ pbShutdownV1.ShutdownV1Server = &impl{}
var _ svc.Handler = &impl{}

type impl struct {
	pbShutdownV1.UnimplementedShutdownV1Server

	cfg Configuration

	close closer.Close

	closer context.CancelFunc
}

func (i *impl) Close() error {
	defer i.closer()

	ctx, c := context.WithTimeout(shutdown.Context(), i.cfg.Debug.Timeout)
	defer c()

	time.Sleep(50 * time.Millisecond)

	if i.cfg.Debug.Enabled {
		logger.Info("Enforce Debug collection")

		// Need to fetch Debug Details
		if addr, ok := utilConstants.INTEGRATION_SERVICE_ADDRESS.Lookup(); ok {
			client, close, err := ugrpc.NewGRPCClient(ctx, pbStorageV2.NewStorageV2Client, addr)
			if err != nil {
				return err
			}

			defer close.Close()

			if err := filepath.Walk(i.cfg.Debug.Path, func(p string, info fs.FileInfo, err error) error {
				logger := logger.Str("file", p)
				if info.IsDir() {
					logger.Info("Skip Directory")
					return nil
				}

				if info.Size() == 0 {
					logger.Info("Skip Empty File")
					return nil
				}

				logger = logger.Int64("size", info.Size())

				k := goStrings.TrimPrefix(p, i.cfg.Debug.Path)

				prefix := []string{
					"debug",
					"pod",
				}

				if v, ok := utilConstants.EnvOperatorPodNamespaceEnv.Lookup(); ok {
					prefix = append(prefix, v)
				}

				if v, ok := utilConstants.EnvOperatorPodNameEnv.Lookup(); ok {
					prefix = append(prefix, v)
				}

				prefix = append(prefix, k)

				logger.Str("key", k).Info("Sync file")

				if _, err := pbStorageV2.SendFile(ctx, client, path.Join(prefix...), p); err != nil {
					logger.Err(err).Warn("Failed to send file to server for DebugPackage")
				} else {
					logger.Info("Send Completed")
				}

				return nil
			}); err != nil {
				return err
			}
		} else {
			return errors.Errorf("Address of the Service not defined")
		}
	} else {
		logger.Info("Skip Debug collection")
	}

	return nil
}

func (i *impl) Name() string {
	return pbShutdownV1.Name
}

func (i *impl) Health() svc.HealthState {
	return svc.Healthy
}

func (i *impl) Register(registrar *grpc.Server) {
	pbShutdownV1.RegisterShutdownV1Server(registrar, i)
}

func (i *impl) Gateway(ctx context.Context, mux *runtime.ServeMux) error {
	return pbShutdownV1.RegisterShutdownV1HandlerServer(ctx, mux, i)
}

func (i *impl) Shutdown(ctx context.Context, empty *pbSharedV1.Empty) (*pbSharedV1.Empty, error) {
	go func() {
		if err := i.close.Close(); err != nil {
			logger.Err(err).Warn("Shutting down failed")
		}
	}()

	logger.Info("Shutting down")

	return &pbSharedV1.Empty{}, nil
}
