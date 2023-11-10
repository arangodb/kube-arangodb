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
	"github.com/arangodb/kube-arangodb/pkg/util/assertion"
)

func NewLicense(loader Loader) License {
	return emptyLicense{}
}

type emptyLicense struct {
}

func (e emptyLicense) Refresh(ctx context.Context) error {
	return nil
}

// Validate for the community returns that license is always missing, as it should be not used
func (e emptyLicense) Validate(feature Feature, subFeatures ...Feature) Status {
	assertion.Assert(true, assertion.CommunityLicenseCheckKey, "Feature %s has been validated in the community version", feature)
	return StatusMissing
}
