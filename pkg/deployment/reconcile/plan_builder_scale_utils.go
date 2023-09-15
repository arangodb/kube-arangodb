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

package reconcile

import (
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

type ScaleFilterFunc func(context PlanBuilderContext, status api.DeploymentStatus, group api.ServerGroup, in api.MemberStatusList) (api.MemberStatusList, bool, error)
type ScaleSelectFunc func(context PlanBuilderContext, status api.DeploymentStatus, group api.ServerGroup, in api.MemberStatusList) (api.MemberStatus, bool, error)

func NewScaleFilter(context PlanBuilderContext, status api.DeploymentStatus, group api.ServerGroup, in api.MemberStatusList) ScaleFilter {
	return scaleFilter{
		context: context,
		in:      in,
		group:   group,
		status:  status,
	}
}

type ScaleFilter interface {
	Filter(in ScaleFilterFunc) ScaleFilter
	Select(in ScaleSelectFunc) ScaleFilter
	Get() (api.MemberStatus, error)
}

type scaleFilter struct {
	context PlanBuilderContext
	in      api.MemberStatusList
	group   api.ServerGroup
	status  api.DeploymentStatus
}

func (s scaleFilter) Filter(in ScaleFilterFunc) ScaleFilter {
	if len(s.in) == 1 {
		return s
	}

	filtered, changed, err := in(s.context, s.status, s.group, s.in)
	if err != nil {
		return scaleFilterError{
			err: err,
		}
	}

	if changed {
		return scaleFilter{
			context: s.context,
			group:   s.group,
			in:      filtered,
		}
	}

	return s
}

func (s scaleFilter) Select(in ScaleSelectFunc) ScaleFilter {
	if len(s.in) == 1 {
		return s
	}

	selected, changed, err := in(s.context, s.status, s.group, s.in)
	if err != nil {
		return scaleFilterError{
			err: err,
		}
	}

	if changed {
		return scaleFilter{
			context: s.context,
			group:   s.group,
			in:      []api.MemberStatus{selected},
		}
	}

	return s
}

func (s scaleFilter) Get() (api.MemberStatus, error) {
	el, ok := util.RandomElement(s.in)
	if !ok {
		return api.MemberStatus{}, errors.Newf("Unable to select member")
	}

	return el, nil
}

type scaleFilterError struct {
	err error
}

func (s scaleFilterError) Filter(in ScaleFilterFunc) ScaleFilter {
	return s
}

func (s scaleFilterError) Select(in ScaleSelectFunc) ScaleFilter {
	return s
}

func (s scaleFilterError) Get() (api.MemberStatus, error) {
	return api.MemberStatus{}, s.err
}
