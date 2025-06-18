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

package v1

import (
	"time"

	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/timestamppb"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	pbMetaV1 "github.com/arangodb/kube-arangodb/integrations/meta/v1/definition"
	ugrpc "github.com/arangodb/kube-arangodb/pkg/util/grpc"
)

type Object struct {
	Key string  `json:"_key,omitempty"`
	Rev *string `json:"_rev,omitempty"`

	Meta *ObjectMeta `json:"meta,omitempty"`

	Object ugrpc.GRPC[*anypb.Any] `json:"object"`
}

func (o *Object) SetKey(s string) {
	if o != nil {
		o.Key = s
	}
}

func (o *Object) GetKey() string {
	if o == nil {
		return ""
	}

	return o.Key
}

func (o *Object) GetRev() string {
	if o == nil || o.Rev == nil {
		return ""
	}

	return *o.Rev
}

func (o *Object) Expires() time.Time {
	if o == nil || o.Meta == nil || o.Meta.Expires == nil {
		return time.Time{}
	}

	return o.Meta.Expires.Time
}

func (o *Object) AsResponse() *pbMetaV1.ObjectResponse {
	if o == nil {
		return nil
	}

	var r pbMetaV1.ObjectResponse

	r.Meta = o.Meta.AsResponse()

	r.Key = o.Key
	r.Revision = o.Rev

	r.Object = o.Object.Object

	return &r
}

type ObjectMeta struct {
	Updated meta.Time `json:"updatedAt,omitempty"`

	Expires *meta.Time `json:"expiresAt,omitempty"`
}

func (o *ObjectMeta) AsResponse() *pbMetaV1.ObjectResponseMeta {
	if o == nil {
		return nil
	}

	var r pbMetaV1.ObjectResponseMeta

	r.Updated = timestamppb.New(o.Updated.Time)

	if e := o.Expires; e != nil {
		r.Expires = timestamppb.New(o.Expires.Time)
	}

	return &r
}
