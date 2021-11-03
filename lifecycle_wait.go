package main

import (
	"fmt"
	"os"
	"time"

	"github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/constants"

	"github.com/spf13/cobra"
)

const (
	ArgDeploymentWatchTimeout = "watch-timeout"
	WatchDefaultTimeout       = time.Minute * 5
	WatchCheckInterval        = time.Second * 5
)

var (
	cmdLifecycleWait = &cobra.Command{
		Use:   "wait",
		Short: "Wait for specific ArangoDeployment",
		Long:  "Wait for ArangoDeployment till it will reach UpToDate condition",
		Run:   cmdLifecycleWaitCheck,
	}
)

func init() {
	var deploymentName string
	var watchTimeout time.Duration

	cmdLifecycleWait.Flags().StringVarP(&deploymentName, ArgDeploymentName, "d", "",
		"Name of ArangoDeployment to watch - necessary when more than one deployment exist within on namespace")
	cmdLifecycleWait.Flags().DurationVarP(&watchTimeout, ArgDeploymentWatchTimeout, "t", WatchDefaultTimeout,
		"Watch timeout")
}

func cmdLifecycleWaitCheck(cmd *cobra.Command, _ []string) {
	ctx := getInterruptionContext()

	deploymentName, err := cmd.Flags().GetString(ArgDeploymentName)
	if err != nil {
		cliLog.Fatal().Err(err).Msg(fmt.Sprintf("error parsing argument: %s", ArgDeploymentName))
	}
	watchTimeout, err := cmd.Flags().GetDuration(ArgDeploymentWatchTimeout)
	if err != nil {
		cliLog.Fatal().Err(err).Msg(fmt.Sprintf("error parsing argument: %s", ArgDeploymentWatchTimeout))
	}

	for {
		d, err := getDeployment(ctx, os.Getenv(constants.EnvOperatorPodNamespace), deploymentName)
		if err != nil {
			cliLog.Fatal().Err(err).Msg(fmt.Sprintf("error getting ArangoDeployment: %s", d.Name))
		}

		isUpToDate, err := d.IsUpToDate()
		if err != nil {
			cliLog.Err(err).Msg(fmt.Sprintf("error checking Status for ArangoDeployment: %s", d.Name))
		}

		if isUpToDate {
			cliLog.Info().Msg(fmt.Sprintf("ArangoDeployment: %s is %s", d.Name, v1.ConditionTypeUpToDate))
			return
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(WatchCheckInterval):
			cliLog.Info().Msg(fmt.Sprintf("ArangoDeployment: %s is not ready yet. Waiting...", d.Name))
			continue
		case <-time.After(watchTimeout):
			cliLog.Error().Msg(fmt.Sprintf("ArangoDeployment: %s is not %s yet - operation timed out!", d.Name, v1.ConditionTypeUpToDate))
			return
		}
	}
}
