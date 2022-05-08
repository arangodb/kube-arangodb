//
// DISCLAIMER
//
// Copyright 2016-2022 ArangoDB GmbH, Cologne, Germany
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

type connectionWrap func(c driver.Connection) (driver.Connection, error)

var _ driver.Connection = connectionPass{}

type connectionPass struct {
	c    driver.Connection
	wrap connectionWrap
}

func (c connectionPass) NewRequest(method, path string) (driver.Request, error) {
	return c.c.NewRequest(method, path)
}

func (c connectionPass) Do(ctx context.Context, req driver.Request) (driver.Response, error) {
	return c.c.Do(ctx, req)
}

func (c connectionPass) Unmarshal(data driver.RawObject, result interface{}) error {
	return c.c.Unmarshal(data, result)
}

func (c connectionPass) Endpoints() []string {
	return c.c.Endpoints()
}

func (c connectionPass) UpdateEndpoints(endpoints []string) error {
	return c.c.UpdateEndpoints(endpoints)
}

func (c connectionPass) SetAuthentication(authentication driver.Authentication) (driver.Connection, error) {
	newC, err := c.c.SetAuthentication(authentication)
	if err != nil {
		return nil, err
	}

	if f := c.wrap; f != nil {
		return f(newC)
	}
	return newC, nil
}

func (c connectionPass) Protocols() driver.ProtocolSet {
	return c.c.Protocols()
}
