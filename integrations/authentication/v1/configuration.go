//
// DISCLAIMER
//
// Copyright 2024 ArangoDB GmbH, Cologne, Germany
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

package v1

import (
	"time"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/strings"
)

const (
	DefaultAdminUser = "root"

	DefaultUser = "root"

	DefaultTTL = 15 * time.Second

	DefaultTokenMinTTL     = time.Minute
	DefaultTokenMaxTTL     = time.Hour
	DefaultTokenDefaultTTL = time.Hour
)

type Mod func(c Configuration) Configuration

func NewConfiguration() Configuration {
	return Configuration{
		Enabled: true,
		TTL:     DefaultTTL,
		Path:    "",
		Create: Token{
			DefaultUser:  DefaultUser,
			AllowedUsers: nil,
			MinTTL:       DefaultTokenMinTTL,
			MaxTTL:       DefaultTokenMaxTTL,
			DefaultTTL:   DefaultTokenDefaultTTL,
		},
	}
}

type Configuration struct {
	Enabled bool

	TTL time.Duration

	Path string

	Create Token
}

func (c Configuration) With(mods ...Mod) Configuration {
	n := c

	for _, mod := range mods {
		n = mod(n)
	}

	return n
}

func (c Configuration) Validate() error {
	if c.Path == "" {
		return errors.Errorf("Path should not be empty")
	}

	if c.TTL < 0 {
		return errors.Errorf("TTLS should be not negative")
	}

	if err := c.Create.Validate(); err != nil {
		return errors.Wrapf(err, "Token validation failed")
	}

	return nil
}

type Token struct {
	DefaultUser string

	AllowedUsers []string

	MinTTL, MaxTTL, DefaultTTL time.Duration
}

func (t Token) Validate() error {
	if t.MinTTL < 0 {
		return errors.Errorf("MinTTL Cannot be lower than 0")
	}

	if t.MaxTTL < t.MinTTL {
		return errors.Errorf("MaxTTL Cannot be lower than MinTTL")
	}

	if t.DefaultTTL < t.MinTTL {
		return errors.Errorf("DefautTTL Cannot be lower than MinTTL")
	}

	if t.DefaultTTL > t.MaxTTL {
		return errors.Errorf("DefautTTL Cannot be higher than MaxTTL")
	}

	if len(t.AllowedUsers) > 0 {
		// We are enforcing allowed users

		if !strings.ListContains(t.AllowedUsers, t.DefaultUser) {
			return errors.Errorf("DefaultUser should be always allowed")
		}
	}

	return nil
}
