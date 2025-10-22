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

	"github.com/google/uuid"
	"github.com/spf13/cobra"

	"github.com/arangodb/kube-arangodb/pkg/license/manager"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

func NewLicenseManager(prefix string) LicenseManager {
	return licenseManager{
		endpoint: Flag[string]{
			Name:        fmt.Sprintf("%s.endpoint", prefix),
			Default:     "license.arango.ai",
			Description: "LicenseManager Endpoint",
			Check: func(in string) error {
				if len(in) == 0 {
					return errors.Errorf("empty endpoint")
				}

				return nil
			},
		},

		client: licenseManagerClient{
			clientID: Flag[string]{
				Name:        fmt.Sprintf("%s.client.id", prefix),
				Description: "LicenseManager Client ID",
				Default:     "",
				Persistent:  false,
				Check: func(in string) error {
					if in == "" {
						return errors.New("Platform Client ID is required")
					}

					return nil
				},
			},

			stages: Flag[[]string]{
				Name:        fmt.Sprintf("%s.client.stage", prefix),
				Description: "LicenseManager Stages",
				Default:     []string{"prd"},
				Persistent:  false,
				Check: func(in []string) error {
					if len(in) == 0 {
						return errors.New("At least one stage needs to be defined")
					}

					return nil
				},
			},

			clientSecret: Flag[string]{
				Name:        "license.client.secret",
				Description: "LicenseManager Client Secret",
				Default:     "",
				Persistent:  false,
				Check: func(in string) error {
					if _, err := uuid.Parse(in); err != nil {
						return err
					}

					return nil
				},
			},
		},
	}
}

type LicenseManager interface {
	FlagRegisterer

	Endpoint(cmd *cobra.Command) (string, error)
	Stages(cmd *cobra.Command) ([]string, error)

	ClientID(cmd *cobra.Command) (string, error)
	ClientSecret(cmd *cobra.Command) (string, error)

	Client(cmd *cobra.Command) (manager.Client, error)
}

type licenseManager struct {
	endpoint Flag[string]

	client licenseManagerClient
}

func (l licenseManager) Endpoint(cmd *cobra.Command) (string, error) {
	return l.endpoint.Get(cmd)
}

func (l licenseManager) Stages(cmd *cobra.Command) ([]string, error) {
	return l.client.stages.Get(cmd)
}

func (l licenseManager) ClientID(cmd *cobra.Command) (string, error) {
	return l.client.clientID.Get(cmd)
}

func (l licenseManager) ClientSecret(cmd *cobra.Command) (string, error) {
	return l.client.clientSecret.Get(cmd)
}

func (l licenseManager) GetName() string {
	return "lm"
}

func (l licenseManager) Client(cmd *cobra.Command) (manager.Client, error) {
	endpoint, err := l.endpoint.Get(cmd)
	if err != nil {
		return nil, err
	}

	cid, err := l.client.clientID.Get(cmd)
	if err != nil {
		return nil, err
	}

	cs, err := l.client.clientSecret.Get(cmd)
	if err != nil {
		return nil, err
	}

	return manager.NewClient(endpoint, cid, cs)
}

func (l licenseManager) Register(cmd *cobra.Command) error {
	return RegisterFlags(
		cmd,
		l.endpoint,
		l.client,
	)
}

func (l licenseManager) Validate(cmd *cobra.Command) error {
	return ValidateFlags(
		l.endpoint,
		l.client,
	)(cmd, nil)
}
