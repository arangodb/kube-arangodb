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
	"reflect"
	"time"

	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

func Produce[T any](out chan<- *Item, key string, dimensions map[string]string, v T) error {
	val, err := AsItemValue[T](v)
	if err != nil {
		return err
	}

	q := &Item{
		Type:       key,
		Dimensions: dimensions,
		Value:      &ItemValue{Value: val},
	}

	if err := q.Validate(); err != nil {
		return err
	}

	out <- q

	return nil
}

func (i *ItemValue) Type() (reflect.Type, error) {
	switch i.Value.(type) {
	case *ItemValue_Str:
		return reflect.TypeFor[string](), nil
	case *ItemValue_Num:
		return reflect.TypeFor[int32](), nil
	case *ItemValue_LongNum:
		return reflect.TypeFor[int64](), nil
	case *ItemValue_Bool:
		return reflect.TypeFor[bool](), nil
	case *ItemValue_Duration:
		return reflect.TypeFor[time.Duration](), nil
	case *ItemValue_Dec:
		return reflect.TypeFor[float32](), nil
	case *ItemValue_Time:
		return reflect.TypeFor[time.Time](), nil
	default:
		return nil, errors.Errorf("Unknown Type: %T", i.Value)
	}
}

func AsItemValue[T any](in T) (isItemValue_Value, error) {
	v := reflect.ValueOf(in).Interface()

	if reflect.TypeFor[string]() == reflect.TypeFor[T]() {
		t, ok := v.(string)
		if !ok {
			return nil, errors.Errorf("expected type %T, got %T", t, v)
		}
		return &ItemValue_Str{Str: t}, nil
	}

	if reflect.TypeFor[float32]() == reflect.TypeFor[T]() {
		t, ok := v.(float32)
		if !ok {
			return nil, errors.Errorf("expected type %T, got %T", t, v)
		}
		return &ItemValue_Dec{Dec: t}, nil
	}

	if reflect.TypeFor[int32]() == reflect.TypeFor[T]() {
		t, ok := v.(int32)
		if !ok {
			return nil, errors.Errorf("expected type %T, got %T", t, v)
		}
		return &ItemValue_Num{Num: t}, nil
	}

	if reflect.TypeFor[int64]() == reflect.TypeFor[T]() {
		t, ok := v.(int64)
		if !ok {
			return nil, errors.Errorf("expected type %T, got %T", t, v)
		}
		return &ItemValue_LongNum{LongNum: t}, nil
	}

	if reflect.TypeFor[time.Time]() == reflect.TypeFor[T]() {
		t, ok := v.(time.Time)
		if !ok {
			return nil, errors.Errorf("expected type %T, got %T", t, v)
		}
		return &ItemValue_Time{Time: timestamppb.New(t.Truncate(time.Second))}, nil
	}

	if reflect.TypeFor[time.Duration]() == reflect.TypeFor[T]() {
		t, ok := v.(time.Duration)
		if !ok {
			return nil, errors.Errorf("expected type %T, got %T", t, v)
		}
		return &ItemValue_Duration{Duration: durationpb.New(t)}, nil
	}

	return nil, errors.Errorf("not supported type: %T", in)
}
