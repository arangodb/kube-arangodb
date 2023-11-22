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

package kubernetes

import (
	"github.com/rs/zerolog"

	"github.com/arangodb/kube-arangodb/pkg/debug_package/shared"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
)

func ML() shared.Factory {
	return shared.NewFactory("ml", true, ml)
}

func ml(logger zerolog.Logger, files chan<- shared.File) error {
	k, ok := kclient.GetDefaultFactory().Client()
	if !ok {
		return errors.Newf("Client is not initialised")
	}

	if err := mlExtensions(logger, files, k); err != nil {
		logger.Err(err).Msgf("Error while collecting arango ml extension")
		return err
	}

	if err := mlStorages(logger, files, k); err != nil {
		logger.Err(err).Msgf("Error while collecting arango ml storage")
		return err
	}

	if err := mlBatchJobs(logger, files, k); err != nil {
		logger.Err(err).Msgf("Error while collecting arango ml batch jobs")
		return err
	}

	if err := mlCronJobs(logger, files, k); err != nil {
		logger.Err(err).Msgf("Error while collecting arango ml cron jobs")
		return err
	}

	return nil
}
