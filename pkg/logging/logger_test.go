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
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

func logScanner() (Factory, <-chan string, func()) {
	b := bytes.NewBuffer(nil)
	l := zerolog.New(b)
	f := NewFactory(l)

	out := make(chan string)

	closer := make(chan struct{})

	go func() {
		defer close(out)
		t := time.NewTicker(time.Millisecond)
		defer t.Stop()

		z := ""

		for {
			for b.Len() > 0 {
				q, _, _ := b.ReadRune()
				if q == '\n' {
					out <- z
					z = ""
				} else {
					z = z + string(q)
				}
			}

			select {
			case <-closer:
				return
			case <-t.C:
			}
		}
	}()

	return f, out, func() {
		close(closer)
	}
}

func readData(in <-chan string) (string, bool) {
	t := time.NewTimer(100 * time.Millisecond)
	defer t.Stop()

	select {
	case text := <-in:
		return text, true
	case <-t.C:
		return "", false
	}
}

func expectTimeout(t *testing.T, in <-chan string) {
	_, ok := readData(in)
	require.False(t, ok, "Data should be not present")
}

func expectData(t *testing.T, in <-chan string) {
	s, ok := readData(in)
	require.True(t, ok, "Data should be present")

	var q map[string]string

	require.NoError(t, json.Unmarshal([]byte(s), &q))
}

func Test_Logger(t *testing.T) {
	f, data, c := logScanner()
	defer c()

	q := f.Get("foo")

	t.Run("Run on unregistered logger", func(t *testing.T) {
		q.Info("Data")

		expectTimeout(t, data)
	})

	t.Run("Register logger", func(t *testing.T) {
		f.RegisterLogger("foo", Info)
	})

	t.Run("Run on registered logger", func(t *testing.T) {
		q.Info("Data")

		expectData(t, data)
	})

	t.Run("Run on too low log level logger", func(t *testing.T) {
		q.Debug("Data")

		expectTimeout(t, data)
	})

	t.Run("Change log level", func(t *testing.T) {
		f.ApplyLogLevels(map[string]Level{
			"foo": Debug,
		})

		q.Debug("Data")

		expectData(t, data)

		require.Equal(t, Debug, f.LogLevels()["foo"])
	})

	t.Run("Change all log levels", func(t *testing.T) {
		f.ApplyLogLevels(map[string]Level{
			"all": Info,
		})

		q.Debug("Data")

		expectTimeout(t, data)

		require.Equal(t, Info, f.LogLevels()["foo"])
	})

	t.Run("Change invalid level", func(t *testing.T) {
		f.ApplyLogLevels(map[string]Level{
			"invalid": Info,
		})

		q.Debug("Data")

		expectTimeout(t, data)

		require.Equal(t, Info, f.LogLevels()["foo"])
	})

	t.Run("Change all log levels with override", func(t *testing.T) {
		f.ApplyLogLevels(map[string]Level{
			"all": Debug,
			"foo": Info,
		})

		q.Debug("Data")

		expectTimeout(t, data)

		require.Equal(t, Info, f.LogLevels()["foo"])
	})
}
