//
// DISCLAIMER
//
// Copyright 2024-2025 ArangoDB GmbH, Cologne, Germany
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
	"fmt"
	"os"
	"path"
	goStrings "strings"
	"syscall"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	pbConfigV1 "github.com/arangodb/kube-arangodb/integrations/config/v1/definition"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

func (i *impl) fileDetails(module ModuleDefinition, file string, checksum bool) (*pbConfigV1.ConfigV1File, error) {
	expectedPath := path.Clean(path.Join(module.Path, file))

	if !goStrings.HasPrefix(expectedPath, fmt.Sprintf("%s/", module.Path)) {
		return nil, status.Errorf(codes.InvalidArgument, "File name cannot be empty")
	}

	stat, err := os.Stat(expectedPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, status.Errorf(codes.NotFound, "File `%s` not found within module `%s`", file, module.Name)
		}

		logger.Err(err).Str("module", module.Name).Str("file", file).Str("real-path", expectedPath).Warn("Unable to get file")
		return nil, status.Errorf(codes.Internal, "Unable to list directory for module `%s`", module.Name)
	}

	finfo, ok := stat.Sys().(*syscall.Stat_t)
	if !ok {
		logger.Str("module", module.Name).Str("file", file).Str("real-path", expectedPath).Warn("Invalid Stat Pointer for file")
		return nil, status.Errorf(codes.Internal, "Fetch of file `%s` within module `%s` failed", file, module.Name)
	}

	var f pbConfigV1.ConfigV1File

	f.Path = goStrings.TrimPrefix(expectedPath, fmt.Sprintf("%s/", module.Path))
	f.Size = finfo.Size
	f.Created = timestamppb.New(time.Unix(finfo.Ctim.Sec, finfo.Ctim.Nsec))
	f.Updated = timestamppb.New(time.Unix(finfo.Mtim.Sec, finfo.Mtim.Nsec))

	if checksum {
		c, err := util.SHA256FromFile(expectedPath)
		if err != nil {
			logger.Str("module", module.Name).Str("file", file).Str("real-path", expectedPath).Warn("Unable to get file checksum")
			return nil, status.Errorf(codes.Internal, "Unable to calculate checksum of file `%s` within module `%s` failed", file, module.Name)
		}
		f.Checksum = util.NewType(c)
	}

	return &f, nil
}
