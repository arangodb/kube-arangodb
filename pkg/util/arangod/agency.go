//
// DISCLAIMER
//
// Copyright 2018 ArangoDB GmbH, Cologne, Germany
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
// Author Ewout Prangsma
//

package arangod

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	driver "github.com/arangodb/go-driver"
	"github.com/pkg/errors"
)

// Agency provides API implemented by the ArangoDB agency.
type Agency interface {
	// ReadKey reads the value of a given key in the agency.
	ReadKey(ctx context.Context, key []string, value interface{}) error
	// Endpoint returns the endpoint of this agent connection
	Endpoint() string
}

// NewAgencyClient creates a new Agency connection from the given client
// connection.
// The number of endpoints of the client must be exactly 1.
func NewAgencyClient(c driver.Client) (Agency, error) {
	if len(c.Connection().Endpoints()) > 1 {
		return nil, maskAny(fmt.Errorf("Got multiple endpoints"))
	}
	return &agency{
		conn: c.Connection(),
	}, nil
}

type agency struct {
	conn driver.Connection
}

// ReadKey reads the value of a given key in the agency.
func (a *agency) ReadKey(ctx context.Context, key []string, value interface{}) error {
	conn := a.conn
	req, err := conn.NewRequest("POST", "_api/agency/read")
	if err != nil {
		return maskAny(err)
	}
	fullKey := createFullKey(key)
	input := [][]string{{fullKey}}
	req, err = req.SetBody(input)
	if err != nil {
		return maskAny(err)
	}
	//var raw []byte
	//ctx = driver.WithRawResponse(ctx, &raw)
	resp, err := conn.Do(ctx, req)
	if err != nil {
		return maskAny(err)
	}
	if resp.StatusCode() == 307 {
		// Not leader
		location := resp.Header("Location")
		return NotLeaderError{Leader: location}
	}
	if err := resp.CheckStatus(200, 201, 202); err != nil {
		return maskAny(err)
	}
	//fmt.Printf("Agent response: %s\n", string(raw))
	elems, err := resp.ParseArrayBody()
	if err != nil {
		return maskAny(err)
	}
	if len(elems) != 1 {
		return maskAny(fmt.Errorf("Expected 1 element, got %d", len(elems)))
	}
	// If empty key parse directly
	if len(key) == 0 {
		if err := elems[0].ParseBody("", &value); err != nil {
			return maskAny(err)
		}
	} else {
		// Now remove all wrapping objects for each key element
		var rawObject map[string]interface{}
		if err := elems[0].ParseBody("", &rawObject); err != nil {
			return maskAny(err)
		}
		var rawMsg interface{}
		for keyIndex := 0; keyIndex < len(key); keyIndex++ {
			if keyIndex > 0 {
				var ok bool
				rawObject, ok = rawMsg.(map[string]interface{})
				if !ok {
					return maskAny(fmt.Errorf("Data is not an object at key %s", key[:keyIndex+1]))
				}
			}
			var found bool
			rawMsg, found = rawObject[key[keyIndex]]
			if !found {
				return errors.Wrapf(KeyNotFoundError, "Missing data at key %s", key[:keyIndex+1])
			}
		}
		// Encode to json ...
		encoded, err := json.Marshal(rawMsg)
		if err != nil {
			return maskAny(err)
		}
		// and decode back into result
		if err := json.Unmarshal(encoded, &value); err != nil {
			return maskAny(err)
		}
	}

	//	fmt.Printf("result as JSON: %s\n", rawResult)
	return nil
}

// Endpoint returns the endpoint of this agent connection
func (a *agency) Endpoint() string {
	ep := a.conn.Endpoints()
	if len(ep) == 0 {
		return ""
	}
	return ep[0]
}

func createFullKey(key []string) string {
	return "/" + strings.Join(key, "/")
}
