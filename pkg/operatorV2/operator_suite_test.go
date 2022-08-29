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

package operator

import (
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/arangodb/kube-arangodb/pkg/operatorV2/operation"
)

const (
	charset = "abcdefghijklmnoprstuvwxyz"
)

func randomString(len int) string {
	r := make([]byte, len)

	for id := range r {
		r[id] = charset[rand.Intn(len)]
	}

	return string(r)
}

type mockHandler struct {
	name string

	wrapperHandle       func(item operation.Item) error
	wrapperCanBeHandled func(item operation.Item) bool
}

func (m *mockHandler) Name() string {
	return m.name
}

func (m *mockHandler) Handle(item operation.Item) error {
	return m.wrapperHandle(item)
}

func (m *mockHandler) CanBeHandled(item operation.Item) bool {
	return m.wrapperCanBeHandled(item)
}

func newMockHandler(name string, handle func(item operation.Item) error, canBeHandled func(item operation.Item) bool) Handler {
	return &mockHandler{
		name:                name,
		wrapperHandle:       handle,
		wrapperCanBeHandled: canBeHandled,
	}
}

func randomItem() operation.Item {
	return operation.Item{
		Operation: operation.Add,

		Group:   randomString(5),
		Version: randomString(2),
		Kind:    randomString(5),

		Namespace: randomString(5),
		Name:      randomString(5),
	}
}

func mockSimpleObjectFunc(name string, canBeHandled func(item operation.Item) bool) (Handler, chan operation.Item) {
	c := make(chan operation.Item, 1024)
	return newMockHandler(name,
		func(item operation.Item) error {
			c <- item
			return nil
		},
		canBeHandled), c
}

func mockSimpleObject(name string, canBeHandled bool) (Handler, chan operation.Item) {
	return mockSimpleObjectFunc(name, func(item operation.Item) bool {
		return canBeHandled
	})
}

func waitForItems(t *testing.T, i <-chan operation.Item, expectedSize int) []operation.Item {
	tmout := time.NewTimer(time.Second)
	defer tmout.Stop()
	received := make([]operation.Item, 0, expectedSize)
	for {
		select {
		case item := <-i:
			received = append(received, item)
			if len(received) == expectedSize {
				return received
			}
		case <-tmout.C:
			require.Fail(t, "Timeout")
		}
	}
}
