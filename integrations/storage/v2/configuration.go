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

package v2

import (
	"context"

	pbImplStorageV2Shared "github.com/arangodb/kube-arangodb/integrations/storage/v2/shared"
	pbImplStorageV2SharedGCS "github.com/arangodb/kube-arangodb/integrations/storage/v2/shared/gcs"
	pbImplStorageV2SharedS3 "github.com/arangodb/kube-arangodb/integrations/storage/v2/shared/s3"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

type ConfigurationType string

const (
	ConfigurationTypeS3  ConfigurationType = "s3"
	ConfigurationTypeGCS ConfigurationType = "gcs"
)

func NewConfiguration(mods ...util.ModR[Configuration]) Configuration {
	var cfg Configuration

	return cfg.With(mods...)
}

type Configuration struct {
	Type ConfigurationType

	S3  pbImplStorageV2SharedS3.Configuration
	GCS pbImplStorageV2SharedGCS.Configuration
}

func (c Configuration) IO(ctx context.Context) (pbImplStorageV2Shared.IO, error) {
	switch c.Type {
	case ConfigurationTypeS3:
		return c.S3.New()
	case ConfigurationTypeGCS:
		return c.GCS.New(ctx)
	default:
		return nil, errors.Errorf("Unknown Type: %s", c.Type)
	}
}

func (c Configuration) Validate() error {
	return nil
}

func (c Configuration) With(mods ...util.ModR[Configuration]) Configuration {
	n := c

	for _, mod := range mods {
		n = mod(n)
	}

	return n
}
