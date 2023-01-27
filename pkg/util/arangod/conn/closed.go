//
// DISCLAIMER
//
// Copyright 2023 ArangoDB GmbH, Cologne, Germany
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

package conn

import (
	"context"

	"github.com/arangodb/go-driver"
)

func NewClosedConnection() driver.Connection {
	return closedConnection{}
}

func newClosedConnectionError() error {
	return closedConnectionError{}
}

type closedConnectionError struct {
}

func (c closedConnectionError) Error() string {
	return "Connection Closed"
}

type closedConnection struct {
}

func (c closedConnection) NewRequest(method, path string) (driver.Request, error) {
	return nil, newClosedConnectionError()
}

func (c closedConnection) Do(ctx context.Context, req driver.Request) (driver.Response, error) {
	return nil, newClosedConnectionError()
}

func (c closedConnection) Unmarshal(data driver.RawObject, result interface{}) error {
	return newClosedConnectionError()
}

func (c closedConnection) Endpoints() []string {
	return nil
}

func (c closedConnection) UpdateEndpoints(endpoints []string) error {
	return newClosedConnectionError()
}

func (c closedConnection) SetAuthentication(authentication driver.Authentication) (driver.Connection, error) {
	return c, nil
}

func (c closedConnection) Protocols() driver.ProtocolSet {
	return nil
}
