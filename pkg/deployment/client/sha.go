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

package client

import (
	"strings"
)

type Sha string

func (s Sha) String() string {
	return string(s)
}

func (s Sha) Type() string {
	z := strings.Split(s.String(), ":")
	if len(z) < 2 {
		return "sha256"
	}
	return z[0]
}

func (s Sha) Checksum() string {
	z := strings.Split(s.String(), ":")
	if len(z) < 2 {
		return z[0]
	}
	return z[1]
}

type Entry struct {
	Sha256    *Sha `json:"sha256,omitempty"`
	Sha256Old *Sha `json:"SHA256,omitempty"`
}

func (e *Entry) GetSHA() Sha {
	if e == nil {
		return ""
	}

	if e.Sha256 != nil {
		return *e.Sha256
	}
	if e.Sha256Old != nil {
		return *e.Sha256Old
	}

	return ""
}

type Entries []Entry

func (e Entries) KeysPresent(m map[string][]byte) bool {
	if len(e) != len(m) {
		return false
	}

	for key := range m {
		ok := false
		for _, entry := range e {
			if entry.GetSHA().Checksum() == key {
				ok = true
				break
			}
		}
		if !ok {
			return false
		}
	}

	return true
}

func (e Entries) Contains(s string) bool {
	for _, entry := range e {
		if entry.GetSHA().String() == s {
			return true
		}
	}
	return false
}
