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

package helpers

import (
	"github.com/pkg/errors"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/secret"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
)

func SecretConfigGetter(s secret.Inspector, name, key string) kclient.ConfigGetter {
	return func() (*rest.Config, string, error) {
		secret, ok := s.Secret().V1().GetSimple(name)
		if !ok {
			return nil, "", errors.Errorf("Secret %s not found", name)
		}

		v, ok := secret.Data[key]
		if !ok {
			return nil, "", errors.Errorf("Key %s/%s not found", name, key)
		}

		cfg, err := clientcmd.RESTConfigFromKubeConfig(v)
		if err != nil {
			return nil, "", err
		}

		return cfg, util.SHA256(v), nil
	}
}
