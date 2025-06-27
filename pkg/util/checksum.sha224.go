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

package util

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	goStrings "strings"
)

func SHA224FromExtract[T any](extract func(T) string, obj ...T) string {
	return SHA224FromStringArray(goStrings.Join(FormatList(obj, extract), "|"))
}

func SHA224FromHashArray[T Hash](data []T) string {
	return SHA224FromExtract(func(t T) string {
		return t.Hash()
	}, data...)
}

func SHA224FromStringArray(data ...string) string {
	return SHA224FromString(goStrings.Join(data, "|"))
}

func SHA224FromStringMap(data map[string]string) string {
	return SHA224FromExtract(func(t KV[string, string]) string {
		return fmt.Sprintf("%s:%s", t.K, SHA224FromString(t.V))
	}, ExtractWithSort(data, func(i, j string) bool {
		return i < j
	})...)
}

func SHA224FromString(data string) string {
	return SHA224([]byte(data))
}

func SHA224(data []byte) string {
	return fmt.Sprintf("%0x", sha256.Sum224(data))
}

func SHA224FromFile(file string) (string, error) {
	in, err := os.OpenFile(file, os.O_RDONLY, 0644)
	if err != nil {
		return "", err
	}

	defer in.Close()

	return SHA224FromIO(in)
}

func SHA224FromIO(in io.Reader) (string, error) {
	c := sha256.New224()

	if _, err := io.CopyBuffer(c, in, make([]byte, 4096)); err != nil {
		return "", err
	}

	return fmt.Sprintf("%0x", c.Sum(nil)), nil
}

func SHA224FromJSON[T interface{}](a T) (string, error) {
	d, err := json.Marshal(a)
	if err != nil {
		return "", err
	}

	return SHA224(d), nil
}
