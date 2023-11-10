package crds

import (
	"github.com/stretchr/testify/require"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"testing"
)

func EnsureWithoutValidation(t *testing.T, v v1.CustomResourceDefinitionVersion) {
	t.Run(v.Name, func(t *testing.T) {
		require.NotNil(t, v.Schema)
		require.NotNil(t, v.Schema.OpenAPIV3Schema)

		require.Equal(t, "object", v.Schema.OpenAPIV3Schema.Type)
		require.NotNil(t, v.Schema.OpenAPIV3Schema.XPreserveUnknownFields)
		require.True(t, *v.Schema.OpenAPIV3Schema.XPreserveUnknownFields)
		require.Nil(t, v.Schema.OpenAPIV3Schema.Properties)
	})
}

func EnsureWithValidation(t *testing.T, v v1.CustomResourceDefinitionVersion, preserve bool) {
	t.Run(v.Name, func(t *testing.T) {
		require.NotNil(t, v.Schema)
		require.NotNil(t, v.Schema.OpenAPIV3Schema)

		require.Equal(t, "object", v.Schema.OpenAPIV3Schema.Type)
		require.NotNil(t, v.Schema.OpenAPIV3Schema.XPreserveUnknownFields)
		if preserve {
			require.True(t, *v.Schema.OpenAPIV3Schema.XPreserveUnknownFields)
		} else {
			require.False(t, *v.Schema.OpenAPIV3Schema.XPreserveUnknownFields)
		}
		require.NotNil(t, v.Schema.OpenAPIV3Schema.Properties)
	})
}
