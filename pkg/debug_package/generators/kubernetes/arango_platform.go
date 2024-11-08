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

package kubernetes

import (
	"github.com/rs/zerolog"

	"github.com/arangodb/kube-arangodb/pkg/debug_package/shared"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
)

func Platform() shared.Factory {
	return shared.NewFactory("platform", true, platform)
}

func platform(logger zerolog.Logger, files chan<- shared.File) error {
	k, ok := kclient.GetDefaultFactory().Client()
	if !ok {
		return errors.Errorf("Client is not initialised")
	}

	if err := platformArangoPlatformStorages(logger, files, k); err != nil {
		logger.Err(err).Msgf("Error while collecting platform arango storages")
		return err
	}

	return nil
}
