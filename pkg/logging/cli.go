//
// DISCLAIMER
//
// Copyright 2024-2025 ArangoDB GmbH, Cologne, Germany
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
	"os"
	goStrings "strings"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

const (
	defaultLogLevel = "info"
)

var (
	enableLock sync.Mutex
	enabled    bool

	cli struct {
		format   string
		levels   []string
		sampling bool
		stdout   bool
	}
)

func Init(cmd *cobra.Command) error {
	f := cmd.PersistentFlags()

	f.StringVar(&cli.format, "log.format", "pretty", "Set log format. Allowed values: 'pretty', 'JSON'. If empty, default format is used")
	f.StringArrayVar(&cli.levels, "log.level", []string{defaultLogLevel}, fmt.Sprintf("Set log levels in format <level> or <logger>=<level>. Possible loggers: %s", goStrings.Join(Global().Names(), ", ")))
	f.BoolVar(&cli.sampling, "log.sampling", true, "If true, operator will try to minimize duplication of logging events")
	f.BoolVar(&cli.stdout, "log.stdout", true, "If true, operator will log to the stdout")

	return nil
}

func Enable() error {
	enableLock.Lock()
	defer enableLock.Unlock()

	if enabled {
		return errors.Errorf("Logger already enabled")
	}

	levels, err := ParseLogLevelsFromArgs(cli.levels)
	if err != nil {
		return errors.WithMessagef(err, "Unable to parse levels")
	}

	out := os.Stderr

	if cli.stdout {
		out = os.Stdout
	}

	switch goStrings.ToUpper(cli.format) {
	case "JSON":
		Global().SetRoot(zerolog.New(out).With().Timestamp().Logger())
	case "PRETTY", "":
		Global().SetRoot(zerolog.New(zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339Nano,
			NoColor:    true,
		}).With().Timestamp().Logger())
	default:
		return errors.Errorf("Unknown log format: %s", cli.format)
	}

	Global().Configure(Config{
		Levels:   levels,
		Sampling: cli.sampling,
	})

	return nil
}

func Runner(cmd *cobra.Command, args []string) error {
	return Enable()
}
