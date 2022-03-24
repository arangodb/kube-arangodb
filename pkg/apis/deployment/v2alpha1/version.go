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

package v2alpha1

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

var _ json.Marshaler = Version{}
var _ json.Unmarshaler = &Version{}

type Version struct {
	Major int `json:"major"`
	Minor int `json:"minor"`
	Patch int `json:"patch"`
	ID    int `json:"ID,omitempty"`
}

func (v Version) Compare(b Version) int {
	if v.Major > b.Major {
		return 1
	} else if v.Major < b.Major {
		return -1
	}

	if v.Minor > b.Minor {
		return 1
	} else if v.Minor < b.Minor {
		return -1
	}

	if v.Patch > b.Patch {
		return 1
	} else if v.Patch < b.Patch {
		return -1
	}

	if v.ID < b.ID {
		return 1
	} else if b.ID < v.ID {
		return -1
	}

	return 0
}

func (v *Version) Equal(b *Version) bool {
	if v == nil && b == nil {
		return true
	}
	if v == nil || b == nil {
		return true
	}
	return v.Major == b.Major && v.Minor == b.Minor && v.Patch == b.Patch && v.ID == b.ID
}

func (v *Version) UnmarshalJSON(bytes []byte) error {
	if v == nil {
		return errors.Errorf("Nil version provided")
	}

	var s string

	if err := json.Unmarshal(bytes, &s); err != nil {
		*v = Version{
			Major: 0,
			Minor: 0,
			Patch: 0,
		}
		return nil
	}

	z := strings.Split(s, ".")

	i := make([]int, len(z))

	for id, z := range z {
		if q, err := strconv.Atoi(z); err != nil {
			*v = Version{
				Major: 0,
				Minor: 0,
				Patch: 0,
			}
			return nil
		} else {
			i[id] = q
		}
	}
	switch l := len(i); l {
	case 1:
		var n Version

		n.Major = i[0]

		*v = n
	case 2:
		var n Version

		n.Major = i[0]
		n.Minor = i[1]

		*v = n
	case 3:
		var n Version

		n.Major = i[0]
		n.Minor = i[1]
		n.Patch = i[2]

		*v = n
	case 4:
		var n Version

		n.Major = i[0]
		n.Minor = i[1]
		n.Patch = i[2]
		n.ID = i[3]

		*v = n
	default:
		*v = Version{
			Major: 0,
			Minor: 0,
			Patch: 0,
		}
		return nil
	}

	return nil
}

func (v Version) MarshalJSON() ([]byte, error) {
	if v.ID == 0 {
		return json.Marshal(fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch))
	}

	return json.Marshal(fmt.Sprintf("%d.%d.%d.%d", v.Major, v.Minor, v.Patch, v.ID))
}
