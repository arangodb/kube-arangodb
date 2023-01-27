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
	"time"

	"github.com/dchest/uniuri"
	"k8s.io/apimachinery/pkg/api/equality"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/uuid"

	"github.com/arangodb/kube-arangodb/pkg/util"
)

// ActionPriority define action priority
type ActionPriority int

const (
	// ActionPriorityNormal define Normal priority plan
	ActionPriorityNormal ActionPriority = iota
	// ActionPriorityHigh define High priority plan
	ActionPriorityHigh
	// ActionPriorityResource define Resource priority plan
	ActionPriorityResource
	ActionPriorityUnknown
)

// ActionType is a strongly typed name for a plan action item
type ActionType string

func (a ActionType) String() string {
	return string(a)
}

func GetActionPriority(a ActionType) ActionPriority {
	return a.Priority()
}

func ActionDefaultTimeout(a ActionType) time.Duration {
	return a.DefaultTimeout()
}

const (
	// MemberIDPreviousAction is used for Action.MemberID when the MemberID
	// should be derived from the previous action.
	MemberIDPreviousAction = "@previous"
)

const (
	ParamPodUID = "PodUID"
)

// Action represents a single action to be taken to update a deployment.
type Action struct {
	// ID of this action (unique for every action)
	ID string `json:"id"`
	// SetID define the unique ID of current action set
	SetID types.UID `json:"setID,omitempty"`
	// Type of action.
	Type ActionType `json:"type"`
	// ID reference of the member involved in this action (if any)
	MemberID string `json:"memberID,omitempty"`
	// Group involved in this action
	Group ServerGroup `json:"group,omitempty"`
	// CreationTime is set the when the action is created.
	CreationTime meta.Time `json:"creationTime"`
	// StartTime is set the when the action has been started, but needs to wait to be finished.
	StartTime *meta.Time `json:"startTime,omitempty"`
	// Reason for this action
	Reason string `json:"reason,omitempty"`
	// Image used in can of a SetCurrentImage action.
	Image string `json:"image,omitempty"`
	// Params additional parameters used for action
	Params map[string]string `json:"params,omitempty"`
	// Locals additional storage for local variables which are produced during the action.
	Locals PlanLocals `json:"locals,omitempty"`
	// ID reference of the task involved in this action (if any)
	TaskID types.UID `json:"taskID,omitempty"`
	// Architecture of the member involved in this action (if any)
	Architecture ArangoDeploymentArchitectureType `json:"arch,omitempty"`
	// Progress describes what is a status of the current action.
	Progress string `json:"progress,omitempty"`
}

// Equal compares two Actions
func (a Action) Equal(other Action) bool {
	return a.ID == other.ID &&
		a.Type == other.Type &&
		a.SetID == other.SetID &&
		a.MemberID == other.MemberID &&
		a.Group == other.Group &&
		util.TimeCompareEqual(a.CreationTime, other.CreationTime) &&
		util.TimeCompareEqualPointer(a.StartTime, other.StartTime) &&
		a.Reason == other.Reason &&
		a.Image == other.Image &&
		equality.Semantic.DeepEqual(a.Params, other.Params) &&
		a.Locals.Equal(other.Locals) &&
		a.TaskID == other.TaskID &&
		a.Architecture == other.Architecture &&
		a.Progress == other.Progress
}

// AddParam returns copy of action with set parameter
func (a Action) AddParam(key, value string) Action {
	if a.Params == nil {
		a.Params = map[string]string{}
	}

	a.Params[key] = value

	return a
}

// GetParam returns action parameter
func (a Action) GetParam(key string) (string, bool) {
	if a.Params == nil {
		return "", false
	}

	i, ok := a.Params[key]

	return i, ok
}

// NewActionSet add new SetID vale to the actions
func NewActionSet(actions ...Action) []Action {
	sid := uuid.NewUUID()
	for id := range actions {
		actions[id].SetID = sid
	}
	return actions
}

