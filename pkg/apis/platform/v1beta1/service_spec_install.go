//
// DISCLAIMER
//
// Copyright 2026 ArangoDB GmbH, Cologne, Germany
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

package v1beta1

import (
	"time"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

type ArangoPlatformServiceSpecInstall struct {
	// Timeout defines the install timeout
	// +doc/default: 20m
	Timeout *meta.Duration `json:"timeout,omitempty"`
}

func (c *ArangoPlatformServiceSpecInstall) GetTimeout() time.Duration {
	if c == nil || c.Timeout == nil {
		return time.Minute * 20
	}

	return c.Timeout.Duration
}

func (c *ArangoPlatformServiceSpecInstall) Validate() error {
	if c == nil {
		return nil
	}

	return shared.WithErrors(
		shared.ValidateOptionalPath("timeout", c.Timeout, func(duration meta.Duration) error {
			if duration.Duration < 0 {
				return errors.New("timeout invalid - negative duration")
			}

			return nil
		}),
	)
}
