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

package platform

import (
	"log/slog"
	goHttp "net/http"

	"github.com/regclient/regclient"
	"github.com/regclient/regclient/config"
	"github.com/regclient/regclient/scheme/reg"
	"github.com/spf13/cobra"
)

func getRegClient(cmd *cobra.Command) (*regclient.RegClient, error) {
	var flags = make([]regclient.Opt, 0, 3)

	slog.SetLogLoggerLevel(slog.LevelDebug)

	flags = append(flags, regclient.WithRegOpts(reg.WithTransport(&goHttp.Transport{
		MaxConnsPerHost: 64,
		MaxIdleConns:    100,
	})), regclient.WithSlog(slog.Default()))

	configs := map[string]config.Host{}

	ins, err := flagRegistryInsecure.Get(cmd)
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

		configs[el] = v
	}

	if creds, err := flagRegistryUseCredentials.Get(cmd); err != nil {
		return nil, err
	} else if creds {
		flags = append(flags, regclient.WithDockerCreds())
	}

	for _, v := range configs {
		flags = append(flags, regclient.WithConfigHost(v))
	}

	return regclient.New(flags...), nil
}
