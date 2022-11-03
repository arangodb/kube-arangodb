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

package k8sutil

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/arangodb/kube-arangodb/pkg/util/tests"
	"github.com/arangodb/kube-arangodb/pkg/util/timer"
)

type mockInformer struct {
	delay time.Duration
	done  bool
}

func (m *mockInformer) WaitForCacheSync(stopCh <-chan struct{}) map[reflect.Type]bool {
	defer func() {
		m.done = true
	}()
	c := timer.After(m.delay)

	select {
	case <-c:
	case <-stopCh:
	}

	return nil
}

func MockInformer(delay time.Duration) *mockInformer {
	return &mockInformer{delay: delay}
}

func Test_WaitForInformers(t *testing.T) {
	t.Run("Delayed Sync", func(t *testing.T) {
		a := MockInformer(time.Second)

		stop := make(chan struct{})

		n := tests.DurationBetween()
		WaitForInformers(stop, 5*time.Second, a)

		n(t, time.Second, 0.1)
		require.True(t, a.done)
	})

	t.Run("Instant Sync", func(t *testing.T) {
		a := MockInformer(time.Millisecond)

		stop := make(chan struct{})

		n := tests.DurationBetween()
		WaitForInformers(stop, 5*time.Second, a)

		n(t, time.Millisecond, 5)
		require.True(t, a.done)
	})

	t.Run("Timeout Sync", func(t *testing.T) {
		a := MockInformer(10 * time.Second)

		stop := make(chan struct{})

		n := tests.DurationBetween()
		WaitForInformers(stop, 5*time.Second, a)

		n(t, 5*time.Second, 0.05)
		require.False(t, a.done)
	})
}
