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
	pbImplStorageV1SharedS3 "github.com/arangodb/kube-arangodb/integrations/storage/v1/shared/s3"
)

type Mod func(c Configuration) Configuration

type ConfigurationType string

const (
	ConfigurationTypeS3 ConfigurationType = "s3"
)

func NewConfiguration(mods ...Mod) Configuration {
	var cfg Configuration

	return cfg.With(mods...)
}

type Configuration struct {
	Type ConfigurationType

	S3 pbImplStorageV1SharedS3.Configuration
}

func (c Configuration) Validate() error {
	return nil
}

func (c Configuration) With(mods ...Mod) Configuration {
	n := c

	for _, mod := range mods {
		n = mod(n)
	}

	return n
}
