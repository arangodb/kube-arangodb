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
	"time"

	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"
	integrationsShared "github.com/arangodb/kube-arangodb/pkg/integrations/shared"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

func NewConfiguration() Configuration {
	return Configuration{}
}

type Configuration struct {
	integrationsShared.Endpoint
	integrationsShared.Database

	Async ConfigurationAsync
}

func (c Configuration) Validate() error {
	return errors.Errors(
		shared.PrefixResourceError("async", c.Async.Validate()),
		shared.PrefixResourceError("endpoint", c.Endpoint.Validate()),
		shared.PrefixResourceError("database", c.Database.Validate()),
	)
}

type ConfigurationAsync struct {
	Enabled bool
	Size    int
	Retry   ConfigurationRetry
}

func (c ConfigurationAsync) Validate() error {
	if !c.Enabled {
		return nil
	}
	return errors.Errors(
		shared.PrefixResourceErrorFunc("size", func() error {
			if c.Size <= 0 {
				return errors.Errorf("size must be greater than zero")
			}

			return nil
		}),
		shared.PrefixResourceError("retry", c.Retry.Validate()),
	)
}

type ConfigurationRetry struct {
	Timeout time.Duration
	Delay   time.Duration
}

func (c ConfigurationRetry) Validate() error {
	return errors.Errors(
		shared.PrefixResourceErrorFunc("timeout", func() error {
			if c.Timeout <= 0 {
				return errors.Errorf("timeout must be greater than zero")
			}

			return nil
		}),
		shared.PrefixResourceErrorFunc("delay", func() error {
			if c.Delay <= 0 {
				return errors.Errorf("delay must be greater than zero")
			}

			return nil
		}),
	)
}

func (c Configuration) With(mods ...util.ModR[Configuration]) Configuration {
	n := c

	for _, mod := range mods {
		n = mod(n)
	}

	return n
}
