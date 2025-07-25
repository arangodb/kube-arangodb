//
// DISCLAIMER
//
// Copyright 2023-2025 ArangoDB GmbH, Cologne, Germany
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

import "sync"

type Close interface {
	Close() error
}

type closeOnce struct {
	lock sync.Mutex

	close Close

	closed bool
}

func (c *closeOnce) Close() error {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.closed {
		return nil
	}

	c.closed = true

	return c.close.Close()
}

func CloseOnce(c Close) Close {
	return &closeOnce{
		close: c,
	}
}

func IsChannelClosed[T any](in <-chan T) bool {
	select {
	case <-in:
		return true
	default:
		return false
	}
}
