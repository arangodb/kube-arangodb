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

package azure

import (
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

type Config struct {
	AccountName string

	Endpoint string

	Provider Provider
}

func (c Config) GetCredentials() (azcore.TokenCredential, error) {
	return c.Provider.GetCredentials()
}

func (c Config) GetEndpoint() (string, error) {
	if f := c.Endpoint; f != "" {
		return f, nil
	}

	if f := c.AccountName; f != "" {
		return fmt.Sprintf("https://%s.blob.core.windows.net/", f), nil
	}

	return "", errors.Errorf("account name or url not provided")
}
