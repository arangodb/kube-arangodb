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

package pack

import (
	"fmt"

	"github.com/arangodb/kube-arangodb/pkg/util"
)

type Proto struct {
	Charts ProtoCharts `json:"charts,omitempty"`

	Manifests map[string]string `json:"manifests,omitempty"`
}

type ProtoCharts map[string]ProtoChart

type ProtoChart struct {
	Version string `json:"version"`

	Images ProtoImages `json:"images,omitempty"`
}

type ProtoValues struct {
	Images ProtoImages `json:"images,omitempty"`
}

type ProtoImages map[string]ProtoImage

type ProtoImage struct {
	Registry *string `json:"registry,omitempty"`
	Image    *string `json:"image,omitempty"`
	Tag      *string `json:"tag,omitempty"`
	Kind     *string `json:"kind,omitempty"`
}

func (p ProtoImage) IsTest() bool {
	return util.OptionalType(p.Kind, "") == "Test"
}

func (p ProtoImage) GetShortImage() string {
	return fmt.Sprintf("%s:%s", p.Image, p.Tag)
}

func (p ProtoImage) GetImage() string {
	if p.Registry == nil {
		return fmt.Sprintf("%s:%s", p.Image, p.Tag)
	}

	return fmt.Sprintf("%s/%s:%s", *p.Registry, p.Image, p.Tag)
}
