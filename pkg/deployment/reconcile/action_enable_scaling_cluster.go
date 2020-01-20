package reconcile

import (
	"context"
	"time"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/rs/zerolog"
)

// actionEnableScalingCluster implements enabling scaling DBservers and coordinators.
type actionEnableScalingCluster struct {
	log         zerolog.Logger
	action      api.Action
	actionCtx   ActionContext
	newMemberID string
}

// NewEnableScalingCluster creates the new action with enabling scaling DBservers and coordinators.
func NewEnableScalingCluster(log zerolog.Logger, action api.Action, actionCtx ActionContext) Action {
	return &actionEnableScalingCluster{
		log:       log,
		action:    action,
		actionCtx: actionCtx,
	}
}

// Start enables scaling DBservers and coordinators
func (a *actionEnableScalingCluster) Start(ctx context.Context) (bool, error) {
	err := a.actionCtx.EnableScalingCluster()
	if err != nil {
		return false, err
	}
	return true, nil
}

// CheckProgress does not matter. Everything is done in 'Start' function
func (a *actionEnableScalingCluster) CheckProgress(ctx context.Context) (bool, bool, error) {
	return true, false, nil
}

// Timeout does not matter. Everything is done in 'Start' function
func (a *actionEnableScalingCluster) Timeout() time.Duration {
	return 0
}

// MemberID is not used
func (a *actionEnableScalingCluster) MemberID() string {
	return ""
}
