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

package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

type cmdUUIDInputStruct struct {
	uuidPath, uuid, engine, enginePath string
	require                            bool
}

var (
	cmdUUID = &cobra.Command{
		Use:    "uuid",
		RunE:   cmdUUIDSave,
		Hidden: true,
	}

	cmdUUIDInput cmdUUIDInputStruct
)

func init() {
	cmdMain.AddCommand(cmdUUID)
	f := cmdUUID.Flags()
	f.StringVar(&cmdUUIDInput.uuidPath, "uuid-path", "", "Path to the UUID file")
	f.StringVar(&cmdUUIDInput.engine, "engine", "", "Path to the ENGINE file")
	f.StringVar(&cmdUUIDInput.uuid, "uuid", "", "UUID of server")
	f.StringVar(&cmdUUIDInput.enginePath, "engine-path", "", "ENGINE of server")
	f.BoolVar(&cmdUUIDInput.require, "require", false, "Ensure that UUID and ENGINE file is in place")
}

func cmdUUIDSave(cmd *cobra.Command, args []string) error {
	if cmdUUIDInput.uuidPath == "" {
		return errors.Errorf("Path cannot be empty")
	}

	if cmdUUIDInput.require {
		if cmdUUIDInput.enginePath == "" {
			return errors.Errorf("Path to the ENGINE cannot be empty")
		}

		if cmdUUIDInput.engine == "" {
			return errors.Errorf("ENGINE cannot be empty")
		}
	}

	if cmdUUIDInput.uuid == "" {
		return errors.Errorf("UUID cannot be empty")
	}

	log.Info().Msg("Saving UUID in file")

	if exists, err := fileExists(cmdUUIDInput.uuidPath); err != nil {
		log.Error().Err(err).Msg("Unable to get file info")
		return err
	} else if !exists {
		if cmdUUIDInput.require {
			log.Warn().Msg("Init phase is not defined, but file does not exists")

			if exists, err := fileExists(cmdUUIDInput.uuidPath); err != nil {
				log.Error().Err(err).Msg("Unable to get ENGINE info")
				return err
			} else if !exists {
				log.Info().Msg("ENGINE file does not exist - able to proceed")
			} else {
				log.Error().Msg("ENGINE file found but UUID is missing - will not proceed")
				return errors.Errorf("ENGINE file found but UUID is missing - will not proceed")
			}
		}

		fileContent := fmt.Sprintf("%s\n", cmdUUIDInput.uuid)
		if err := os.WriteFile(cmdUUIDInput.uuidPath, []byte(fileContent), 0644); err != nil {
			log.Error().Err(err).Msg("Unable to save UUID")
			return err
		}

		log.Info().Msg("UUID saved")
		return nil
	}

	if equal, content, err := checkFileContent(cmdUUIDInput.uuidPath, cmdUUIDInput.uuid); err != nil {
		log.Error().Err(err).Msg("Unable to get UUID info")
		return err
	} else if !equal {
		log.Error().Str("expected", cmdUUIDInput.uuid).Str("actual", content).Msg("UUID does not match expected")
		return errors.Errorf("UUID mismatch")
	} else {
		log.Info().Msg("UUID is valid")
	}

	if cmdUUIDInput.require {
		if exists, err := fileExists(cmdUUIDInput.enginePath); err != nil {
			log.Error().Err(err).Msg("Unable to get ENGINE info")
			return err
		} else if exists {
			if equal, content, err := checkFileContent(cmdUUIDInput.enginePath, cmdUUIDInput.engine); err != nil {
				log.Error().Err(err).Msg("Unable to get ENGINE info")
				return err
			} else if !equal {
				log.Error().Str("expected", cmdUUIDInput.engine).Str("actual", content).Msg("ENGINE does not match expected")
				return errors.Errorf("ENGINE mismatch")
			} else {
				log.Info().Msg("ENGINE is valid")
			}
		} else {
			log.Info().Msg("ENGINE file is missing, but UUID is in place - we can proceed")
		}
	}

	log.Info().Msg("Init Phase is valid")
	return nil
}

func fileExists(path string) (bool, error) {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

func checkFileContent(path, expected string) (bool, string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return false, "", err
	}

	contentString := strings.TrimSuffix(string(content), "\n")

	return contentString == expected, contentString, nil
}
