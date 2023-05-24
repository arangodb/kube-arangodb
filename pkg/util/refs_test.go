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

package util

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	core "k8s.io/api/core/v1"
)

func testRefs[T interface{}](t *testing.T, in, v1, def T) {
	tp := reflect.TypeOf(in)
	t.Run(tp.String(), func(t *testing.T) {
		t.Run("New", func(t *testing.T) {
			n := NewType[T](in)

			tn := reflect.ValueOf(n)

			require.Equal(t, reflect.Pointer, tn.Kind())
			require.Equal(t, tp.Kind(), tn.Elem().Kind())
			require.Equal(t, in, tn.Elem().Interface())
		})
		t.Run("NewOrNil", func(t *testing.T) {
			t.Run("Nil", func(t *testing.T) {
				n := NewTypeOrNil[T](nil)

				tn := reflect.ValueOf(n)

				require.Equal(t, reflect.Pointer, tn.Kind())
				require.True(t, tn.IsNil())
			})
			t.Run("Value", func(t *testing.T) {
				n := NewTypeOrNil[T](&in)

				tn := reflect.ValueOf(n)

				require.Equal(t, reflect.Pointer, tn.Kind())
				require.False(t, tn.IsNil())
				require.Equal(t, tp.Kind(), tn.Elem().Kind())
				require.Equal(t, in, tn.Elem().Interface())
			})
		})
		t.Run("Default", func(t *testing.T) {
			t.Run("Ensure default", func(t *testing.T) {
				var val T
				require.Equal(t, def, val)
			})

			t.Run("With Input", func(t *testing.T) {
				n := TypeOrDefault[T](&in)

				tn := reflect.ValueOf(n)

				require.Equal(t, tp.Kind(), tn.Kind())
				require.Equal(t, in, tn.Interface())
			})

			t.Run("With Input & Default", func(t *testing.T) {
				n := TypeOrDefault[T](&in, v1)

				tn := reflect.ValueOf(n)

				require.Equal(t, tp.Kind(), tn.Kind())
				require.Equal(t, in, tn.Interface())
			})

			t.Run("With Nil & Default", func(t *testing.T) {
				n := TypeOrDefault[T](nil, v1)

				tn := reflect.ValueOf(n)

				require.Equal(t, tp.Kind(), tn.Kind())
				require.Equal(t, v1, tn.Interface())
			})

			t.Run("With Nil", func(t *testing.T) {
				n := TypeOrDefault[T](nil)

				tn := reflect.ValueOf(n)

				require.Equal(t, tp.Kind(), tn.Kind())
				require.Equal(t, def, tn.Interface())
			})
		})
	})
}

func Test_Refs(t *testing.T) {
	testRefs[string](t, "test", "otherValue", "")
	testRefs[int](t, 777, 555, 0)
	testRefs[int32](t, 777, 555, 0)
	testRefs[int64](t, 777, 555, 0)
	testRefs[uint16](t, 777, 555, 0)
	testRefs[bool](t, false, true, false)
	testRefs[time.Duration](t, time.Duration(500), time.Duration(1500), time.Duration(0))
	testRefs[core.PullPolicy](t, core.PullAlways, core.PullNever, "")
}
