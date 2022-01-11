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

package util

import (
	"crypto/sha256"
	"fmt"

	"k8s.io/apimachinery/pkg/util/json"
)

func SHA256FromString(data string) string {
	return SHA256([]byte(data))
}

func SHA256(data []byte) string {
	return fmt.Sprintf("%0x", sha256.Sum256(data))
}

func SHA256FromJSON(a interface{}) (string, error) {
	d, err := json.Marshal(a)
	if err != nil {
		return "", err
	}

	return SHA256(d), nil
}

func CompareJSON(a, b interface{}) (bool, error) {
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
