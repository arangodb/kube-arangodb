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

package v1beta1

type ChartDetails struct {
	Name     string                `json:"name,omitempty"`
	Version  string                `json:"version,omitempty"`
	Platform *ChartDetailsPlatform `json:"platform,omitempty"`
}

func (c *ChartDetails) GetPlatform() *ChartDetailsPlatform {
	if c == nil {
		return nil
	}

	return c.Platform
}

func (c *ChartDetails) GetName() string {
	if c == nil {
		return ""
	}

	return c.Name
}

func (c *ChartDetails) GetVersion() string {
	if c == nil {
		return ""
	}

	return c.Version
}
