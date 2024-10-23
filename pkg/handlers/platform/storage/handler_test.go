package storage

import (
	"testing"

	"github.com/stretchr/testify/require"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/arangodb/kube-arangodb/pkg/operatorV2/operation"
	"github.com/arangodb/kube-arangodb/pkg/util/tests"
)

func Test_ObjectNotFound(t *testing.T) {
	// Arrange
	handler := newFakeHandler()

	i := newItem(operation.Add, "test", "test")

	actions := map[operation.Operation]bool{
		operation.Add:    false,
		operation.Update: false,
		operation.Delete: false,
	}

	// Act
	for operation, shouldFail := range actions {
		t.Run(string(operation), func(t *testing.T) {
			err := tests.Handle(handler, i)

			// Assert
			if shouldFail {
				require.Error(t, err)
				require.True(t, apiErrors.IsNotFound(err))
			} else {
				require.NoError(t, err)
			}
		})
	}
}
