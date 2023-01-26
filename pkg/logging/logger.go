//
// DISCLAIMER
//
// Copyright 2016-2023 ArangoDB GmbH, Cologne, Germany
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
	"io"
	"sort"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const TopicAll = "all"

type Factory interface {
	Get(name string) Logger

	LogLevels() map[string]Level
	ApplyLogLevels(in map[string]Level)
	SetRoot(log zerolog.Logger)

	RegisterLogger(name string, level Level) bool
	RegisterAndGetLogger(name string, level Level) Logger

	RegisterWrappers(w ...Wrap)

	Names() []string
}

func NewDefaultFactory() Factory {
	return NewFactory(log.Logger)
}

func NewFactory(root zerolog.Logger) Factory {
	return &factory{
		root:     root,
		loggers:  map[string]*zerolog.Logger{},
		defaults: map[string]Level{},
		levels:   map[string]Level{},
	}
}

type factory struct {
	lock sync.Mutex

	root zerolog.Logger

	wrappers []Wrap

	loggers map[string]*zerolog.Logger

	defaults map[string]Level
	levels   map[string]Level
}

func (f *factory) Names() []string {
	z := f.LogLevels()

	r := make([]string, 0, len(z))

	for k := range z {
		r = append(r, k)
	}

	sort.Strings(r)

	return r
}

func (f *factory) RegisterWrappers(w ...Wrap) {
	f.lock.Lock()
	defer f.lock.Unlock()

	f.wrappers = append(f.wrappers, w...)
}

func (f *factory) RegisterAndGetLogger(name string, level Level) Logger {
	f.RegisterLogger(name, level)
	return f.Get(name)
}

func (f *factory) SetRoot(log zerolog.Logger) {
	f.lock.Lock()
	defer f.lock.Unlock()

	f.root = log

	for k := range f.loggers {
		l := log.Level(f.loggers[k].GetLevel())
		f.loggers[k] = &l
	}
}

func (f *factory) ApplyLogLevels(in map[string]Level) {
	f.lock.Lock()
	defer f.lock.Unlock()

	z := make(map[string]Level, len(in))

	for k, v := range in {
		z[k] = v
	}

	f.levels = z

	for k := range f.loggers {
		f.applyForLogger(k)
	}
}

func (f *factory) applyForLogger(name string) {
	if def, ok := f.levels[TopicAll]; ok {
		if ov, ok := f.levels[name]; ok {
			// override on logger level
			l := f.root.Level(zerolog.Level(ov))
			f.loggers[name] = &l
		} else {
			// override on global level
			l := f.root.Level(zerolog.Level(def))
			f.loggers[name] = &l
		}
	} else {
		if ov, ok := f.levels[name]; ok {
			// override on logger level
			l := f.root.Level(zerolog.Level(ov))
			f.loggers[name] = &l
		} else {
			// override on global level
			l := f.root.Level(zerolog.Level(f.defaults[name]))
			f.loggers[name] = &l
		}
	}
}

func (f *factory) RegisterLogger(name string, level Level) bool {
	f.lock.Lock()
	defer f.lock.Unlock()

	if _, ok := f.loggers[name]; ok {
		return false
	}

	f.defaults[name] = level
	f.applyForLogger(name)

	return true
}

func (f *factory) LogLevels() map[string]Level {
	f.lock.Lock()
	defer f.lock.Unlock()

	q := make(map[string]Level, len(f.loggers))

	for k, v := range f.loggers {
		q[k] = Level(v.GetLevel())
	}

	return q
}

func (f *factory) getLogger(name string) *zerolog.Logger {
	f.lock.Lock()
	defer f.lock.Unlock()

	l, ok := f.loggers[name]
	if ok {
		return l
	}

	return nil
}

func (f *factory) Get(name string) Logger {
	return &chain{
		logger: &logger{
			factory: f,
			name:    name,
		},
	}
}

type LoggerIO interface {
	io.Writer
}

type loggerIO struct {
	parent *chain

	caller func(l Logger, msg string)
}

func (l loggerIO) Write(p []byte) (n int, err error) {
	n = len(p)
	if n > 0 && p[n-1] == '\n' {
		// Trim CR added by stdlog.
		p = p[0 : n-1]
	}
	l.caller(l.parent, string(p))
	return
}

type Logger interface {
	Wrap(w Wrap) Logger
	WrapObj(w WrapObj) Logger

	Bool(key string, i bool) Logger
	Str(key, value string) Logger
	Strs(key string, values ...string) Logger
	SinceStart(key string, start time.Time) Logger
	Err(err error) Logger
	Int(key string, i int) Logger
	Int32(key string, i int32) Logger
	Int64(key string, i int64) Logger
	Interface(key string, i interface{}) Logger
	Dur(key string, dur time.Duration) Logger
	Time(key string, time time.Time) Logger

	Trace(msg string, args ...interface{})
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
	Fatal(msg string, args ...interface{})

	TraceIO() LoggerIO
	DebugIO() LoggerIO
	InfoIO() LoggerIO
	WarnIO() LoggerIO
	ErrorIO() LoggerIO
	FatalIO() LoggerIO
}

