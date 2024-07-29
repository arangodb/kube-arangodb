//
// DISCLAIMER
//
// Copyright 2016-2024 ArangoDB GmbH, Cologne, Germany
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
	"crypto/md5"
	"crypto/sha256"
	"fmt"

	"k8s.io/apimachinery/pkg/util/json"

	"github.com/arangodb/kube-arangodb/pkg/util/strings"
)

type Hash interface {
	Hash() string
}

func SHA256FromExtract[T any](extract func(T) string, obj ...T) string {
	return SHA256FromStringArray(strings.Join(FormatList(obj, extract), "|"))
}

func SHA256FromHashArray[T Hash](data []T) string {
	return SHA256FromExtract(func(t T) string {
		return t.Hash()
	}, data...)
}

func SHA256FromStringArray(data ...string) string {
	return SHA256FromString(strings.Join(data, "|"))
}

func SHA256FromString(data string) string {
	return SHA256([]byte(data))
}

func SHA256(data []byte) string {
	return fmt.Sprintf("%0x", sha256.Sum256(data))
}

func MD5FromString(data string) string {
	return MD5([]byte(data))
}

func MD5(data []byte) string {
	return fmt.Sprintf("%0x", md5.Sum(data))
}

func SHA256FromJSON[T interface{}](a T) (string, error) {
	d, err := json.Marshal(a)
	if err != nil {
		return "", err
	}

	return SHA256(d), nil
}

func CompareJSON[T interface{}](a, b T) (bool, error) {
	ad, err := SHA256FromJSON(a)
	if err != nil {
		return false, err
	}
	bd, err := SHA256FromJSON(b)
	if err != nil {
		return false, err
	}

	return ad == bd, nil
}

func CompareJSONP[T interface{}](a, b *T) (bool, error) {
	var a1, b1 T

	if a != nil {
		a1 = *a
	}

	if b != nil {
		b1 = *b
	}

	return CompareJSON(a1, b1)
}
