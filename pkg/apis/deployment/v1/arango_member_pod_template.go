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

package v1

import (
	"crypto/sha256"
	"fmt"

	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/json"
)

func GetArangoMemberPodTemplate(pod *core.PodTemplateSpec, podSpecChecksum string) (*ArangoMemberPodTemplate, error) {
	data, err := json.Marshal(pod.Spec)
	if err != nil {
		return nil, err
	}

	checksum := fmt.Sprintf("%0x", sha256.Sum256(data))

	if podSpecChecksum == "" {
		podSpecChecksum = checksum
	}

	return &ArangoMemberPodTemplate{
		PodSpec:         pod,
		PodSpecChecksum: podSpecChecksum,
		Checksum:        checksum,
	}, nil
}

type ArangoMemberPodTemplate struct {
	PodSpec         *core.PodTemplateSpec `json:"podSpec,omitempty"`
	PodSpecChecksum string                `json:"podSpecChecksum,omitempty"`
	Checksum        string                `json:"checksum,omitempty"`

	// deprecated
	Endpoint *string `json:"endpoint,omitempty"`
}

func (a *ArangoMemberPodTemplate) GetChecksum() string {
	if a == nil {
		return ""
	}
	return a.Checksum
}

func (a *ArangoMemberPodTemplate) Equals(b *ArangoMemberPodTemplate) bool {
	if a == nil && b == nil {
		return true
	}

	if a == nil || b == nil {
		return false
	}

	return a.Checksum == b.Checksum
}

func (a *ArangoMemberPodTemplate) RotationNeeded(b *ArangoMemberPodTemplate) bool {
	if a == nil && b == nil {
		return true
	}

	if a == nil || b == nil {
		return true
	}

	return a.Checksum != b.Checksum
}

func (a *ArangoMemberPodTemplate) EqualPodSpecChecksum(checksum string) bool {
	if a == nil {
		return false
	}
	return checksum == a.PodSpecChecksum
}
