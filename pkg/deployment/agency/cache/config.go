//
// DISCLAIMER
//
// Copyright 2023 ArangoDB GmbH, Cologne, Germany
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

package cache

import (
	"time"

	"github.com/spf13/cobra"
)

func Init(cmd *cobra.Command) error {
	f := cmd.PersistentFlags()

	f.Bool("agency.poll-enabled", false, "The Agency poll functionality enablement (EnterpriseEdition Only)")

	if err := f.MarkHidden("agency.poll-enabled"); err != nil {
		return err
	}
	if err := f.MarkDeprecated("agency.poll-enabled", "Flag moved to feature"); err != nil {
		return err
	}

	f.DurationVar(&global.RefreshDelay, "agency.refresh-delay", 500*time.Millisecond, "The Agency refresh delay (0 = no delay)")
	f.DurationVar(&global.RefreshInterval, "agency.refresh-interval", 0, "The Agency refresh interval (0 = do not refresh)")

	return nil
}

var global Config

func GlobalConfig() Config {
	return global
}

type Config struct {
	RefreshDelay    time.Duration
	RefreshInterval time.Duration
}
