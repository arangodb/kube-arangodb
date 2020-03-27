package reconcile

import (
	"context"
	"time"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/rs/zerolog"
)

type actionEmptyCheckProgress struct {
}

// CheckProgress define optional check progress for action
// Returns: ready, abort, error.
func (e actionEmptyCheckProgress) CheckProgress(ctx context.Context) (bool, bool, error) {
	return true, false, nil
}

type actionEmptyStart struct {
}

func (e actionEmptyStart) Start(ctx context.Context) (bool, error) {
	return false, nil
}

func newActionImplDefRef(log zerolog.Logger, action api.Action, actionCtx ActionContext, timeout time.Duration) actionImpl {
	return newActionImpl(log, action, actionCtx, timeout, &action.MemberID)
}

func newActionImpl(log zerolog.Logger, action api.Action, actionCtx ActionContext, timeout time.Duration, memberIDRef *string) actionImpl {
	if memberIDRef == nil {
		panic("Action cannot have nil reference to member!")
	}

	return actionImpl{
		log:         log,
		action:      action,
		actionCtx:   actionCtx,
		timeout:     timeout,
		memberIDRef: memberIDRef,
	}
}

type actionImpl struct {
	log       zerolog.Logger
	action    api.Action
	actionCtx ActionContext

	timeout     time.Duration
	memberIDRef *string
}

// Timeout returns the amount of time after which this action will timeout.
func (a actionImpl) Timeout() time.Duration {
	return a.timeout
}

// Return the MemberID used / created in this action
func (a actionImpl) MemberID() string {
	return *a.memberIDRef
}
