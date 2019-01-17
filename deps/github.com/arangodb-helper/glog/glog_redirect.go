//
// DISCLAIMER
//
// Copyright 2018 ArangoDB GmbH, Cologne, Germany
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
// Author Ewout Prangsma
//

//
// Note: This code is added to the standard glog package.
//       It has to be here because it needs package level
//       access to some members.
//       Do not remove this when updating the vendored glog package!
//

package glog

import "strings"

type LogLevel int

const (
	// Make sure these constants end up having the same indexes as the severity constants
	LogLevelInfo LogLevel = iota
	LogLevelWarning
	LogLevelError
	LogLevelFatal
)

// redirectWriter wraps a callback that is called when data is written to it.
type redirectWriter struct {
	cb    func(level LogLevel, message string)
	level LogLevel
}

func (w *redirectWriter) Flush() error {
	return nil
}

func (w *redirectWriter) Sync() error {
	return nil
}

func (w *redirectWriter) Write(p []byte) (n int, err error) {
	msg := string(p)
	if msg[len(msg)-1] == '\n' {
		msg = msg[:len(msg)-1]
	}
	if idx := strings.IndexByte(msg, ']'); idx > 0 {
		msg = strings.TrimSpace(msg[idx+1:])
	}
	w.cb(w.level, msg)
	return len(p), nil
}

// RedirectOutput redirects output of the given logging to the given callback.
func (l *loggingT) RedirectOutput(cb func(level LogLevel, message string)) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.toStderr = false
	l.alsoToStderr = false
	for i := range logging.file {
		logging.file[i] = &redirectWriter{
			cb:    cb,
			level: LogLevel(i),
		}
	}
	return
}

// RedirectOutput redirects output of thestandard logging to the given callback.
func RedirectOutput(cb func(level LogLevel, message string)) {
	logging.RedirectOutput(cb)
}

