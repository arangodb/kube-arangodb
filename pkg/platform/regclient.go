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
	"github.com/regclient/regclient"
	"github.com/spf13/cobra"
)

func getRegClient(cmd *cobra.Command) (*regclient.RegClient, error) {
	var flags = make([]regclient.Opt, 0, 1)

	if creds, err := flagRegistryUseCredentials.Get(cmd); err != nil {
		return nil, err
	} else if creds {
		flags = append(flags, regclient.WithDockerCreds())
	}

	return regclient.New(flags...), nil
}
