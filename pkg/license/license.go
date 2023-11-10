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

package license

import (
	"context"
)

type Status int

const (
	// StatusMissing define state when the license could not be loaded from the start or was not provided
	// NotLicensed
	StatusMissing Status = iota

	// StatusInvalid define state when the license and any of the fields are not valid
	// NotLicensed
	StatusInvalid

	// StatusInvalidSignature define state when license signature could not be validated
	// NotLicensed
	StatusInvalidSignature

	// StatusNotYetValid define state when the license contains nbf and current time (UTC) is before a specified time
	// NotLicensed
	StatusNotYetValid

	// StatusNotAnymoreValid define state when the license contains exp and current time (UTC) is after a specified time
	// NotLicensed
	StatusNotAnymoreValid

	// StatusFeatureNotEnabled define state when features requirements does not match one requested by the feature in Operator
	// NotLicensed
	StatusFeatureNotEnabled

	// StatusFeatureExpired define state when token is valid, but feature itself is expired
	// NotLicensed
	StatusFeatureExpired

	// StatusValid define state when  Operator should continue execution
	// Licensed
	StatusValid
)

func (s Status) Valid() bool {
	return s == StatusValid
}

func (s Status) Validate(feature Feature, subFeatures ...Feature) Status {
	return s
}

type Feature string

type License interface {
	// Validate validates the license scope. In case of:
	// - if feature is '*' - checks if:
	// -- license is valid and not expired
	// - if feature is not '*' and subFeatures list is empty - checks if:
	// -- license is valid and not expired
	// -- feature is enabled and not expired
	// - if feature is not '*' and subFeatures list is not empty - checks if:
	// -- license is valid and not expired
	// -- feature is enabled and not expired
	// -- for each subFeature defined in subFeatures:
	// --- checks if subFeature or '*' is in the list of License Feature enabled SubFeatures
	Validate(feature Feature, subFeatures ...Feature) Status

	Refresh(ctx context.Context) error
}