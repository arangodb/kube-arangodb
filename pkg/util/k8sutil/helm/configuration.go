//
// DISCLAIMER
//
// Copyright 2024 ArangoDB GmbH, Cologne, Germany
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

package helm

import (
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
)

type Configuration struct {
	Namespace string

	Client kclient.Client

	Driver *ConfigurationDriver
}

func (c *Configuration) Validate() error {
	if c == nil {
		return errors.Errorf("Configuration cannot be nil")
	}

	if c.Namespace == "" {
		return errors.Errorf("Namespace cannot be empty")
	}

	if c.Client == nil {
		return errors.Errorf("Namespace cannot be empty")
	}

	if err := c.Driver.Validate(); err != nil {
		return err
	}

	return nil
}
