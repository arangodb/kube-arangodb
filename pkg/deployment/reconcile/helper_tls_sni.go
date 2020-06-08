//
// DISCLAIMER
//
// Copyright 2020 ArangoDB GmbH, Cologne, Germany
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
// Author Adam Janikowski
//

package reconcile

import (
	"context"
	"crypto/sha256"
	"fmt"

	"github.com/arangodb/go-driver"
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/client"
	"github.com/arangodb/kube-arangodb/pkg/deployment/resources/inspector"
	"github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

func mapTLSSNIConfig(log zerolog.Logger, sni api.TLSSNISpec, cachedStatus inspector.Inspector) (map[string]string, error) {
	fetchedSecrets := map[string]string{}

	mapping := sni.Mapping
	if len(mapping) == 0 {
		return fetchedSecrets, nil
	}

	for name, servers := range mapping {
		secret, exists := cachedStatus.Secret(name)
		if !exists {
			return nil, errors.Errorf("Secret %s does not exist", name)
		}

		tlsKey, ok := secret.Data[constants.SecretTLSKeyfile]
		if !ok {
			return nil, errors.Errorf("Not found tls keyfile key in SNI secret")
		}

		tlsKeyChecksum := fmt.Sprintf("%0x", sha256.Sum256(tlsKey))

		for _, server := range servers {
			if _, ok := fetchedSecrets[server]; ok {
				return nil, errors.Errorf("Not found tls key in SNI secret")
			}
			fetchedSecrets[server] = tlsKeyChecksum
		}
	}

	return fetchedSecrets, nil
}

func compareTLSSNIConfig(ctx context.Context, c driver.Connection, m map[string]string, refresh bool) (bool, error) {
	tlsClient := client.NewClient(c)

	f := tlsClient.GetTLS
	if refresh {
		f = tlsClient.RefreshTLS
	}

	tlsDetails, err := f(ctx)
	if err != nil {
		return false, errors.WithMessage(err, "Unable to fetch TLS SNI state")
	}

	if len(m) != len(tlsDetails.Result.SNI) {
		return false, errors.Errorf("Count of SNI mounted secrets does not match")
	}

	for key, value := range tlsDetails.Result.SNI {
		currentValue, ok := m[key]
		if !ok {
			return false, errors.Errorf("Unable to fetch TLS SNI state")
		}

		if value.Checksum != currentValue {
			return false, nil
		}
	}

	return true, nil
}
