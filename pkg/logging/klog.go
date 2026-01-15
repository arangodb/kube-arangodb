//
// DISCLAIMER
//
// Copyright 2026 ArangoDB GmbH, Cologne, Germany
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
	"fmt"

	"github.com/go-logr/logr"
	"k8s.io/klog/v2"
)

var _ klog.LogSink = &klogSink{}

func NewKLogSink(l Logger) klog.LogSink {
	return &klogSink{
		l:    l,
		name: l.Name(),
	}
}

type klogSink struct {
	l     Logger
	name  string
	depth int
}

func levelToLoggerLevel(level int) Level {
	if level > 6 {
		return Trace
	}

	if level > 2 {
		return Debug
	}

	return Info
}

func (k *klogSink) Init(info logr.RuntimeInfo) {
	k.depth = info.CallDepth + 2
}

func (k *klogSink) Enabled(level int) bool {
	return levelToLoggerLevel(level) >= Level(k.l.Logger().GetLevel())
}

func (k *klogSink) Info(level int, msg string, keysAndValues ...any) {
	k.l.Str("klogname", k.name).LevelOutput(levelToLoggerLevel(level), msg, keysAndValues...)
}

func (k *klogSink) Error(err error, msg string, keysAndValues ...any) {
	k.l.Err(err).Str("klogname", k.name).Error(msg, keysAndValues...)
}

func (k *klogSink) WithValues(keysAndValues ...any) logr.LogSink {
	return &klogSink{
		l:     k.l.Fields(keysAndValues),
		name:  k.name,
		depth: k.depth,
	}
}

func (k *klogSink) WithName(name string) logr.LogSink {
	return &klogSink{
		l:     k.l,
		name:  fmt.Sprintf("%s.%s", k.name, name),
		depth: k.depth,
	}
}
