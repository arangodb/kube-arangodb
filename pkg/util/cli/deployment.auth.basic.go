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

package cli

import (
	"github.com/spf13/cobra"

	"github.com/arangodb/go-driver"
)

type deploymentBasicAuth struct {
	username Flag[string]
	password Flag[string]
}

func (d deploymentBasicAuth) GetName() string {
	return "basic"
}

func (d deploymentBasicAuth) Validate(cmd *cobra.Command) error {
	return nil
}

func (d deploymentBasicAuth) Register(cmd *cobra.Command) error {
	return RegisterFlags(
		cmd,
		d.username,
		d.password,
	)
}

func (d deploymentBasicAuth) Authentication(cmd *cobra.Command) (driver.Authentication, error) {
	if err := ValidateFlags(d.username, d.password)(cmd, nil); err != nil {
		return nil, err
	}

	username, err := d.username.Get(cmd)
	if err != nil {
		return nil, err
	}

	password, err := d.password.Get(cmd)
	if err != nil {
		return nil, err
	}

	return driver.BasicAuthentication(username, password), nil
}
