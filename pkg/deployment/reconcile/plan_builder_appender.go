//
// DISCLAIMER
//
// Copyright 2016-2022 ArangoDB GmbH, Cologne, Germany
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
	"time"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/logging"
)

func newPlanAppender(pb WithPlanBuilder, backoff api.BackOff, current api.Plan) PlanAppender {
	return &planAppenderType{
		current: current,
		pb:      pb,
		backoff: backoff.DeepCopy(),
	}
}

func recoverPlanAppender(log logging.Logger, p PlanAppender) PlanAppender {
	return planAppenderRecovery{
		appender: p,
		log:      log,
	}
}

type PlanAppender interface {
	Apply(pb planBuilder) PlanAppender
	ApplyWithCondition(c planBuilderCondition, pb planBuilder) PlanAppender
	ApplySubPlan(pb planBuilderSubPlan, plans ...planBuilder) PlanAppender

	ApplyIfEmpty(pb planBuilder) PlanAppender
	ApplyWithConditionIfEmpty(c planBuilderCondition, pb planBuilder) PlanAppender
	ApplySubPlanIfEmpty(pb planBuilderSubPlan, plans ...planBuilder) PlanAppender

	ApplyWithBackOff(key api.BackOffKey, delay time.Duration, pb planBuilder) PlanAppender
	ApplyIfEmptyWithBackOff(key api.BackOffKey, delay time.Duration, pb planBuilder) PlanAppender

	BackOff() api.BackOff

	Plan() api.Plan
}

type planAppenderRecovery struct {
	log      logging.Logger
	appender PlanAppender
}

func (p planAppenderRecovery) BackOff() api.BackOff {
	return p.appender.BackOff()
}

func (p planAppenderRecovery) ApplyWithBackOff(key api.BackOffKey, delay time.Duration, pb planBuilder) PlanAppender {
	return p.create(func(in PlanAppender) PlanAppender {
		return in.ApplyWithBackOff(key, delay, pb)
	})
}

func (p planAppenderRecovery) ApplyIfEmptyWithBackOff(key api.BackOffKey, delay time.Duration, pb planBuilder) PlanAppender {
	return p.create(func(in PlanAppender) PlanAppender {
		return in.ApplyIfEmptyWithBackOff(key, delay, pb)
	})
}

func (p planAppenderRecovery) create(ret func(in PlanAppender) PlanAppender) (r PlanAppender) {
	defer func() {
		if e := recover(); e != nil {
			r = p
			p.log.Interface("panic", e).Error("Recovering from panic")
		}
	}()

	return planAppenderRecovery{
		appender: ret(p.appender),
		log:      p.log,
	}
}

func (p planAppenderRecovery) Apply(pb planBuilder) PlanAppender {
	return p.create(func(in PlanAppender) PlanAppender {
		return in.Apply(pb)
	})
}

func (p planAppenderRecovery) ApplyWithCondition(c planBuilderCondition, pb planBuilder) PlanAppender {
	return p.create(func(in PlanAppender) PlanAppender {
		return in.ApplyWithCondition(c, pb)
	})
}

func (p planAppenderRecovery) ApplySubPlan(pb planBuilderSubPlan, plans ...planBuilder) (r PlanAppender) {
	return p.create(func(in PlanAppender) PlanAppender {
		return in.ApplySubPlan(pb, plans...)
	})
}

func (p planAppenderRecovery) ApplyIfEmpty(pb planBuilder) (r PlanAppender) {
	return p.create(func(in PlanAppender) PlanAppender {
		return in.ApplyIfEmpty(pb)
	})
}

func (p planAppenderRecovery) ApplyWithConditionIfEmpty(c planBuilderCondition, pb planBuilder) (r PlanAppender) {
	return p.create(func(in PlanAppender) PlanAppender {
		return in.ApplyWithConditionIfEmpty(c, pb)
	})
}

func (p planAppenderRecovery) ApplySubPlanIfEmpty(pb planBuilderSubPlan, plans ...planBuilder) (r PlanAppender) {
	return p.create(func(in PlanAppender) PlanAppender {
		return in.ApplySubPlanIfEmpty(pb, plans...)
	})
}

func (p planAppenderRecovery) Plan() api.Plan {
	return p.appender.Plan()
}

type planAppenderType struct {
	pb      WithPlanBuilder
	current api.Plan

	backoff api.BackOff
}

func (p *planAppenderType) BackOff() api.BackOff {
	return p.backoff.DeepCopy()
}

func (p *planAppenderType) ApplyWithBackOff(key api.BackOffKey, delay time.Duration, pb planBuilder) PlanAppender {
	if !p.backoff.Process(key) {
		return p
	}

	p.backoff = p.backoff.BackOff(key, delay)

	return p.Apply(pb)
}

func (p *planAppenderType) ApplyIfEmptyWithBackOff(key api.BackOffKey, delay time.Duration, pb planBuilder) PlanAppender {
	if p.current.IsEmpty() {
		return p.ApplyWithBackOff(key, delay, pb)
	}
	return p
}

func (p *planAppenderType) Plan() api.Plan {
	return p.current
}

func (p *planAppenderType) ApplyIfEmpty(pb planBuilder) PlanAppender {
	if p.current.IsEmpty() {
		return p.Apply(pb)
	}
	return p
}

func (p *planAppenderType) ApplyWithConditionIfEmpty(c planBuilderCondition, pb planBuilder) PlanAppender {
	if p.current.IsEmpty() {
		return p.ApplyWithCondition(c, pb)
	}
	return p
}

func (p *planAppenderType) ApplySubPlanIfEmpty(pb planBuilderSubPlan, plans ...planBuilder) PlanAppender {
	if p.current.IsEmpty() {
		return p.ApplySubPlan(pb, plans...)
	}
	return p
}

func (p *planAppenderType) new(plan api.Plan) *planAppenderType {
	return &planAppenderType{
		pb:      p.pb,
		current: append(p.current, plan...),
		backoff: p.backoff.DeepCopy(),
	}
}

func (p *planAppenderType) Apply(pb planBuilder) PlanAppender {
	return p.new(p.pb.Apply(pb))
}

func (p *planAppenderType) ApplyWithCondition(c planBuilderCondition, pb planBuilder) PlanAppender {
	return p.new(p.pb.ApplyWithCondition(c, pb))
}

func (p *planAppenderType) ApplySubPlan(pb planBuilderSubPlan, plans ...planBuilder) PlanAppender {
	return p.new(p.pb.ApplySubPlan(pb, plans...))
}
