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

package inventory

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	ugrpc "github.com/arangodb/kube-arangodb/pkg/util/grpc"
)

func testAsItemValue[T any](t *testing.T, in T, ok bool) {
	t.Run(reflect.TypeFor[T]().String(), func(t *testing.T) {
		if ok {
			v, err := AsItemValue[T](in)
			require.NoError(t, err)
			require.NotNil(t, v)

			o := &ItemValue{Value: v}

			z, err := ugrpc.Marshal(o)
			require.NoError(t, err)

			t.Log(string(z))

			q, err := ugrpc.Unmarshal[*ItemValue](z)
			require.NoError(t, err)

			require.EqualValues(t, v, q.GetValue())

			tz, err := o.Type()
			require.NoError(t, err)

			require.Equal(t, reflect.TypeFor[T](), tz)
		} else {
			_, err := AsItemValue[T](in)
			require.Error(t, err)
			require.EqualError(t, err, fmt.Sprintf("not supported type: %T", in))
		}
	})
}

func Test_AsItemValue(t *testing.T) {
	testAsItemValue(t, "test", true)
	testAsItemValue[int32](t, 1, true)
	testAsItemValue[int64](t, 1, true)
	testAsItemValue[float32](t, 1.55, true)
	testAsItemValue[time.Time](t, time.Now(), true)
	testAsItemValue[time.Duration](t, time.Hour, true)
	testAsItemValue[float64](t, 1.6, false)
}
