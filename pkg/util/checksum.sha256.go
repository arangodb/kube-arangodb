//
// DISCLAIMER
//
// Copyright 2025-2026 ArangoDB GmbH, Cologne, Germany
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
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	goStrings "strings"
)

func SHA256FromExtract[T any](extract func(T) string, obj ...T) string {
	return SHA256FromStringArray(goStrings.Join(FormatList(obj, extract), "|"))
}

func SHA256FromExtractMap[K comparable, T any](extract func(K, T) string, obj map[K]T) string {
	return SHA256FromStringArray(goStrings.Join(ExtractMap(obj, extract), "|"))
}

func SHA256FromHashStringMap[T Hash](data map[string]T) string {
	return SHA256FromExtractMap(func(k string, t T) string {
		return fmt.Sprintf("%s:%s", k, t.Hash())
	}, data)
}

func SHA256FromHashArray[T Hash](data []T) string {
	return SHA256FromExtract(func(t T) string {
		return t.Hash()
	}, data...)
}

func SHA256FromNonEmptyStringArray(data ...string) string {
	return SHA256FromFilteredStringArray(func(in string) bool {
		return in != ""
	}, data...)
}

func SHA256FromFilteredStringArray(filter func(in string) bool, data ...string) string {
	return SHA256FromStringArray(FilterList(data, filter)...)
}

func SHA256FromStringArray(data ...string) string {
	return SHA256FromString(goStrings.Join(data, "|"))
}

func SHA256FromStringMap(data map[string]string) string {
	return SHA256FromExtract(func(t KV[string, string]) string {
		return fmt.Sprintf("%s:%s", t.K, SHA256FromString(t.V))
	}, ExtractWithSort(data, func(i, j string) bool {
		return i < j
	})...)
}

func SHA256FromString(data string) string {
	return SHA256([]byte(data))
}

func TrimSpaceSHA256(data []byte) string {
	return SHA256(bytes.TrimSpace(data))
}

func SHA256(data []byte) string {
	return fmt.Sprintf("%0x", sha256.Sum256(data))
}

func SHA256FromFile(file string) (string, error) {
	in, err := os.OpenFile(file, os.O_RDONLY, 0644)
	if err != nil {
		return "", err
	}

	defer in.Close()

	return SHA256FromIO(in)
}

func SHA256FromIO(in io.Reader) (string, error) {
	c := sha256.New()

	if _, err := io.CopyBuffer(c, in, make([]byte, 4096)); err != nil {
		return "", err
	}

	return fmt.Sprintf("%0x", c.Sum(nil)), nil
}

func SHA256FromJSON[T interface{}](a T) (string, error) {
	d, err := json.Marshal(a)
	if err != nil {
		return "", err
	}

	return SHA256(d), nil
}
