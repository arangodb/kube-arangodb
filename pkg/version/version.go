//
// DISCLAIMER
//
// Copyright 2016-2023 ArangoDB GmbH, Cologne, Germany
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

package version

import (
	"runtime"

	"github.com/arangodb/go-driver"

	"github.com/arangodb/kube-arangodb/pkg/util/strings"
)

type License string

const (
	CommunityEdition  License = "community"
	EnterpriseEdition License = "enterprise"
)

func (s License) Title() string {
	return strings.Title(string(s))
}

var (
	version   = "dev"
	build     = "dev"
	buildDate = ""
	goVersion = runtime.Version()
)

type InfoV1 struct {
	Version   driver.Version `json:"version"`
	Build     string         `json:"build"`
	Edition   License        `json:"edition"`
	GoVersion string         `json:"go_version"`
	BuildDate string         `json:"build_date,omitempty"`
}

func (i InfoV1) IsEnterprise() bool {
	return i.Edition == EnterpriseEdition
}

func GetVersionV1() InfoV1 {
	return InfoV1{
		Version:   driver.Version(version),
		Build:     build,
		Edition:   edition,
		GoVersion: goVersion,
		BuildDate: buildDate,
	}
}
