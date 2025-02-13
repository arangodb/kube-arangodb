//
// DISCLAIMER
//
// Copyright 2023-2024 ArangoDB GmbH, Cologne, Germany
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

package generic

import (
	"context"

	autoscaling "k8s.io/api/autoscaling/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type ListContinue interface {
	GetContinue() string
}

type ExtractorList[L ListContinue, S meta.Object] func(in L) []S

type ListInterface[S ListContinue] interface {
	List(ctx context.Context, opts meta.ListOptions) (S, error)
}

type GetInterface[S meta.Object] interface {
	Get(ctx context.Context, name string, opts meta.GetOptions) (S, error)
}

type CreateInterface[S meta.Object] interface {
	Create(ctx context.Context, obj S, opts meta.CreateOptions) (S, error)
}

type UpdateInterface[S meta.Object] interface {
	Update(ctx context.Context, obj S, opts meta.UpdateOptions) (S, error)
}

type UpdateStatusInterface[S meta.Object] interface {
	UpdateStatus(ctx context.Context, obj S, opts meta.UpdateOptions) (S, error)
}

type PatchInterface[S meta.Object] interface {
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts meta.PatchOptions, subresources ...string) (result S, err error)
}

type DeleteInterface[S meta.Object] interface {
	Delete(ctx context.Context, name string, opts meta.DeleteOptions) error
}

type ScaleGet interface {
	GetScale(ctx context.Context, name string, options meta.GetOptions) (*autoscaling.Scale, error)
}

type ScaleUpdate interface {
	UpdateScale(ctx context.Context, name string, scale *autoscaling.Scale, opts meta.UpdateOptions) (*autoscaling.Scale, error)
}

type Scale interface {
	ScaleGet
	ScaleUpdate
}

type ReadClient[S meta.Object] interface {
	GetInterface[S]
}

type ModClient[S meta.Object] interface {
	CreateInterface[S]
	UpdateInterface[S]
	PatchInterface[S]
	DeleteInterface[S]
}

type ModStatusClient[S meta.Object] interface {
	ModClient[S]
	UpdateStatusInterface[S]
}
