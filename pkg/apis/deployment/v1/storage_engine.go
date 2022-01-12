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

package v1

import (
	"strings"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

// StorageEngine specifies the type of storage engine used by the cluster
type StorageEngine string

const (
	// StorageEngineMMFiles yields a cluster using the mmfiles storage engine
	StorageEngineMMFiles StorageEngine = "MMFiles"
	// StorageEngineRocksDB yields a cluster using the rocksdb storage engine
	StorageEngineRocksDB StorageEngine = "RocksDB"
)

// Validate the storage engine.
// Return errors when validation fails, nil on success.
func (se StorageEngine) Validate() error {
	switch se {
	case StorageEngineMMFiles, StorageEngineRocksDB:
		return nil
	default:
		return errors.WithStack(errors.Wrapf(ValidationError, "Unknown storage engine: '%s'", string(se)))
	}
}

// AsArangoArgument returns the value for the given storage engine as it is to be used
// for arangod's --server.storage-engine option.
func (se StorageEngine) AsArangoArgument() string {
	return strings.ToLower(string(se))
}

// NewStorageEngine returns a reference to a string with given value.
func NewStorageEngine(input StorageEngine) *StorageEngine {
	return &input
}

// NewStorageEngineOrNil returns nil if input is nil, otherwise returns a clone of the given value.
func NewStorageEngineOrNil(input *StorageEngine) *StorageEngine {
	if input == nil {
		return nil
	}
	return NewStorageEngine(*input)
}

// StorageEngineOrDefault returns the default value (or empty string) if input is nil, otherwise returns the referenced value.
func StorageEngineOrDefault(input *StorageEngine, defaultValue ...StorageEngine) StorageEngine {
	if input == nil {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return ""
	}
	return *input
}
