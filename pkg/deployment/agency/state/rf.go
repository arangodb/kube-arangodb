//
// DISCLAIMER
//
// Copyright 2016-2023 ArangoDB GmbH, Cologne, Germany
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

package state

import (
	"encoding/json"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	UnknownReplicationFactor   ReplicationFactor = -1000
	SatelliteReplicationFactor ReplicationFactor = -100
)

type ReplicationFactor int

func (r *ReplicationFactor) IsNil() bool {
	return r == nil
}

func (r *ReplicationFactor) IsUnknown() bool {
	if r == nil {
		return false
	}

	return *r == UnknownReplicationFactor
}

func (r *ReplicationFactor) IsSatellite() bool {
	if r == nil {
		return false
	}

	return *r == SatelliteReplicationFactor
}

func (r *ReplicationFactor) UnmarshalJSON(bytes []byte) error {
	var i intstr.IntOrString

	if err := json.Unmarshal(bytes, &i); err != nil {
		return err
	}

	switch i.Type {
	case intstr.Int:
		*r = ReplicationFactor(i.IntVal)
		return nil
	case intstr.String:
		switch i.StrVal {
		case "satellite":
			*r = SatelliteReplicationFactor
			return nil
		default:
			*r = UnknownReplicationFactor
			return nil
		}
	}

	return errors.Errorf("Unable to parse value")
}
