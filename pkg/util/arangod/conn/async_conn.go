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
	"fmt"
	"net/http"

	"github.com/arangodb/go-driver"
)

const (
	asyncHeaderRequest  = "x-arango-async"
	asyncHeaderResponse = "x-arango-async-id"
)

// ASyncFunc function definition which helps to find all occurrences of its usage.
type ASyncFunc func(ctx context.Context, client2 driver.Client) error

// AsyncConnection describes specific asynchronous behaviour.
type AsyncConnection struct {
	// Connection is wrapped by the asynchronous connection.
	driver.Connection
	// jobID is used to follow the result using `/_api/job` API.
	jobID string
}

// Do perform asynchronous request and get job ID from the response header.
func (a *AsyncConnection) Do(ctx context.Context, req driver.Request) (driver.Response, error) {
	resp, err := a.Connection.Do(ctx, req)
	if err != nil {
		return nil, err
	}

	if a.jobID = resp.Header(asyncHeaderResponse); len(a.jobID) == 0 {
		return nil, fmt.Errorf("\"%s\" header should have non-empty value in the response", asyncHeaderResponse)
	}
	asyncResp := &AsyncResponse{
		Response: resp,
	}

	return asyncResp, nil
}

// GetJobID return jobID to follow the progress using `` API.
func (a *AsyncConnection) GetJobID() string {
	return a.jobID
}

// NewRequest adds additional header which allows running request asynchronously.
func (a *AsyncConnection) NewRequest(method, path string) (driver.Request, error) {
	req, err := a.Connection.NewRequest(method, path)
	if err != nil {
		return nil, err
	}

	req.SetHeader(asyncHeaderRequest, "store")
	return req, nil
}

// SetAuthentication wraps original set authentication method.
func (a *AsyncConnection) SetAuthentication(auth driver.Authentication) (driver.Connection, error) {
	if _, err := a.Connection.SetAuthentication(auth); err != nil {
		return nil, err
	}

	return a, nil
}

// AsyncResponse describes specific asynchronous response.
type AsyncResponse struct {
	// Response is wrapped by the asynchronous response.
	driver.Response
}

// CheckStatus checks whether a proper status is returned.
func (a *AsyncResponse) CheckStatus(validStatusCodes ...int) error {
	if a.Response.CheckStatus(http.StatusAccepted) == nil {
		return nil
	}

	return a.Response.CheckStatus(validStatusCodes...)
}
