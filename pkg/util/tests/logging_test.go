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

package tests

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/arangodb/kube-arangodb/pkg/logging"
)

func Test_Logger(t *testing.T) {
	WithLogScanner(t, "Logger", func(t *testing.T, s LogScanner) {
		f := s.Factory()
		q := f.Get("foo")

		t.Run("Run on unregistered logger", func(t *testing.T) {
			q.Info("Data")

			_, ok := s.Get(100 * time.Millisecond)
			require.False(t, ok)
		})

		t.Run("Register logger", func(t *testing.T) {
			f.RegisterLogger("foo", logging.Info)
			f.RegisterLogger("bar", logging.Info)
		})

		t.Run("Run on registered logger", func(t *testing.T) {
			q.Info("Data")

			_, ok := s.Get(100 * time.Millisecond)
			require.True(t, ok)
		})

		t.Run("Run on too low log level logger", func(t *testing.T) {
			q.Debug("Data")

			_, ok := s.Get(100 * time.Millisecond)
			require.False(t, ok)
		})

		t.Run("Change log level", func(t *testing.T) {
			f.ApplyLogLevels(map[string]logging.Level{
				"foo": logging.Debug,
			})

			q.Debug("Data")

			_, ok := s.Get(100 * time.Millisecond)
			require.True(t, ok)

			require.Equal(t, logging.Debug, f.LogLevels()["foo"])
		})

		t.Run("Change all log levels", func(t *testing.T) {
			f.ApplyLogLevels(map[string]logging.Level{
				"all": logging.Info,
			})

			q.Debug("Data")

			_, ok := s.Get(100 * time.Millisecond)
			require.False(t, ok)

			require.Equal(t, logging.Info, f.LogLevels()["foo"])
			require.Equal(t, logging.Info, f.LogLevels()["bar"])
		})

		t.Run("Change invalid level", func(t *testing.T) {
			f.ApplyLogLevels(map[string]logging.Level{
				"invalid": logging.Info,
			})

			q.Debug("Data")

			_, ok := s.Get(100 * time.Millisecond)
			require.False(t, ok)

			require.Equal(t, logging.Info, f.LogLevels()["foo"])
		})

		t.Run("Change all log levels with override", func(t *testing.T) {
			f.ApplyLogLevels(map[string]logging.Level{
				"all": logging.Debug,
				"foo": logging.Info,
			})

			q.Debug("Data")

			_, ok := s.Get(100 * time.Millisecond)
			require.False(t, ok)

			levels := f.LogLevels()
			require.Equal(t, logging.Info, f.LogLevels()["foo"])
			require.Equal(t, logging.Debug, levels["bar"])
		})
	})
}
