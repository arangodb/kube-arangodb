package crds

import (
	"github.com/arangodb/kube-arangodb/pkg/util"
	"testing"
)

func Test_Apps_Job(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		crd := AppsJob()

		for _, v := range crd.Spec.Versions {
			EnsureWithoutValidation(t, v)
		}
	})
	t.Run("Nil Opts", func(t *testing.T) {
		crd := AppsJobWithOptions(nil)

		for _, v := range crd.Spec.Versions {
			EnsureWithoutValidation(t, v)
		}
	})
	t.Run("Empty Opts", func(t *testing.T) {
		crd := AppsJobWithOptions(&CRDOptions{})

		for _, v := range crd.Spec.Versions {
			EnsureWithoutValidation(t, v)
		}
	})
	t.Run("Without schema", func(t *testing.T) {
		crd := AppsJobWithOptions(&CRDOptions{
			WithSchema: util.NewType(false),
		})

		for _, v := range crd.Spec.Versions {
			EnsureWithoutValidation(t, v)
		}
	})
	t.Run("With schema", func(t *testing.T) {
		crd := AppsJobWithOptions(&CRDOptions{
			WithSchema: util.NewType(true),
		})

		for _, v := range crd.Spec.Versions {
			EnsureWithValidation(t, v, true)
		}
	})
}
