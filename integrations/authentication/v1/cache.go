//
// DISCLAIMER
//
// Copyright 2024-2025 ArangoDB GmbH, Cologne, Germany
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

package v1

import (
	"context"
	"io"
	"os"
	"path"
	"sort"
	"time"

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/token"
)

const MaxSize = 128

func newCache(cfg Configuration) func(ctx context.Context) (token.Secret, time.Duration, error) {
	return func(ctx context.Context) (token.Secret, time.Duration, error) {
		files, err := os.ReadDir(cfg.Path)
		if err != nil {
			return nil, 0, err
		}

		var keys []string
		var ts = make(map[string][]byte)

		for _, file := range files {
			if file.IsDir() {
				continue
			}

			data, err := util.OpenWithRead(path.Join(cfg.Path, file.Name()), MaxSize)
			if err != nil {
				continue
			}

			if len(data) == 0 {
				continue
			}

			buff := make([]byte, cfg.Create.MaxSize)

			for id := range buff {
				buff[id] = 0
			}

			copy(buff, data)

			keys = append(keys, file.Name())
			ts[file.Name()] = buff
		}

		if len(keys) == 0 {
			return nil, 0, io.ErrUnexpectedEOF
		}

		sort.Strings(keys)

		data := make([][]byte, len(keys))

		for id := range data {
			data[id] = ts[keys[id]]
		}

		return token.NewSecretSet(token.NewSecret(ts[keys[0]]), util.FormatList(data, func(a []byte) token.Secret {
			return token.NewSecret(a)
		})...), cfg.TTL, nil
	}
}
