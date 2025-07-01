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

package closer

import shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"

type MultiCloser interface {
	Close
	With(closers ...Close) MultiCloser
}

func NewMultiCloser(closers ...Close) MultiCloser {
	return multiCloser(closers)
}

type multiCloser []Close

func (m multiCloser) With(closers ...Close) MultiCloser {
	r := make(multiCloser, len(m)+len(closers))
	copy(r, m)
	copy(r[len(m):], closers)
	return r
}

func (m multiCloser) Close() error {
	e := make([]error, len(m))

	for id := len(m) - 1; id >= 0; id-- {
		e[id] = m[id].Close()
	}

	return shared.WithErrors(e...)
}
