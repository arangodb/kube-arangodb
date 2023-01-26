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
	"context"

	policy "k8s.io/api/policy/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// ModInterface has methods to work with PodDisruptionBudget resources only for creation
type ModInterface interface {
	Create(ctx context.Context, poddisruptionbudget *policy.PodDisruptionBudget, opts meta.CreateOptions) (*policy.PodDisruptionBudget, error)
	Update(ctx context.Context, poddisruptionbudget *policy.PodDisruptionBudget, opts meta.UpdateOptions) (*policy.PodDisruptionBudget, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts meta.PatchOptions, subresources ...string) (result *policy.PodDisruptionBudget, err error)
	Delete(ctx context.Context, name string, opts meta.DeleteOptions) error
}

// Interface has methods to work with PodDisruptionBudget resources.
type Interface interface {
	ModInterface
	ReadInterface
}

// ReadInterface has methods to work with PodDisruptionBudget resources with ReadOnly mode.
type ReadInterface interface {
	Get(ctx context.Context, name string, opts meta.GetOptions) (*policy.PodDisruptionBudget, error)
}