// NewAction instantiates a new Action.
func NewAction(actionType ActionType, group ServerGroup, memberID string, reason ...string) Action {
	a := Action{
		ID:           uniuri.New(),
		Type:         actionType,
		MemberID:     memberID,
		Group:        group,
		CreationTime: meta.Now(),
	}
	if len(reason) != 0 {
		a.Reason = reason[0]
	}
	return a
}

// ActionBuilder allows to generate actions based on predefined group and member id
type ActionBuilder interface {
	// NewAction instantiates a new Action.
	NewAction(actionType ActionType, reason ...string) Action

	// Group returns ServerGroup for this builder
	Group() ServerGroup

	// MemberID returns Member ID for this builder
	MemberID() string
}

type actionBuilder struct {
	group    ServerGroup
	memberID string
}

func (a actionBuilder) NewAction(actionType ActionType, reason ...string) Action {
	return NewAction(actionType, a.group, a.memberID, reason...)
}

func (a actionBuilder) Group() ServerGroup {
	return a.group
}

func (a actionBuilder) MemberID() string {
	return a.memberID
}

// NewActionBuilder create new action builder with provided group and id
func NewActionBuilder(group ServerGroup, memberID string) ActionBuilder {
	return actionBuilder{
		group:    group,
		memberID: memberID,
	}
}

// SetImage sets the Image field to the given value and returns the modified
// action.
func (a Action) SetImage(image string) Action {
	a.Image = image
	return a
}

// SetArch sets the Architecture field to the given value and returns the modified
func (a Action) SetArch(arch ArangoDeploymentArchitectureType) Action {
	a.Architecture = arch
	return a
}

// IsStarted returns true if the action has been started already.
func (a Action) IsStarted() bool {
	return !a.StartTime.IsZero()
}

// AsPlan parse action list into plan
func AsPlan(a []Action) Plan {
	return a
}

// Plan is a list of actions that will be taken to update a deployment.
// Only 1 action is in progress at a time. The operator will wait for that
// action to be completely and then remove the action.
type Plan []Action

// Equal compares two Plan
func (p Plan) Equal(other Plan) bool {
	// For plan the order is relevant!
	if len(p) != len(other) {
		return false
	}

	for i := 0; i < len(p); i++ {
		if !p[i].Equal(other[i]) {
			return false
		}
	}

	return true
}

// IsEmpty checks if plan is empty
func (p Plan) IsEmpty() bool {
	return len(p) == 0
}

func (p Plan) NonInternalActions() int {
	var z int

	for id := range p {
		if !p[id].Type.Internal() {
			z++
		}
	}

	return z
}

// After add action at the end of plan
func (p Plan) After(action ...Action) Plan {
	n := Plan{}

	n = append(n, p...)

	n = append(n, action...)

	return n
}

// Before add action at the beginning of plan
func (p Plan) Before(action ...Action) Plan {
	n := Plan{}

	n = append(n, action...)

	n = append(n, p...)

	return n
}

// Wrap wraps plan with actions
func (p Plan) Wrap(before, after Action) Plan {
	n := Plan{}

	n = append(n, before)

	n = append(n, p...)

	n = append(n, after)

	return n
}

// AfterFirst adds actions when condition will return false
func (p Plan) AfterFirst(condition func(a Action) bool, actions ...Action) Plan {
	var r Plan
	c := p
	for {
		if len(c) == 0 {
			break
		}

		if !condition(c[0]) {
			r = append(r, actions...)

			r = append(r, c...)

			break
		}

		r = append(r, c[0])

		if len(c) == 1 {
			break
		}

		c = c[1:]
	}

	return r
}

// Filter filter list of the actions
func (p Plan) Filter(condition func(a Action) bool) Plan {
	var r Plan

	for _, a := range p {
		if condition(a) {
			r = append(r, a)
		}
	}

	return r
}
