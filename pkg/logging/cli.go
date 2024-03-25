//
// DISCLAIMER
//
// Copyright 2024 ArangoDB GmbH, Cologne, Germany
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
	"strings"
	"sync"

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
	}
)

func Init(cmd *cobra.Command) error {
	f := cmd.PersistentFlags()

	f.StringVar(&cli.format, "log.format", "pretty", "Set log format. Allowed values: 'pretty', 'JSON'. If empty, default format is used")
	f.StringArrayVar(&cli.levels, "log.level", []string{defaultLogLevel}, fmt.Sprintf("Set log levels in format <level> or <logger>=<level>. Possible loggers: %s", strings.Join(Global().Names(), ", ")))
	f.BoolVar(&cli.sampling, "log.sampling", true, "If true, operator will try to minimize duplication of logging events")

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

	// Set root logger to stdout (JSON formatted) if not prettified
	if strings.ToUpper(cli.format) == "JSON" {
		Global().SetRoot(zerolog.New(os.Stdout).With().Timestamp().Logger())
	} else if strings.ToLower(cli.format) != "pretty" && cli.format != "" {
		return errors.Errorf("Unknown log format: %s", cli.format)
	}
	Global().Configure(Config{
		Levels:   levels,
		Sampling: cli.sampling,
	})

	return nil
}
