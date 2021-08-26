//
// DISCLAIMER
//
// Copyright 2021 ArangoDB GmbH, Cologne, Germany
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
// Author Adam Janikowski
//

package reconcile

import (
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
)

func newPlanAppender(pb WithPlanBuilder, current api.Plan) PlanAppender {
	return planAppenderType{
		current: current,
		pb:      pb,
	}
}

type PlanAppender interface {
	Apply(pb planBuilder) PlanAppender
	ApplyWithCondition(c planBuilderCondition, pb planBuilder) PlanAppender
	ApplySubPlan(pb planBuilderSubPlan, plans ...planBuilder) PlanAppender

	ApplyIfEmpty(pb planBuilder) PlanAppender
	ApplyWithConditionIfEmpty(c planBuilderCondition, pb planBuilder) PlanAppender
	ApplySubPlanIfEmpty(pb planBuilderSubPlan, plans ...planBuilder) PlanAppender

	Plan() api.Plan
}

type planAppenderType struct {
	pb      WithPlanBuilder
	current api.Plan
}

func (p planAppenderType) Plan() api.Plan {
	return p.current
}

func (p planAppenderType) ApplyIfEmpty(pb planBuilder) PlanAppender {
	if p.current.IsEmpty() {
		return p.Apply(pb)
	}
	return p
}

func (p planAppenderType) ApplyWithConditionIfEmpty(c planBuilderCondition, pb planBuilder) PlanAppender {
	if p.current.IsEmpty() {
		return p.ApplyWithCondition(c, pb)
	}
	return p
}

func (p planAppenderType) ApplySubPlanIfEmpty(pb planBuilderSubPlan, plans ...planBuilder) PlanAppender {
	if p.current.IsEmpty() {
		return p.ApplySubPlan(pb, plans...)
	}
	return p
}

func (p planAppenderType) new(plan api.Plan) planAppenderType {
	return planAppenderType{
		pb:      p.pb,
		current: append(p.current, plan...),
	}
}

func (p planAppenderType) Apply(pb planBuilder) PlanAppender {
	return p.new(p.pb.Apply(pb))
}

func (p planAppenderType) ApplyWithCondition(c planBuilderCondition, pb planBuilder) PlanAppender {
	return p.new(p.pb.ApplyWithCondition(c, pb))
}

func (p planAppenderType) ApplySubPlan(pb planBuilderSubPlan, plans ...planBuilder) PlanAppender {
	return p.new(p.pb.ApplySubPlan(pb, plans...))
}
