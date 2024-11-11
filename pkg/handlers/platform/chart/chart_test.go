package chart

import (
	platformApi "github.com/arangodb/kube-arangodb/pkg/apis/platform/v1alpha1"
	"github.com/arangodb/kube-arangodb/pkg/operatorV2/operation"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/tests"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_ChartReconcile_EmptyChart(t *testing.T) {
	handler := newFakeHandler()

	// Arrange
	extension := tests.NewMetaObject[*platformApi.ArangoPlatformChart](t, tests.FakeNamespace, "example",
		func(t *testing.T, obj *platformApi.ArangoPlatformChart) {})
	extension_invalid_name := tests.NewMetaObject[*platformApi.ArangoPlatformChart](t, tests.FakeNamespace, "example-wrong-name",
		func(t *testing.T, obj *platformApi.ArangoPlatformChart) {})

	refresh := tests.CreateObjects(t, handler.kubeClient, handler.client, &extension, &extension_invalid_name)

	t.Run("Missing chart", func(t *testing.T) {
		// Test
		require.NoError(t, tests.Handle(handler, tests.NewItem(t, operation.Update, extension)))

		// Refresh
		refresh(t)

		// Validate
		require.False(t, extension.Status.Conditions.IsTrue(platformApi.SpecValidCondition))
		require.False(t, extension.Status.Conditions.IsTrue(platformApi.ReadyCondition))
	})

	t.Run("Invalid chart", func(t *testing.T) {
		// Arrange
		tests.Apply(t, extension, func(t *testing.T, obj *platformApi.ArangoPlatformChart) {
			obj.Spec.Definition = []byte("1234")
		})
		tests.UpdateObjects(t, handler.kubeClient, handler.client, &extension)

		// Test
		require.NoError(t, tests.Handle(handler, tests.NewItem(t, operation.Update, extension)))

		// Refresh
		refresh(t)

		// Validate
		require.True(t, extension.Status.Conditions.IsTrue(platformApi.SpecValidCondition))
		require.NotNil(t, extension.Status.Info)
		require.False(t, extension.Status.Info.Valid)
		require.EqualValues(t, extension.Status.Info.Message, "Chart is invalid")
		require.False(t, extension.Status.Conditions.IsTrue(platformApi.ReadyCondition))
	})

	t.Run("Invalid chart name", func(t *testing.T) {
		// Arrange
		tests.Apply(t, extension_invalid_name, func(t *testing.T, obj *platformApi.ArangoPlatformChart) {
			obj.Spec.Definition = chart_1_0
		})
		tests.UpdateObjects(t, handler.kubeClient, handler.client, &extension_invalid_name)

		// Test
		require.NoError(t, tests.Handle(handler, tests.NewItem(t, operation.Update, extension_invalid_name)))

		// Refresh
		refresh(t)

		// Validate
		require.True(t, extension_invalid_name.Status.Conditions.IsTrue(platformApi.SpecValidCondition))
		require.NotNil(t, extension_invalid_name.Status.Info)
		require.False(t, extension_invalid_name.Status.Info.Valid)
		require.EqualValues(t, extension_invalid_name.Status.Info.Message, "Chart Name mismatch")
		require.False(t, extension_invalid_name.Status.Conditions.IsTrue(platformApi.ReadyCondition))
	})

	t.Run("Valid chart 1.0.0", func(t *testing.T) {
		// Arrange
		tests.Apply(t, extension, func(t *testing.T, obj *platformApi.ArangoPlatformChart) {
			obj.Spec.Definition = chart_1_0
		})
		tests.UpdateObjects(t, handler.kubeClient, handler.client, &extension)

		// Test
		require.NoError(t, tests.Handle(handler, tests.NewItem(t, operation.Update, extension)))

		// Refresh
		refresh(t)

		// Validate
		require.True(t, extension.Status.Conditions.IsTrue(platformApi.SpecValidCondition))
		require.NotNil(t, extension.Status.Info)
		require.True(t, extension.Status.Info.Valid)
		require.EqualValues(t, extension.Status.Info.Message, "")
		require.NotNil(t, extension.Status.Info.Details)
		require.EqualValues(t, "example", extension.Status.Info.Details.GetName())
		require.EqualValues(t, "1.0.0", extension.Status.Info.Details.GetVersion())
		require.EqualValues(t, util.SHA256(chart_1_0), extension.Status.Info.Checksum)
		require.True(t, extension.Status.Conditions.IsTrue(platformApi.ReadyCondition))
	})

	t.Run("Valid chart 1.1.0", func(t *testing.T) {
		// Arrange
		tests.Apply(t, extension, func(t *testing.T, obj *platformApi.ArangoPlatformChart) {
			obj.Spec.Definition = chart_1_1
		})
		tests.UpdateObjects(t, handler.kubeClient, handler.client, &extension)

		// Test
		require.NoError(t, tests.Handle(handler, tests.NewItem(t, operation.Update, extension)))

		// Refresh
		refresh(t)

		// Validate
		require.True(t, extension.Status.Conditions.IsTrue(platformApi.SpecValidCondition))
		require.NotNil(t, extension.Status.Info)
		require.True(t, extension.Status.Info.Valid)
		require.EqualValues(t, extension.Status.Info.Message, "")
		require.NotNil(t, extension.Status.Info.Details)
		require.EqualValues(t, "example", extension.Status.Info.Details.GetName())
		require.EqualValues(t, "1.1.0", extension.Status.Info.Details.GetVersion())
		require.EqualValues(t, util.SHA256(chart_1_1), extension.Status.Info.Checksum)
		require.True(t, extension.Status.Conditions.IsTrue(platformApi.ReadyCondition))
	})
}
