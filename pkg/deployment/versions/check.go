//
// DISCLAIMER
//
// Copyright 2016-2026 ArangoDB GmbH, Cologne, Germany
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

package versions

import (
	adbDriverV2 "github.com/arangodb/go-driver/v2/arangodb"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
)

func NewCheck(version api.ImageInfo) Check {
	return check{
		version: version,
	}
}

type Check interface {
	Enterprise() Check
	Community() Check

	Above(version adbDriverV2.Version) Check
	AboveOrEqual(version adbDriverV2.Version) Check

	Below(version adbDriverV2.Version) Check
	BelowOrEqual(version adbDriverV2.Version) Check

	Evaluate() bool
}

type check struct {
	version api.ImageInfo
}

func (c check) Below(version adbDriverV2.Version) Check {
	if c.version.ArangoDBVersion.CompareTo(version) == 1 {
		return c
	}

	return falseCheck{}
}

func (c check) BelowOrEqual(version adbDriverV2.Version) Check {
	if c.version.ArangoDBVersion.CompareTo(version) <= 0 {
		return c
	}

	return falseCheck{}
}

func (c check) Above(version adbDriverV2.Version) Check {
	if c.version.ArangoDBVersion.CompareTo(version) == -1 {
		return c
	}

	return falseCheck{}
}

func (c check) AboveOrEqual(version adbDriverV2.Version) Check {
	if c.version.ArangoDBVersion.CompareTo(version) >= 0 {
		return c
	}

	return falseCheck{}
}

func (c check) Enterprise() Check {
	if c.version.Enterprise {
		return c
	}

	return falseCheck{}
}

func (c check) Community() Check {
	if !c.version.Enterprise {
		return c
	}

	return falseCheck{}
}

func (c check) Evaluate() bool {
	return true
}

type falseCheck struct {
}

func (f falseCheck) Below(version adbDriverV2.Version) Check {
	return f
}

func (f falseCheck) BelowOrEqual(version adbDriverV2.Version) Check {
	return f
}

func (f falseCheck) Above(version adbDriverV2.Version) Check {
	return f
}

func (f falseCheck) AboveOrEqual(version adbDriverV2.Version) Check {
	return f
}

func (f falseCheck) Enterprise() Check {
	return f
}

func (f falseCheck) Community() Check {
	return f
}

func (f falseCheck) Evaluate() bool {
	return false
}
