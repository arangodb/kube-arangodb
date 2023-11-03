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

package license

import (
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
)

type Config struct {
	Enabled bool

	Secret struct {
		Namespace, Name, Key, ClientFactory string
	}

	Env struct {
		Env string
	}

	Type string

	RefreshInterval time.Duration
	RefreshTimeout  time.Duration
}

func (c *Config) Init(cmd *cobra.Command) error {
	f := cmd.PersistentFlags()

	f.BoolVar(&c.Enabled, "license.enabled", true, "Define if LicenseManager is enabled")
	f.StringVar(&c.Type, "license.type", "secret", "Define type of the license fetch, possible values are secret or env")
	f.StringVar(&c.Secret.Namespace, "license.secret.namespace", "", "Define Secret Namespace for the Secret type of LicenseManager")
	f.StringVar(&c.Secret.Name, "license.secret.name", "", "Define Secret Name for the Secret type of LicenseManager")
	f.StringVar(&c.Secret.Key, "license.secret.key", "license", "Define Secret Key for the Secret type of LicenseManager")
	f.StringVar(&c.Secret.ClientFactory, "license.secret.client-factory", "", "Define K8S Client Factory for the Secret type of LicenseManager")
	f.StringVar(&c.Env.Env, "license.env.name", "ARANGODB_LICENSE", "Define Environment Variable name for the Env type of LicenseManager")
	f.DurationVar(&c.RefreshInterval, "license.refresh.interval", 30*time.Second, "Refresh interval for LicenseManager")
	f.DurationVar(&c.RefreshTimeout, "license.refresh.timeout", 3*time.Second, "Refresh timeout for LicenseManager")

	if err := f.MarkHidden("license.enabled"); err != nil {
		return err
	}
	if err := f.MarkHidden("license.secret.client-factory"); err != nil {
		return err
	}

	return nil
}

func (c Config) Enable() error {
	if !c.Enabled {
		return initManager(c, NewDisabledLoader())
	}

	switch c.Type {
	case "secret":
		return initManager(c, NewSecretLoader(kclient.GetFactory(c.Secret.ClientFactory), c.Secret.Namespace, c.Secret.Name, c.Secret.Key))

	case "env":
		if l, ok := os.LookupEnv(c.Env.Env); ok {
			return initManager(c, NewConstantLoader(l))
		}

		return initManager(c, NewDisabledLoader())
	default:
		return errors.Newf("Unsupported type for license.type: %s", c.Type)
	}
}
