//
// DISCLAIMER
//
// Copyright 2024 ArangoDB GmbH, Cologne, Germany
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
	"io"
	"os"
	"path"
	"sort"
	"time"

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

const MaxSize = 128

type cache struct {
	parent *implementation

	eol time.Time

	signingToken []byte

	validationTokens [][]byte
}

func (i *implementation) newCache() (*cache, error) {
	files, err := os.ReadDir(i.cfg.Path)
	if err != nil {
		return nil, err
	}

	var keys []string
	var tokens = make(map[string][]byte)

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		data, err := util.OpenWithRead(path.Join(i.cfg.Path, file.Name()), MaxSize)
		if err != nil {
			continue
		}

		if len(data) == 0 {
			continue
		}

		buff := make([]byte, 32)

		for id := range buff {
			buff[id] = 0
		}

		copy(buff, data)

		keys = append(keys, file.Name())
		tokens[file.Name()] = buff
	}

	if len(keys) == 0 {
		return nil, io.ErrUnexpectedEOF
	}

	sort.Strings(keys)

	data := make([][]byte, len(keys))

	for id := range data {
		data[id] = tokens[keys[id]]
	}

	cache := cache{
		parent:           i,
		eol:              time.Now().Add(i.cfg.TTL),
		signingToken:     tokens[keys[0]],
		validationTokens: data,
	}

	return &cache, nil
}

func (i *implementation) localGetCache() *cache {
	if c := i.cache; c != nil && c.eol.After(time.Now()) {
		return c
	}

	return nil
}

func (i *implementation) withCache() (*cache, error) {
	if c := i.getCache(); c != nil {
		return c, nil
	}

	return i.refreshCache()
}

func (i *implementation) getCache() *cache {
	i.lock.RLock()
	defer i.lock.RUnlock()

	return i.localGetCache()
}

func (i *implementation) refreshCache() (*cache, error) {
	i.lock.Lock()
	defer i.lock.Unlock()

	if c := i.localGetCache(); c != nil {
		return c, nil
	}

	// Get was not successful, retry

	if c, err := i.newCache(); err != nil {
		return nil, err
	} else if c == nil {
		return nil, errors.Errorf("cache returned is nil")
	} else {
		i.cache = c
		return i.cache, nil
	}
}
