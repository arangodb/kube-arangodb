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

package util

import "github.com/Masterminds/semver/v3"

type VersionConstrain interface {
	Validate(version string) (bool, error)
}

type versionConstrain struct {
	constrain *semver.Constraints
}

func (v versionConstrain) Validate(version string) (bool, error) {
	ver, err := semver.NewVersion(version)
	if err != nil {
		return false, err
	}

	if ver.Prerelease() != "" {
		nver, nerr := ver.SetPrerelease("")
		if nerr != nil {
			return false, nerr
		}
		ver = &nver
	}

	return v.constrain.Check(ver), nil
}

func NewVersionConstrain(constrain string) (VersionConstrain, error) {
	c, err := semver.NewConstraint(constrain)
	if err != nil {
		return nil, err
	}

	return versionConstrain{constrain: c}, nil
}
