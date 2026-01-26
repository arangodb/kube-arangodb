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
	"crypto/tls"
	"fmt"
	"net/url"
	"slices"
	goStrings "strings"

	"github.com/spf13/cobra"

	"github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/http"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

func NewDeployment(prefix string) Deployment {
	return deployment{
		endpoint: Flag[[]string]{
			Name:        fmt.Sprintf("%s.endpoint", prefix),
			Description: "Arango Endpoint",
			Check: func(in []string) error {
				if len(in) == 0 {
					return errors.Errorf("empty endpoint list")
				}

				for _, z := range in {
					if _, err := url.Parse(z); err != nil {
						return errors.WithMessage(err, "invalid endpoint")
					}
				}

				return nil
			},
		},

		insecure: Flag[bool]{
			Name:        fmt.Sprintf("%s.insecure", prefix),
			Description: "Skips TLS certificate verification",
		},

		authentication: Flag[string]{
			Name:        fmt.Sprintf("%s.authentication", prefix),
			Description: "Arango Endpoint Auth Method. One of: Disabled, Basic, Token",
			Default:     "Disabled",
			Check: func(in string) error {
				allowed := []string{"Disabled", "Basic", "Token"}
				if !slices.Contains(allowed, in) {
					return errors.Errorf("invalid auth method: %s. Allowed: %s", in, goStrings.Join(allowed, ", "))
				}
				return nil
			},
		},

		basic: deploymentBasicAuth{

			username: Flag[string]{
				Name:        fmt.Sprintf("%s.basic.username", prefix),
				Description: "Arango Username for Basic Authentication",
				Default:     "",
				Check: func(in string) error {
					if in == "" {
						return errors.Errorf("empty username")
					}

					return nil
				},
			},

			password: Flag[string]{
				Name:        fmt.Sprintf("%s.basic.password", prefix),
				Description: "Arango Password for Basic Authentication",
				Default:     "",
			},
		},

		token: deploymentTokenAuth{
			token: Flag[string]{
				Name:        fmt.Sprintf("%s.token", prefix),
				Description: "Arango JWT Token for Authentication",
				Default:     "",
				Check: func(in string) error {
					if in == "" {
						return errors.Errorf("empty token")
					}

					return nil
				},
			},
		},
	}
}

type Deployment interface {
	FlagRegisterer

	Connection(cmd *cobra.Command) (driver.Connection, error)
	Authentication(cmd *cobra.Command) (driver.Authentication, error)
}

type deployment struct {
	prefix string

	endpoint       Flag[[]string]
	insecure       Flag[bool]
	authentication Flag[string]

	basic deploymentBasicAuth
	token deploymentTokenAuth
}

func (d deployment) GetName() string {
	return d.prefix
}

func (d deployment) Register(cmd *cobra.Command) error {
	return RegisterFlags(
		cmd,
		d.endpoint,
		d.insecure,
		d.authentication,
		d.basic,
		d.token,
	)
}

func (d deployment) Validate(cmd *cobra.Command) error {
	return ValidateFlags(
		d.endpoint,
	)(cmd, nil)
}

func (d deployment) Connection(cmd *cobra.Command) (driver.Connection, error) {
	var t tls.Config

	if insecure, err := d.insecure.Get(cmd); err != nil {
		return nil, err
	} else {
		t.InsecureSkipVerify = insecure
	}

	endpoint, err := d.endpoint.Get(cmd)
	if err != nil {
		return nil, err
	}

	conn, err := http.NewConnection(http.ConnectionConfig{
		Endpoints: endpoint,
		TLSConfig: &t,
	})
	if err != nil {
		return nil, err
	}

	auth, err := d.Authentication(cmd)
	if err != nil {
		return nil, err
	}

	if auth != nil {
		conn, err = conn.SetAuthentication(auth)
		if err != nil {
			return nil, err
		}
	}

	return conn, nil
}

func (d deployment) Authentication(cmd *cobra.Command) (driver.Authentication, error) {
	auth, err := d.authentication.Get(cmd)
	if err != nil {
		return nil, err
	}

	switch auth {
	case "Disabled":
		return nil, nil
	case "Basic":
		return d.basic.Authentication(cmd)
	case "Token":
		return d.token.Authentication(cmd)
	default:
		return nil, errors.Errorf("invalid auth method: %s", auth)
	}
}
