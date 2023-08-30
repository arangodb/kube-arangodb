//
// DISCLAIMER
//
// Copyright 2023 ArangoDB GmbH, Cologne, Germany
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

package http

import (
	"sort"
	"strconv"
	"strings"
)

type Headers map[string]float64

func (h Headers) Accept(headers ...string) string {
	if len(h) == 0 {
		return "identity"
	}

	mapped := map[string]float64{}

	s, sok := h["*"]

	for _, header := range headers {
		switch header {
		case "gzip", "compress", "deflate", "br", "identity":
		default:
			continue
		}
		v, ok := h[header]
		if !ok {
			if !sok {
				continue
			}
			v = s
		}
		mapped[header] = v
	}

	if len(mapped) == 0 {
		return "identity"
	}

	indexes := map[string]int{}

	for id, header := range headers {
		indexes[header] = id
	}

	returns := make([]string, 0, len(mapped))

	for k := range mapped {
		returns = append(returns, k)
	}

	sort.Slice(returns, func(i, j int) bool {
		if iv, jv := mapped[returns[i]], mapped[returns[j]]; iv == jv {
			return indexes[returns[i]] < indexes[returns[j]]
		} else {
			return iv > jv
		}
	})

	return returns[0]
}

func ParseHeaders(in ...string) Headers {
	h := Headers{}
	for _, v := range in {
		for _, el := range strings.Split(v, ",") {
			el = strings.TrimSpace(el)

			els := strings.Split(el, ";")

			if len(els) == 1 {
				h[els[0]] = 1
			} else {
				q := 1.0
				for _, part := range els[1:] {
					parts := strings.Split(part, "=")
					if len(parts) <= 1 {
						continue
					}
					if parts[0] != "q" {
						continue
					}
					v, err := strconv.ParseFloat(parts[1], 64)
					if err != nil {
						continue
					}
					q = v
				}
				h[els[0]] = q
			}
		}
	}
	return h
}
