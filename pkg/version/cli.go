package version

import (
	"github.com/spf13/cobra"

	"github.com/arangodb/kube-arangodb/pkg/util/license"
)

func Command() *cobra.Command {
	var l = cobra.Command{
		Use:   "version",
		Short: "Show the version",
		Long:  "Show the version",
		Run:   versionRun,
	}

	license.RegisterCommand(&l)

	return &l
}

func versionRun(cmd *cobra.Command, args []string) {
	println(GetVersionV1().String())
}
