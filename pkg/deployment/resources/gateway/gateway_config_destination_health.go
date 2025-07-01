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

package gateway

import (
	"time"

	pbEnvoyCoreV3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	"google.golang.org/protobuf/types/known/durationpb"

	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/util"
)

type ConfigDestinationHealthChecks []ConfigDestinationHealthCheck

func (c ConfigDestinationHealthChecks) Validate() error {
	return shared.ValidateInterfaceList(c)
}

func (c ConfigDestinationHealthChecks) Render() []*pbEnvoyCoreV3.HealthCheck {
	ret := make([]*pbEnvoyCoreV3.HealthCheck, len(c))
	for id := range c {
		ret[id] = c[id].Render()
	}
	return ret
}

type ConfigDestinationHealthCheck struct {
	Timeout *time.Duration `json:"timeout,omitempty"`

	Interval *time.Duration `json:"interval,omitempty"`
}

func (c ConfigDestinationHealthCheck) Validate() error {
	return nil
}

func (c ConfigDestinationHealthCheck) Render() *pbEnvoyCoreV3.HealthCheck {
	return &pbEnvoyCoreV3.HealthCheck{
		Timeout:  durationpb.New(util.OptionalType(c.Timeout, time.Second)),
		Interval: durationpb.New(util.OptionalType(c.Interval, time.Second)),

		HealthChecker: &pbEnvoyCoreV3.HealthCheck_TcpHealthCheck_{
			TcpHealthCheck: &pbEnvoyCoreV3.HealthCheck_TcpHealthCheck{},
		},
	}
}
