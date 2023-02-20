package reconcile

import (
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
)

func TestIsActionTimeout(t *testing.T) {
	type testCase struct {
		timeout        api.Timeout
		action         api.Action
		expectedResult bool
	}

	timeFiveMinutesAgo := meta.Time{
		Time: time.Now().Add(-time.Hour),
	}

	testCases := map[string]testCase{
		"nil start time": {
			timeout:        api.Timeout{},
			action:         api.Action{},
			expectedResult: false,
		},
		"infinite timeout": {
			timeout:        api.NewTimeout(0),
			action:         api.Action{},
			expectedResult: false,
		},
		"timeouted case": {
			timeout: api.NewTimeout(time.Minute),
			action: api.Action{
				StartTime: &timeFiveMinutesAgo,
			},
			expectedResult: true,
		},
		"still in progress case": {
			timeout: api.NewTimeout(time.Minute * 10),
			action: api.Action{
				StartTime: &timeFiveMinutesAgo,
			},
			expectedResult: true,
		},
	}

	for n, c := range testCases {
		t.Run(n, func(t *testing.T) {
			require.Equal(t, c.expectedResult, isActionTimeout(c.timeout, c.action))
		})
	}
}
