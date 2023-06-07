//
// DISCLAIMER
//
// Copyright 2016-2023 ArangoDB GmbH, Cologne, Germany
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
// Adam Janikowski
//

package conn

import (
	http2 "net/http"

	"github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/agency"
	"github.com/arangodb/go-driver/http"

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

type Auth func() (driver.Authentication, error)
type Config func() (http.ConnectionConfig, error)

type Factory interface {
	Connection(hosts ...string) (driver.Connection, error)
	AgencyConnection(hosts ...string) (driver.Connection, error)

	Client(hosts ...string) (driver.Client, error)
	Agency(hosts ...string) (agency.Agency, error)

	RawConnection(host string) (Connection, error)

	GetAuth() Auth
}

func NewFactory(auth Auth, config Config) Factory {
	return &factory{
		auth:   auth,
		config: config,
	}
}

type factory struct {
	auth   Auth
	config Config
}

func (f factory) RawConnection(host string) (Connection, error) {
	cfg, err := f.config()
	if err != nil {
		return nil, err
	}

	var authString *string

	if f.auth != nil {
		auth, err := f.auth()
		if err != nil {
			return nil, err
		}

		if auth != nil {
			if auth.Type() != driver.AuthenticationTypeRaw {
				return nil, errors.Newf("Only RAW Authentication is supported")
			}

			authString = util.NewType(auth.Get("value"))
		}
	}

	return connection{
		auth: authString,
		host: host,
		client: &http2.Client{
			Transport: cfg.Transport,
			CheckRedirect: func(req *http2.Request, via []*http2.Request) error {
				return http2.ErrUseLastResponse
			},
		},
	}, nil
}

func (f factory) GetAuth() Auth {
	return f.auth
}

func (f factory) AgencyConnection(hosts ...string) (driver.Connection, error) {
	cfg, err := f.config()
	if err != nil {
		return nil, err
	}

	cfg.Endpoints = hosts

	conn, err := agency.NewAgencyConnection(cfg)
	if err != nil {
		return nil, err
	}

	if f.auth == nil {
		return conn, nil
	}
	auth, err := f.auth()
	if err != nil {
		return nil, err
	}
	if auth == nil {
		return conn, nil
	}
	return conn.SetAuthentication(auth)
}

func (f factory) Client(hosts ...string) (driver.Client, error) {
	conn, err := f.Connection(hosts...)
	if err != nil {
		return nil, err
	}

	config := driver.ClientConfig{
		Connection: conn,
	}

	if f.auth != nil {
		auth, err := f.auth()
		if err != nil {
			return nil, err
		}

		if auth != nil {
			config.Authentication = auth
		}
	}

	return driver.NewClient(config)
}

func (f factory) Agency(hosts ...string) (agency.Agency, error) {
	conn, err := f.AgencyConnection(hosts...)
	if err != nil {
		return nil, err
	}

	return agency.NewAgency(conn)
}

func (f factory) Connection(hosts ...string) (driver.Connection, error) {
	cfg, err := f.config()
	if err != nil {
		return nil, err
	}

	cfg.Endpoints = hosts

	conn, err := http.NewConnection(cfg)
	if err != nil {
		return nil, err
	}

	if f.auth == nil {
		return conn, nil
	}
	auth, err := f.auth()
	if err != nil {
		return nil, err
	}
	if auth == nil {
		return conn, nil
	}
	return conn.SetAuthentication(auth)
}
