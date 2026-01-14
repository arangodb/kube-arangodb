//
// DISCLAIMER
//
// Copyright 2025-2026 ArangoDB GmbH, Cologne, Germany
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
	"github.com/regclient/regclient/config"
	"github.com/spf13/cobra"

	lmanager "github.com/arangodb/kube-arangodb/pkg/license_manager"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

func NewLicenseManager(prefix string) LicenseManager {
	return licenseManager{
		endpoint: Flag[string]{
			Name:        fmt.Sprintf("%s.endpoint", prefix),
			Default:     lmanager.ArangoLicenseManagerEndpoint,
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
				EnvEnabled:  true,
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
				Hidden: true,
			},

			clientSecret: Flag[string]{
				Name:        fmt.Sprintf("%s.client.secret", prefix),
				Description: "LicenseManager Client Secret",
				Default:     "",
				EnvEnabled:  true,
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

	LicenseManagerAuthProvider
	LicenseManagerEndpointProvider
}

type LicenseManagerEndpointProvider interface {
	Endpoint(cmd *cobra.Command) (string, error)
	Stages(cmd *cobra.Command) ([]string, error)
}

type LicenseManagerAuthProvider interface {
	ClientCredentials(cmd *cobra.Command) (string, string, error)
}

type LicenseManagerStaticAuthProvider func(cmd *cobra.Command) (string, string, error)

func (l LicenseManagerStaticAuthProvider) ClientCredentials(cmd *cobra.Command) (string, string, error) {
	return l(cmd)
}

func (l LicenseManagerStaticAuthProvider) ClientSecret(cmd *cobra.Command) (string, error) {
	_, value, err := l(cmd)
	return value, err
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

func (l licenseManager) ClientCredentials(cmd *cobra.Command) (string, string, error) {
	clientID, err := l.client.clientID.Get(cmd)
	if err != nil {
		return "", "", err
	}

	clientSecret, err := l.client.clientSecret.Get(cmd)
	if err != nil {
		return "", "", err
	}

	return clientID, clientSecret, nil
}

func (l licenseManager) GetName() string {
	return "lm"
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
	)(cmd, nil)
}

func LicenseManagerRegistryHosts(cmd *cobra.Command, endpoint LicenseManagerEndpointProvider, auth LicenseManagerAuthProvider) (map[string]util.ModR[config.Host], error) {
	clientID, clientSecret, err := auth.ClientCredentials(cmd)
	if err != nil {
		return nil, err
	}

	stages, err := endpoint.Stages(cmd)
	if err != nil {
		return nil, err
	}

	host, err := endpoint.Endpoint(cmd)
	if err != nil {
		return nil, err
	}

	var apply util.ModR[config.Host] = func(in config.Host) config.Host {
		in.User = clientID
		in.Pass = clientSecret
		in.ReqConcurrent = 8
		in.ReqPerSec = 128
		return in
	}

	ret := map[string]util.ModR[config.Host]{}

	for _, stage := range stages {
		ret[fmt.Sprintf("%s.registry.%s", stage, host)] = apply
		ret[fmt.Sprintf("%s.helm.%s", stage, host)] = apply

		if stage == "prd" {
			ret[fmt.Sprintf("registry.%s", host)] = apply
			ret[fmt.Sprintf("helm.%s", host)] = apply
		}
	}

	return ret, nil
}

func LicenseManagerClient(cmd *cobra.Command, endpoint LicenseManagerEndpointProvider, auth LicenseManagerAuthProvider) (lmanager.Client, error) {
	host, err := endpoint.Endpoint(cmd)
	if err != nil {
		return nil, err
	}

	clientID, clientSecret, err := auth.ClientCredentials(cmd)
	if err != nil {
		return nil, err
	}

	c := lmanager.NewClient(host, clientID, clientSecret)

	id, err := c.Identity(cmd.Context())
	if err != nil {
		return nil, err
	}

	logger.JSON("identity", id).Info("Using identity for client")

	return c, nil
}
