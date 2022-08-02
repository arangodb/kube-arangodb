//
// DISCLAIMER
//
// Copyright 2016-2022 ArangoDB GmbH, Cologne, Germany
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

package reconcile

import (
	"context"
	"crypto/sha256"
	"fmt"

	"github.com/arangodb/go-driver"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/client"
	"github.com/arangodb/kube-arangodb/pkg/logging"
	"github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	inspectorInterface "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector"
)

func mapTLSSNIConfig(sni api.TLSSNISpec, cachedStatus inspectorInterface.Inspector) (map[string]string, error) {
	fetchedSecrets := map[string]string{}

	mapping := sni.Mapping
	if len(mapping) == 0 {
		return fetchedSecrets, nil
	}

	for name, servers := range mapping {
		secret, exists := cachedStatus.Secret().V1().GetSimple(name)
		if !exists {
			return nil, errors.Newf("Secret %s does not exist", name)
		}

		tlsKey, ok := secret.Data[constants.SecretTLSKeyfile]
		if !ok {
			return nil, errors.Newf("Not found tls keyfile key in SNI secret")
		}

		tlsKeyChecksum := fmt.Sprintf("%0x", sha256.Sum256(tlsKey))

		for _, server := range servers {
			if _, ok := fetchedSecrets[server]; ok {
				return nil, errors.Newf("Not found tls key in SNI secret")
			}
			fetchedSecrets[server] = tlsKeyChecksum
		}
	}

	return fetchedSecrets, nil
}

func compareTLSSNIConfig(ctx context.Context, log logging.Logger, c driver.Connection, m map[string]string, refresh bool) (bool, error) {
	tlsClient := client.NewClient(c, log)

	f := tlsClient.GetTLS
	if refresh {
		f = tlsClient.RefreshTLS
	}

	tlsDetails, err := f(ctx)
	if err != nil {
		return false, errors.WithMessage(err, "Unable to fetch TLS SNI state")
	}

	if len(m) != len(tlsDetails.Result.SNI) {
		return false, errors.Newf("Count of SNI mounted secrets does not match")
	}

	for key, value := range tlsDetails.Result.SNI {
		currentValue, ok := m[key]
		if !ok {
			return false, errors.Newf("Unable to fetch TLS SNI state")
		}

		if value.GetSHA().Checksum() != currentValue {
			return false, nil
		}
	}

	return true, nil
}
