//
// DISCLAIMER
//
// Copyright 2026 ArangoDB GmbH, Cologne, Germany
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

package loader

import (
	"context"
	"os"
	"path"
	"time"

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/cache"
	utilToken "github.com/arangodb/kube-arangodb/pkg/util/token"
)

const MaxTokenSize = 1024 * 8

func SecretCacheDirectory(directory string, ttl time.Duration) cache.ObjectFetcher[utilToken.Secret] {
	return func(ctx context.Context) (utilToken.Secret, time.Duration, error) {
		s, err := LoadSecretSetFromDirectory(directory)
		if err != nil {
			return nil, 0, err
		}

		return s, ttl, nil
	}
}

func LoadSecretSetFromDirectory(directory string) (utilToken.Secret, error) {
	s, sz, err := LoadSecretsFromDirectory(directory)
	if err != nil {
		return nil, err
	}

	return utilToken.NewSecretSet(s, sz), nil
}

func LoadSecretsFromDirectory(directory string) (utilToken.Secret, utilToken.Secrets, error) {
	files, err := os.ReadDir(directory)
	if err != nil {
		return nil, nil, err
	}

	var ts = make(map[string][]byte)

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		data, err := util.OpenWithRead(path.Join(directory, file.Name()), MaxTokenSize)
		if err != nil {
			continue
		}

		if len(data) == 0 {
			continue
		}

		buff := make([]byte, MaxTokenSize)

		for id := range buff {
			buff[id] = 0
		}

		copy(buff, data)

		ts[file.Name()] = buff
	}

	return LoadSecretsFromData(ts)
}
