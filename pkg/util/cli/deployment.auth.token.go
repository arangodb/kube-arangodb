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
	"fmt"

	"github.com/spf13/cobra"

	"github.com/arangodb/go-driver"
)

type deploymentTokenAuth struct {
	token Flag[string]
}

func (d deploymentTokenAuth) GetName() string {
	return "token"
}

func (d deploymentTokenAuth) Validate(cmd *cobra.Command) error {
	return nil
}

func (d deploymentTokenAuth) Register(cmd *cobra.Command) error {
	return RegisterFlags(
		cmd,
		d.token,
	)
}

func (d deploymentTokenAuth) Authentication(cmd *cobra.Command) (driver.Authentication, error) {
	if err := ValidateFlags(d.token)(cmd, nil); err != nil {
		return nil, err
	}

	token, err := d.token.Get(cmd)
	if err != nil {
		return nil, err
	}

	return driver.RawAuthentication(fmt.Sprintf("bearer %s", token)), nil
}
