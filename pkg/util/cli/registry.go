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

package cli

import (
	goHttp "net/http"

	"github.com/regclient/regclient"
	"github.com/regclient/regclient/config"
	"github.com/regclient/regclient/scheme/reg"
	"github.com/spf13/cobra"
)

func NewRegistry() Registry {
	return registry{
		flagRegistryUseCredentials: Flag[bool]{
			Name:        "registry.docker.credentials",
			Description: "Use Docker Credentials",
			Default:     false,
		},

		flagRegistryInsecure: Flag[[]string]{
			Name:        "registry.docker.insecure",
			Description: "List of insecure registries",
			Default:     nil,
		},

		flagRegistryList: Flag[[]string]{
			Name:        "registry.docker.endpoint",
			Description: "List of boosted registries",
			Default:     nil,
			Hidden:      true,
		},
	}
}

type Registry interface {
	FlagRegisterer

	Client(cmd *cobra.Command, lm LicenseManager) (*regclient.RegClient, error)
}

type registry struct {
	flagRegistryUseCredentials Flag[bool]
	flagRegistryInsecure       Flag[[]string]
	flagRegistryList           Flag[[]string]
}

func (r registry) GetName() string {
	return "registry"
}

func (r registry) Register(cmd *cobra.Command) error {
	return RegisterFlags(
		cmd,
		r.flagRegistryList,
		r.flagRegistryInsecure,
		r.flagRegistryUseCredentials,
	)
}

func (r registry) Validate(cmd *cobra.Command) error {
	return ValidateFlags(
		r.flagRegistryList,
		r.flagRegistryInsecure,
		r.flagRegistryUseCredentials,
	)(cmd, nil)
}

func (r registry) Client(cmd *cobra.Command, lm LicenseManager) (*regclient.RegClient, error) {
	var flags = make([]regclient.Opt, 0, 3)

	flags = append(flags, regclient.WithConfigHostDefault(config.Host{
		ReqConcurrent: 8,
	}))

	flags = append(flags, regclient.WithRegOpts(reg.WithTransport(&goHttp.Transport{
		MaxConnsPerHost: 64,
		MaxIdleConns:    100,
	})))

	configs := map[string]config.Host{}

	ins, err := r.flagRegistryInsecure.Get(cmd)
	if err != nil {
		return nil, err
	}

	for _, el := range ins {
		v, ok := configs[el]
		if !ok {
			v.Name = el
			v.Hostname = el
		}

		v.TLS = config.TLSDisabled
		v.ReqConcurrent = 8

		configs[el] = v
	}

	regs, err := r.flagRegistryList.Get(cmd)
	if err != nil {
		return nil, err
	}

	for _, el := range regs {
		v, ok := configs[el]
		if !ok {
			v.Name = el
			v.Hostname = el
		}

		v.ReqConcurrent = 8

		configs[el] = v
	}

	// Hosts
	if lm != nil {
		registryConfigs, err := lm.RegistryHosts(cmd)
		if err == nil {
			for n, m := range registryConfigs {
				v, ok := configs[n]
				if !ok {
					v.Name = n
					v.Hostname = n
				}

				v = m(v)

				configs[n] = v
			}
		} else {
			logger.Err(err).Debug("Failed to initialize license manager, continuing...")
		}
	}

	if creds, err := r.flagRegistryUseCredentials.Get(cmd); err != nil {
		return nil, err
	} else if creds {
		flags = append(flags, regclient.WithDockerCreds())
	}

	for _, v := range configs {
		flags = append(flags, regclient.WithConfigHost(v))
	}

	return regclient.New(flags...), nil
}
