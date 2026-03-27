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
	"net/url"
	"slices"
	goStrings "strings"

	"github.com/spf13/cobra"

	adbDriverV2Connection "github.com/arangodb/go-driver/v2/connection"

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	operatorHTTP "github.com/arangodb/kube-arangodb/pkg/util/http"
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

	Connection(cmd *cobra.Command) (adbDriverV2Connection.Connection, error)
	Authentication(cmd *cobra.Command) (adbDriverV2Connection.Authentication, error)
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

func (d deployment) Connection(cmd *cobra.Command) (adbDriverV2Connection.Connection, error) {
	insecure, err := d.insecure.Get(cmd)
	if err != nil {
		return nil, err
	}

	endpoint, err := d.endpoint.Get(cmd)
	if err != nil {
		return nil, err
	}

	auth, err := d.Authentication(cmd)
	if err != nil {
		return nil, err
	}

	return adbDriverV2Connection.NewHttpConnection(adbDriverV2Connection.HttpConfiguration{
		Authentication: auth,
		Endpoint:       adbDriverV2Connection.NewRoundRobinEndpoints(endpoint),
		ContentType:    adbDriverV2Connection.ApplicationJSON,
		ArangoDBConfig: adbDriverV2Connection.ArangoDBConfiguration{},
		Transport: operatorHTTP.RoundTripperWithShortTransport(
			operatorHTTP.WithTransportTLS(
				util.BoolSwitch(insecure, operatorHTTP.Insecure, nil),
			),
		),
	}), nil
}

func (d deployment) Authentication(cmd *cobra.Command) (adbDriverV2Connection.Authentication, error) {
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
