//
// DISCLAIMER
//
// Copyright 2018 ArangoDB GmbH, Cologne, Germany
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

package v1alpha

import (
	"github.com/pkg/errors"
)

// StorageEngine specifies the type of storage engine used by the cluster
type StorageEngine string

const (
	// StorageEngineMMFiles yields a cluster using the mmfiles storage engine
	StorageEngineMMFiles StorageEngine = "mmfiles"
	// StorageEngineRocksDB yields a cluster using the rocksdb storage engine
	StorageEngineRocksDB StorageEngine = "rocksdb"
)

// Validate the storage engine.
// Return errors when validation fails, nil on success.
func (se StorageEngine) Validate() error {
	switch se {
	case StorageEngineMMFiles, StorageEngineRocksDB:
		return nil
	default:
		return maskAny(errors.Wrapf(ValidationError, "Unknown storage engine: '%s'", string(se)))
	}
}
