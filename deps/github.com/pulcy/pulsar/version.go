package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

const (
	hdr = "   ___       __            " + "\n" +
		"  / _ \\__ __/ /__ ___ _____" + "\n" +
		" / ___/ // / (_-</ _ `/ __/" + "\n" +
		"/_/   \\_,_/_/___/\\_,_/_/   " + "\n"
)

var (
	versionQuiet bool
	cmdVersion   = &cobra.Command{
		Use: "version",
		Run: showVersion,
	}
	cmdVersionBuild = &cobra.Command{
		Use: "build",
		Run: showVersionBuild,
	}
)

func init() {
	cmdVersion.Flags().BoolVarP(&versionQuiet, "quiet", "q", false, "If set, print only the version identifier")
	mainCmd.AddCommand(cmdVersion)
	cmdVersion.AddCommand(cmdVersionBuild)
}

func showVersion(cmd *cobra.Command, args []string) {
	if versionQuiet {
		fmt.Printf("%s-%s\n", projectVersion, projectBuild)
	} else {
		for _, line := range strings.Split(hdr, "\n") {
			log.Info(line)
		}
		log.Info("%s %s, build %s\n", mainCmd.Use, projectVersion, projectBuild)
	}
}

func showVersionBuild(cmd *cobra.Command, args []string) {
	fmt.Printf("%s\n", projectBuild)
}
