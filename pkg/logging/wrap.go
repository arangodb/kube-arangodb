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

package logging

import (
	"time"

	"github.com/rs/zerolog"
)

type WrapObj interface {
	WrapLogger(in *zerolog.Event) *zerolog.Event
}

type Wrap func(in *zerolog.Event) *zerolog.Event

func Str(key, value string) Wrap {
	return func(in *zerolog.Event) *zerolog.Event {
		return in.Str(key, value)
	}
}

func Strs(key string, values ...string) Wrap {
	return func(in *zerolog.Event) *zerolog.Event {
		return in.Strs(key, values)
	}
}

func Bool(key string, i bool) Wrap {
	return func(in *zerolog.Event) *zerolog.Event {
		return in.Bool(key, i)
	}
}

func Int32(key string, i int32) Wrap {
	return func(in *zerolog.Event) *zerolog.Event {
		return in.Int32(key, i)
	}
}

func Int64(key string, i int64) Wrap {
	return func(in *zerolog.Event) *zerolog.Event {
		return in.Int64(key, i)
	}
}

func Time(key string, time time.Time) Wrap {
	return func(in *zerolog.Event) *zerolog.Event {
		return in.Time(key, time)
	}
}

func SinceStart(key string, start time.Time) Wrap {
	return func(in *zerolog.Event) *zerolog.Event {
		return in.Str(key, time.Since(start).String())
	}
}

func Int(key string, i int) Wrap {
	return func(in *zerolog.Event) *zerolog.Event {
		return in.Int(key, i)
	}
}

func Interface(key string, i interface{}) Wrap {
	return func(in *zerolog.Event) *zerolog.Event {
		return in.Interface(key, i)
	}
}

func Err(err error) Wrap {
	return func(in *zerolog.Event) *zerolog.Event {
		return in.Err(err)
	}
}

func Dur(key string, dur time.Duration) Wrap {
	return func(in *zerolog.Event) *zerolog.Event {
		return in.Dur(key, dur)
	}
}
