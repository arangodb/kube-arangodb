//
// DISCLAIMER
//
// Copyright 2020 ArangoDB GmbH, Cologne, Germany
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

package logging

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"

	"github.com/rs/zerolog"
)

// Service exposes the interfaces for a logger service
// that supports different loggers with different levels.
type Service interface {
	// MustGetLogger creates a logger with given name.
	MustGetLogger(name string) zerolog.Logger
	// MustSetLevel sets the log level for the component with given name to given level.
	MustSetLevel(name, level string)
	// ConfigureRootLogger calls the given callback to modify the root logger.
	ConfigureRootLogger(cb func(rootLog zerolog.Logger) zerolog.Logger)
}

// loggingService implements Service
type loggingService struct {
	mutex        sync.Mutex
	rootLog      zerolog.Logger
	defaultLevel zerolog.Level
	levels       map[string]zerolog.Level
}

// NewRootLogger creates a new zerolog logger with default settings.
func NewRootLogger() zerolog.Logger {
	return zerolog.New(zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339Nano,
		NoColor:    true,
	}).With().Timestamp().Logger()
}

// NewService creates a new Service.
func NewService(defaultLevel string, overrides []string) (Service, error) {
	l, err := stringToLevel(defaultLevel)
	if err != nil {
		return nil, maskAny(err)
	}
	rootLog := NewRootLogger()
	s := &loggingService{
		rootLog:      rootLog,
		defaultLevel: l,
		levels:       make(map[string]zerolog.Level),
	}

	for _, override := range overrides {
		levelParts := strings.Split(override, "=")
		switch size := len(levelParts); size {
		case 1:
			level, err := stringToLevel(levelParts[0])
			if err != nil {
				return nil, maskAny(err)
			}
			s.defaultLevel = level
		case 2:
			level, err := stringToLevel(levelParts[1])
			if err != nil {
				return nil, maskAny(err)
			}
			s.levels[levelParts[0]] = level
		default:
			return nil, errors.Errorf("invalid log definition %s: Length %d is not equal 1 or 2", override, size)
		}
	}
	return s, nil
}

// ConfigureRootLogger calls the given callback to modify the root logger.
func (s *loggingService) ConfigureRootLogger(cb func(rootLog zerolog.Logger) zerolog.Logger) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.rootLog = cb(s.rootLog)
}

// MustGetLogger creates a logger with given name
func (s *loggingService) MustGetLogger(name string) zerolog.Logger {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	level, found := s.levels[name]
	if !found {
		level = s.defaultLevel
	}
	return s.rootLog.With().Str("component", name).Logger().Level(level)
}

// MustSetLevel sets the log level for the component with given name to given level.
func (s *loggingService) MustSetLevel(name, level string) {
	l, err := stringToLevel(level)
	if err != nil {
		panic(err)
	}
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.levels[name] = l
}

// stringToLevel converts a level string to a zerolog level
func stringToLevel(l string) (zerolog.Level, error) {
	switch strings.ToLower(l) {
	case "debug":
		return zerolog.DebugLevel, nil
	case "info":
		return zerolog.InfoLevel, nil
	case "warn", "warning":
		return zerolog.WarnLevel, nil
	case "error":
		return zerolog.ErrorLevel, nil
	case "fatal":
		return zerolog.FatalLevel, nil
	case "panic":
		return zerolog.PanicLevel, nil
	}
	return zerolog.InfoLevel, fmt.Errorf("Unknown log level '%s'", l)
}
