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

package cmd

import (
	"encoding/json"
	"os"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

const cmdVersionCheckInitContainersInvalidVersionExitCode = 11

type cmdVersionCheckInitContainersInputStruct struct {
	versionPath string

	major, minor int
}

var (
	cmdVersionCheckInitContainers = &cobra.Command{
		Use:  "version-check",
		RunE: cmdVersionCheckInitContainersInput.Run,
	}

	cmdVersionCheckInitContainersInput cmdVersionCheckInitContainersInputStruct
)

type cmdVersionCheckInitContainersData struct {
	Version int `json:"version,omitempty"`
}

func init() {
	cmdInitContainers.AddCommand(cmdVersionCheckInitContainers)
	f := cmdVersionCheckInitContainers.Flags()
	f.StringVar(&cmdVersionCheckInitContainersInput.versionPath, "path", "", "Path to the VERSION file")
	f.IntVar(&cmdVersionCheckInitContainersInput.major, "major", 0, "Major version of the ArangoDB. 0 if check is disabled")
	f.IntVar(&cmdVersionCheckInitContainersInput.minor, "minor", 0, "Minor version of the ArangoDB. 0 if check is disabled")
}

func (c cmdVersionCheckInitContainersInputStruct) Run(cmd *cobra.Command, args []string) error {
	if c.versionPath == "" {
		return errors.Errorf("Path cannot be empty")
	}

	if data, err := os.ReadFile(c.versionPath); err != nil {
		log.Err(err).Msg("File is not readable, continue")
		return nil
	} else {
		major, minor, _, ok := extractVersionFromData(data)
		if !ok {
			return nil
		}

		if c.major != 0 {
			if c.major != major {
				return Exit(cmdVersionCheckInitContainersInvalidVersionExitCode)
			}
			if c.minor != 0 {
				if c.minor != minor {
					return Exit(cmdVersionCheckInitContainersInvalidVersionExitCode)
				}
			}
		}

		return nil
	}
}

func extractVersionFromData(data []byte) (int, int, int, bool) {
	var c cmdVersionCheckInitContainersData

	if err := json.Unmarshal(data, &c); err != nil {
		log.Err(err).Msg("Invalid json, continue")
		return 0, 0, 0, false
	}

	return c.Version / 10000, c.Version % 10000 / 100, c.Version % 100, true
}
