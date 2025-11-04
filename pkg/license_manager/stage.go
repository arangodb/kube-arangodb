//
// DISCLAIMER
//
// Copyright 2025 ArangoDB GmbH, Cologne, Germany
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

package license_manager

import (
	"fmt"
	goStrings "strings"

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

type Stage int

func ParseStages(s ...string) []Stage {
	return util.FormatList(s, func(a string) Stage {
		return ParseStage(a)
	})
}

func ParseStage(s string) Stage {
	switch goStrings.ToLower(s) {
	case "dev":
		return StageDev
	case "qa":
		return StageQA
	case "prd":
		return StagePrd
	default:
		return StageUnknown
	}
}

const (
	StageUnknown Stage = iota
	StageDev
	StageQA
	StagePrd
)

func (s Stage) RegistryDomain(domain string) (string, error) {
	switch s {
	case StageDev:
		return fmt.Sprintf("dev.registry.%s", domain), nil
	case StageQA:
		return fmt.Sprintf("qa.registry.%s", domain), nil
	case StagePrd:
		return fmt.Sprintf("registry.%s", domain), nil
	}

	return "", errors.Errorf("invalid stage")
}
