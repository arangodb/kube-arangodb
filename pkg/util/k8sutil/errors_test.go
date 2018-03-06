package k8sutil

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

var (
	conflictError = apierrors.NewConflict(schema.GroupResource{"groupName", "resourceName"}, "something", os.ErrInvalid)
	existsError   = apierrors.NewAlreadyExists(schema.GroupResource{"groupName", "resourceName"}, "something")
	invalidError  = apierrors.NewInvalid(schema.GroupKind{"groupName", "kindName"}, "something", field.ErrorList{})
	notFoundError = apierrors.NewNotFound(schema.GroupResource{"groupName", "resourceName"}, "something")
)

func TestIsAlreadyExists(t *testing.T) {
	assert.False(t, IsAlreadyExists(conflictError))
	assert.True(t, IsAlreadyExists(existsError))
}

func TestIsConflict(t *testing.T) {
	assert.False(t, IsConflict(existsError))
	assert.True(t, IsConflict(conflictError))
}

func TestIsNotFound(t *testing.T) {
	assert.False(t, IsNotFound(invalidError))
	assert.True(t, IsNotFound(notFoundError))
}
