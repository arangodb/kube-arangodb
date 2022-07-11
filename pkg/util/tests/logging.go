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
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"

	"github.com/arangodb/kube-arangodb/pkg/logging"
)

type LogScanner interface {
	Factory() logging.Factory

	Get(timeout time.Duration) (string, bool)
	GetData(t *testing.T, timeout time.Duration, obj interface{}) bool
}

type logScanner struct {
	factory logging.Factory

	in <-chan string
}

func (l *logScanner) GetData(t *testing.T, timeout time.Duration, obj interface{}) bool {
	if s, ok := l.Get(timeout); !ok {
		return false
	} else {
		require.NoError(t, json.Unmarshal([]byte(s), obj))
		return true
	}
}

func (l *logScanner) Factory() logging.Factory {
	return l.factory
}

func (l *logScanner) Get(timeout time.Duration) (string, bool) {
	t := time.NewTicker(timeout)
	defer t.Stop()

	for {
		select {
		case text := <-l.in:
			return text, true
		case <-t.C:
			return "", false
		}
	}
}

func WithLogScanner(t *testing.T, name string, in func(t *testing.T, s LogScanner)) {
	t.Run(name, func(t *testing.T) {
		b := bytes.NewBuffer(nil)
		l := zerolog.New(b)
		f := logging.NewFactory(l)

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

		in(t, &logScanner{
			factory: f,
			in:      out,
		})
	})
}
