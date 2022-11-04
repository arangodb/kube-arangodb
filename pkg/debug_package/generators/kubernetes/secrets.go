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

package kubernetes

import (
	"context"
	"crypto/sha256"
	"fmt"

	"github.com/rs/zerolog"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/arangodb/kube-arangodb/pkg/debug_package/cli"
	"github.com/arangodb/kube-arangodb/pkg/debug_package/shared"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
)

func Secrets() shared.Factory {
	return shared.NewFactory("kubernetes-secrets", true, secrets)
}

func secrets(logger zerolog.Logger, files chan<- shared.File) error {
	k, ok := kclient.GetDefaultFactory().Client()
	if !ok {
		return errors.Newf("Client is not initialised")
	}

	secrets := map[types.UID]*core.Secret{}
	next := ""
	for {
		r, err := k.Kubernetes().CoreV1().Secrets(cli.GetInput().Namespace).List(context.Background(), meta.ListOptions{
			Continue: next,
		})

		if err != nil {
			return err
		}

		for _, e := range r.Items {
			hashed := make(map[string][]byte, len(e.Data))
			for k, v := range e.Data {
				if cli.GetInput().HideSensitiveData {
					hashed[k] = []byte(fmt.Sprintf("%02x", sha256.Sum256(v)))
				} else {
					hashed[k] = v
				}
			}
			secrets[e.UID] = e.DeepCopy()
			secrets[e.UID].Data = hashed
		}

		next = r.Continue
		if next == "" {
			break
		}
	}

	files <- shared.NewJSONFile("kubernetes/secrets.json", func() (interface{}, error) {
		q := make([]*core.Secret, 0, len(secrets))

		for _, e := range secrets {
			q = append(q, e)
		}

		return q, nil
	})

	return nil
}
