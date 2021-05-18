//
// DISCLAIMER
//
// Copyright 2020 ArangoDB GmbH, Cologne, Germany
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
// Author Adam Janikowski
//

package k8sutil

import (
	"fmt"
	"sort"
	"strings"
)

func CreateOptionPairs(lens ...int) OptionPairs {
	l := 16

	if len(lens) > 0 {
		l = lens[0]
	}

	return make(OptionPairs, 0, l)
}

// OptionPairs list of pair builder
type OptionPairs []OptionPair

func (o *OptionPairs) Append(pairs ...OptionPair) {
	if o == nil {
		*o = pairs
		return
	}

	*o = append(*o, pairs...)
}

func (o *OptionPairs) Addf(key, format string, i ...interface{}) {
	o.Add(key, fmt.Sprintf(format, i...))
}

func (o *OptionPairs) Add(key string, value interface{}) {
	switch v := value.(type) {
	case string:
		o.add(key, v)
	case bool:
		f := "false"
		if v {
			f = "true"
		}
		o.add(key, f)
	case int:
		o.add(key, fmt.Sprintf("%d", v))
	default:
		o.add(key, fmt.Sprintf("%s", v))
	}
}

func (o *OptionPairs) add(key, value string) {
	o.Append(OptionPair{
		Key:   key,
		Value: value,
	})
}

func (o *OptionPairs) Merge(pairs ...OptionPairs) {
	for _, pair := range pairs {
		if len(pair) == 0 {
			continue
		}

		o.Append(pair...)
	}
}

func (o OptionPairs) Unique() OptionPairs {
	r := make(OptionPairs, 0, len(o))

	for _, pair := range o {
		replaced := false
		for id, existing := range r {
			if replaced {
				break
			}
			if existing.Key == pair.Key {
				r[id] = pair
				replaced = true
			}
		}

		if replaced {
			continue
		}

		r = append(r, pair)
	}

	return r
}

func (o OptionPairs) Copy() OptionPairs {
	r := make(OptionPairs, len(o))
	copy(r, o)

	return r
}

func (o OptionPairs) Sort() OptionPairs {
	sort.Slice(o, func(i, j int) bool {
		return o[i].CompareTo(o[j]) < 0
	})

	return o
}

func (o OptionPairs) AsArgs() []string {
	s := make([]string, len(o))

	for id, pair := range o {
		s[id] = pair.String()
	}

	return s
}

// OptionPair key value pair builder
type OptionPair struct {
	Key   string
	Value string
}

func (o OptionPair) String() string {
	return fmt.Sprintf("%s=%s", o.Key, o.Value)
}

// CompareTo returns -1 if o < other, 0 if o == other, 1 otherwise
func (o OptionPair) CompareTo(other OptionPair) int {
	rc := strings.Compare(o.Key, other.Key)
	if rc < 0 {
		return -1
	} else if rc > 0 {
		return 1
	}
	return strings.Compare(o.Value, other.Value)
}

func NewOptionPair(pairs ...OptionPair) OptionPairs {
	return pairs
}
