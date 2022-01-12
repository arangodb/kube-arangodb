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

package utils

import (
	"time"

	"github.com/rs/zerolog/log"
)

// Retry retries action with defined intervals
func Retry(retries int, interval time.Duration, action func() error) error {
	t := time.NewTicker(interval)
	defer t.Stop()

	retry := 0

	for {

		err := action()

		if err == nil {
			return nil
		}

		retry++
		if retry >= retries {
			return err
		}

		log.Error().Err(err).Msgf("Failure, retrying")
		<-t.C
	}
}
