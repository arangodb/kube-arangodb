//
// DISCLAIMER
//
// Copyright 2024-2026 ArangoDB GmbH, Cologne, Germany
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
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	goStrings "strings"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pbConfigV1 "github.com/arangodb/kube-arangodb/integrations/config/v1/definition"
	pbSharedV1 "github.com/arangodb/kube-arangodb/integrations/shared/v1/definition"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/svc"
)

func New(config Config) (svc.Handler, error) {
	config.Init()

	if len(config.Modules) == 0 {
		return nil, errors.Errorf("Requires at least 1 module")
	}

	for module, moduleConfig := range config.Modules {
		if moduleConfig.Path == "" {
			return nil, errors.Errorf("Path for module `%s` cannot be empty", module)
		}

		if !path.IsAbs(moduleConfig.Path) {
			return nil, errors.Errorf("Path `%s` for module `%s` needs to be absolute", moduleConfig.Path, module)
		}

		info, err := os.Stat(moduleConfig.Path)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return nil, errors.Errorf("Path `%s` for module `%s` does not exists", moduleConfig.Path, module)
			}

			return nil, errors.Wrapf(err, "Path `%s` for module `%s` received unknown error", moduleConfig.Path, module)
		}

		if !info.IsDir() {
			return nil, errors.Errorf("Path `%s` for module `%s` is not a directory", moduleConfig.Path, module)
		}
	}

	return &impl{
		config: config,
	}, nil
}

var _ pbConfigV1.ConfigV1Server = &impl{}
var _ svc.Handler = &impl{}

type impl struct {
	pbConfigV1.UnsafeConfigV1Server

	config Config
}

func (i *impl) Gateway(ctx context.Context, mux *runtime.ServeMux) error {
	return nil
}

func (i *impl) Name() string {
	return Name
}

func (i *impl) Health(ctx context.Context) svc.HealthState {
	return svc.Healthy
}

func (i *impl) Register(registrar *grpc.Server) {
	pbConfigV1.RegisterConfigV1Server(registrar, i)
}

func (i *impl) Modules(ctx context.Context, empty *pbSharedV1.Empty) (*pbConfigV1.ConfigV1ModulesResponse, error) {
	res := &pbConfigV1.ConfigV1ModulesResponse{}

	res.Modules = util.SortKeys(i.config.Modules)

	return res, nil
}

func (i *impl) ModuleDetails(ctx context.Context, request *pbConfigV1.ConfigV1ModuleDetailsRequest) (*pbConfigV1.ConfigV1ModuleDetailsResponse, error) {
	if request.GetModule() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "Module name cannot be empty")
	}

	module, ok := i.config.Modules[request.GetModule()]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "Module `%s` not found", request.GetModule())
	}

	var resp pbConfigV1.ConfigV1ModuleDetailsResponse

	resp.Module = request.GetModule()

	var files []*pbConfigV1.ConfigV1File

	if err := filepath.Walk(module.Path, func(p string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.Mode().IsRegular() {
			return nil
		}

		if !goStrings.HasPrefix(p, fmt.Sprintf("%s/", module.Path)) {
			return nil
		}

		f, err := i.fileDetails(module, goStrings.TrimPrefix(p, fmt.Sprintf("%s/", module.Path)), request.GetChecksum())
		if err != nil {
			return err
		}

		files = append(files, f)

		return nil
	}); err != nil {
		if gErr, ok := svc.AsGRPCErrorStatus(err); ok {
			return nil, gErr
		}
		return nil, status.Errorf(codes.Internal, "Unable to list directory for module `%s`", request.GetModule())
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].GetPath() < files[j].GetPath()
	})

	resp.Files = files

	if request.GetChecksum() {
		checksums := make([]string, len(files))
		for id := range files {
			checksums[id] = fmt.Sprintf("%s:%s", files[id].GetPath(), files[id].GetChecksum())
		}

		resp.Checksum = util.NewType(util.SHA256FromStringArray(checksums...))
	}

	return &resp, nil
}

func (i *impl) FileDetails(ctx context.Context, request *pbConfigV1.ConfigV1FileDetailsRequest) (*pbConfigV1.ConfigV1FileDetailsResponse, error) {
	if request.GetModule() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "Module name cannot be empty")
	}

	module, ok := i.config.Modules[request.GetModule()]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "Module `%s` not found", request.GetModule())
	}

	if request.GetFile() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "File name cannot be empty")
	}

	if request.GetFile() == "" {
		return nil, status.Errorf(codes.NotFound, "File name cannot be empty")
	}

	f, err := i.fileDetails(module, request.GetFile(), request.GetChecksum())
	if err != nil {
		return nil, err
	}

	return &pbConfigV1.ConfigV1FileDetailsResponse{
		Module: request.GetModule(),
		File:   f,
	}, nil
}
