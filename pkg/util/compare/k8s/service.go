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

package k8s

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"

	core "k8s.io/api/core/v1"

	"github.com/arangodb/kube-arangodb/pkg/logging"
	"github.com/arangodb/kube-arangodb/pkg/util/compare"
)

func GetKubernetesServiceSpecDiff(logger logging.Logger, spec, current core.ServiceSpec) (compare.Mode, error) {
	specTmpl, err := compare.NewGenericChecksumTemplate[core.ServiceSpec, *core.ServiceSpec](&spec, checksumServiceSpec)
	if err != nil {
		return compare.SilentRotation, err
	}

	actualTmpl, err := compare.NewGenericChecksumTemplate[core.ServiceSpec, *core.ServiceSpec](&current, checksumServiceSpec)
	if err != nil {
		return compare.SilentRotation, err
	}

	mode, _, err := compare.P0[core.ServiceSpec](logger, compare.NewActionBuilderStub(), checksumServiceSpec, specTmpl, actualTmpl)
	return mode, err
}

func checksumServiceSpec(s *core.ServiceSpec) (string, error) {
	parts := map[string]interface{}{
		"type":     s.Type,
		"ports":    s.Ports,
		"selector": s.Selector,
		// add here more fields when needed
	}

	data, err := json.Marshal(parts)
	if err != nil {
		return "", err
	}

	checksum := fmt.Sprintf("%0x", sha256.Sum256(data))
	return checksum, nil
}
