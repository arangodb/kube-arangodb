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

package v1alpha1

import (
	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"
	sharedApi "github.com/arangodb/kube-arangodb/pkg/apis/shared/v1"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

type ArangoPlatformChartSpec struct {
	// Definition keeps the Chart base64 encoded definition
	Definition sharedApi.Data `json:"definition,omitempty"`

	// Overrides keeps the Chart overrides
	// +doc/type: Object
	Overrides sharedApi.Any `json:"overrides,omitempty"`
}

func (c *ArangoPlatformChartSpec) Validate() error {
	if c == nil {
		return errors.Errorf("Nil spec not allowed")
	}

	if len(c.Definition) == 0 {
		return shared.PrefixResourceError("definition", errors.Errorf("Chart definition cannot be empty"))
	}

	return nil
}

func (c *ArangoPlatformChartSpec) Checksum() string {
	if c == nil {
		return ""
	}

	checksums := make([]string, 0, 2)
	checksums = append(checksums, c.Definition.SHA256())
	if v := c.Overrides; !v.IsZero() {
		checksums = append(checksums, v.SHA256())
	}

	return util.SHA256FromStringArray(checksums...)
}
