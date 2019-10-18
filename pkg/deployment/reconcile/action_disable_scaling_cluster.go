package reconcile

import (
	"context"
	"time"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/rs/zerolog"
)

// actionDisableScalingCluster implements disabling scaling DBservers and coordinators.
type actionDisableScalingCluster struct {
	log         zerolog.Logger
	action      api.Action
	actionCtx   ActionContext
	newMemberID string
}

// NewDisableScalingCluster creates the new action with disabling scaling DBservers and coordinators.
func NewDisableScalingCluster(log zerolog.Logger, action api.Action, actionCtx ActionContext) Action {
	return &actionDisableScalingCluster{
		log:       log,
		action:    action,
		actionCtx: actionCtx,
	}
}

// Start disables scaling DBservers and coordinators
func (a *actionDisableScalingCluster) Start(ctx context.Context) (bool, error) {
	err := a.actionCtx.DisableScalingCluster()
	if err != nil {
		return false, err
	}
	return true, nil
}

// CheckProgress does not matter. Everything is done in 'Start' function
func (a *actionDisableScalingCluster) CheckProgress(ctx context.Context) (bool, bool, error) {
	return true, false, nil
}

// Timeout does not matter. Everything is done in 'Start' function
func (a *actionDisableScalingCluster) Timeout() time.Duration {
	return 0
}

// MemberID is not used
func (a *actionDisableScalingCluster) MemberID() string {
	return ""
}
