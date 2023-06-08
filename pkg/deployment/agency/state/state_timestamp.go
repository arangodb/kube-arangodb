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
	"time"

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

type Timestamp struct {
	hash   string
	data   string
	time   time.Time
	parsed bool
	exists bool
}

func (s *Timestamp) Hash() string {
	if s == nil {
		return util.SHA256FromString("")
	}

	return s.hash
}

func (s *Timestamp) Exists() bool {
	if s == nil {
		return false
	}

	return s.exists
}

func (s *Timestamp) Time() (time.Time, bool) {
	if s == nil {
		return time.Time{}, false
	}

	if !s.parsed || !s.exists {
		return time.Time{}, false
	}

	return s.time, true
}

func (s *Timestamp) UnmarshalJSON(bytes []byte) error {
	if s == nil {
		return errors.Newf("Object is nil")
	}

	var t string

	if err := json.Unmarshal(bytes, &t); err != nil {
		return err
	}

	*s = unmarshalJSONStateTimestamp(t)

	return nil
}

func unmarshalJSONStateTimestamp(s string) Timestamp {
	var ts = Timestamp{
		hash:   util.SHA256([]byte(s)),
		data:   s,
		exists: true,
	}

	t, ok := util.ParseAgencyTime(s)
	if !ok {
		ts.parsed = false
		return ts
	}

	ts.time = t
	ts.parsed = true
	ts.hash = util.SHA256FromString(s)

	return ts
}
