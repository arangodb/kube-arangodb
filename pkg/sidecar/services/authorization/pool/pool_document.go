//
// DISCLAIMER
//
// Copyright 2026 ArangoDB GmbH, Cologne, Germany
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

package pool

import (
	"google.golang.org/protobuf/proto"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type DocumentAction string

var (
	DocumentCreateAction DocumentAction = "create"
	DocumentUpdateAction DocumentAction = "update"
	DocumentDeleteAction DocumentAction = "delete"
)

type Document[T proto.Message] struct {
	Key string `json:"_key"`

	Name string `json:"name"`

	Rev *string `json:"_rev,omitempty"`

	Sequence uint32 `json:"sequence"`

	Created meta.Time `json:"created,omitempty"`

	Deleted meta.Time `json:"deleted,omitempty"`

	Action DocumentAction `json:"action,omitempty"`

	Spec T `json:"spec,omitempty"`
}
