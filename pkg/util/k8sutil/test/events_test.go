package k8sutil_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/arangodb/k8s-operator/pkg/apis/arangodb/v1alpha"
	"github.com/arangodb/k8s-operator/pkg/util/k8sutil"
)

var apiObjectForTest = api.ArangoDeployment{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "Willy",
		Namespace: "Wonka",
	},
	Spec: api.DeploymentSpec{
		Mode: api.DeploymentModeCluster,
	},
}

func TestMemberAddEvent(t *testing.T) {
	event := k8sutil.NewMemberAddEvent("name", "role", &apiObjectForTest)
	assert.Equal(t, event.Reason, "New Role Added")
	assert.Equal(t, event.Message, "New role name added to deployment")
	assert.Equal(t, event.Type, v1.EventTypeNormal)
}

func TestMemberRemoveEvent(t *testing.T) {
	event := k8sutil.NewMemberRemoveEvent("name", "role", &apiObjectForTest)
	assert.Equal(t, event.Reason, "Role Removed")
	assert.Equal(t, event.Message, "Existing role name removed from the deployment")
	assert.Equal(t, event.Type, v1.EventTypeNormal)
}

func TestPodGoneEvent(t *testing.T) {
	event := k8sutil.NewPodGoneEvent("name", "role", &apiObjectForTest)
	assert.Equal(t, event.Reason, "Pod Of Role Gone")
	assert.Equal(t, event.Message, "Pod name of member role is gone")
	assert.Equal(t, event.Type, v1.EventTypeNormal)
}

func TestImmutableFieldEvent(t *testing.T) {
	event := k8sutil.NewImmutableFieldEvent("name", &apiObjectForTest)
	assert.Equal(t, event.Reason, "Immutable Field Change")
	assert.Equal(t, event.Message, "Changing field name is not possible. It has been reset to its original value.")
	assert.Equal(t, event.Type, v1.EventTypeNormal)
}

func TestErrorEvent(t *testing.T) {
	event := k8sutil.NewErrorEvent("reason", errors.New("something"), &apiObjectForTest)
	assert.Equal(t, event.Reason, "Reason")
	assert.Equal(t, event.Type, v1.EventTypeWarning)
}

// // not accessible outside the package
// func TestDeploymentEvent(t *testing.T) {
// 	event := k8sutil.New("member name", "role", &apiObjectForTest)
// 	assert.Equal(t, event.Type, v1.EventTypeNormal)
// }