type logger struct {
	factory *factory

	name string
}

type chain struct {
	*logger

	parent *chain

	wrap Wrap
}

func (c *chain) TraceIO() LoggerIO {
	return loggerIO{
		parent: c,
		caller: func(l Logger, msg string) {
			l.Trace(msg)
		},
	}
}

func (c *chain) DebugIO() LoggerIO {
	return loggerIO{
		parent: c,
		caller: func(l Logger, msg string) {
			l.Debug(msg)
		},
	}
}

func (c *chain) InfoIO() LoggerIO {
	return loggerIO{
		parent: c,
		caller: func(l Logger, msg string) {
			l.Info(msg)
		},
	}
}

func (c *chain) WarnIO() LoggerIO {
	return loggerIO{
		parent: c,
		caller: func(l Logger, msg string) {
			l.Warn(msg)
		},
	}
}

func (c *chain) ErrorIO() LoggerIO {
	return loggerIO{
		parent: c,
		caller: func(l Logger, msg string) {
			l.Error(msg)
		},
	}
}

func (c *chain) FatalIO() LoggerIO {
	return loggerIO{
		parent: c,
		caller: func(l Logger, msg string) {
			l.Fatal(msg)
		},
	}
}

func (c *chain) Int64(key string, i int64) Logger {
	return c.Wrap(Int64(key, i))
}

func (c *chain) WrapObj(w WrapObj) Logger {
	return c.Wrap(w.WrapLogger)
}

func (c *chain) Bool(key string, i bool) Logger {
	return c.Wrap(Bool(key, i))
}

func (c *chain) Int32(key string, i int32) Logger {
	return c.Wrap(Int32(key, i))
}

func (c *chain) Time(key string, time time.Time) Logger {
	return c.Wrap(Time(key, time))
}

func (c *chain) Strs(key string, values ...string) Logger {
	return c.Wrap(Strs(key, values...))
}

func (c *chain) Dur(key string, dur time.Duration) Logger {
	return c.Wrap(Dur(key, dur))
}

func (c *chain) Int(key string, i int) Logger {
	return c.Wrap(Int(key, i))
}

func (c *chain) Interface(key string, i interface{}) Logger {
	return c.Wrap(Interface(key, i))
}

func (c *chain) Err(err error) Logger {
	if err == nil {
		return c
	}
	return c.Wrap(Err(err))
}

func (c *chain) SinceStart(key string, start time.Time) Logger {
	return c.Wrap(SinceStart(key, start))
}

func (c *chain) Str(key, value string) Logger {
	return c.Wrap(Str(key, value))
}

func (c *chain) apply(in *zerolog.Event) *zerolog.Event {
	if p := c.parent; c.parent != nil {
		in = p.apply(in)
	} else {
		// We are on root, check factory
		if w := c.factory.wrappers; len(w) > 0 {
			for id := range w {
				if w[id] == nil {
					continue
				}

				in = w[id](in)
			}
		}
	}

	if c.wrap != nil {
		return c.wrap(in)
	}

	return in
}

func (c *chain) Trace(msg string, args ...interface{}) {
	l := c.factory.getLogger(c.name)
	if l == nil {
		return
	}

	c.apply(l.Trace()).Msgf(msg, args...)
}

func (c *chain) Debug(msg string, args ...interface{}) {
	l := c.factory.getLogger(c.name)
	if l == nil {
		return
	}

	c.apply(l.Debug()).Msgf(msg, args...)
}

func (c *chain) Info(msg string, args ...interface{}) {
	l := c.factory.getLogger(c.name)
	if l == nil {
		return
	}

	c.apply(l.Info()).Msgf(msg, args...)
}

func (c *chain) Warn(msg string, args ...interface{}) {
	l := c.factory.getLogger(c.name)
	if l == nil {
		return
	}

	c.apply(l.Warn()).Msgf(msg, args...)
}

func (c *chain) Error(msg string, args ...interface{}) {
	l := c.factory.getLogger(c.name)
	if l == nil {
		return
	}

	c.apply(l.Error()).Msgf(msg, args...)
}

func (c *chain) Fatal(msg string, args ...interface{}) {
	l := c.factory.getLogger(c.name)
	if l == nil {
		return
	}

	c.apply(l.Fatal()).Msgf(msg, args...)
}

func (c *chain) Wrap(w Wrap) Logger {
	return &chain{
		logger: c.logger,
		parent: c,
		wrap:   w,
	}
}
