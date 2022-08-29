//
// DISCLAIMER
//
// Copyright 2016-2022 ArangoDB GmbH, Cologne, Germany
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// Copyright holder is ArangoDB GmbH, Cologne, Germany
//

package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	v1 "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/constants"
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
		"Name of ArangoDeployment to watch - necessary when more than one deployment exist within one namespace")
	cmdLifecycleWait.Flags().DurationVarP(&watchTimeout, ArgDeploymentWatchTimeout, "t", WatchDefaultTimeout,
		"Watch timeout")
}

func cmdLifecycleWaitCheck(cmd *cobra.Command, _ []string) {
	ctx := util.CreateSignalContext(context.Background())

	deploymentName, err := cmd.Flags().GetString(ArgDeploymentName)
	if err != nil {
		logger.Err(err).Fatal("error parsing argument: %s", ArgDeploymentName)
	}
	watchTimeout, err := cmd.Flags().GetDuration(ArgDeploymentWatchTimeout)
	if err != nil {
		logger.Err(err).Fatal("error parsing argument: %s", ArgDeploymentWatchTimeout)
	}

	for {
		d, err := getDeployment(ctx, os.Getenv(constants.EnvOperatorPodNamespace), deploymentName)
		if err != nil {
			logger.Err(err).Fatal(fmt.Sprintf("error getting ArangoDeployment: %s", d.Name))
		}

		isUpToDate, err := d.IsUpToDate()
		if err != nil {
			logger.Err(err).Error(fmt.Sprintf("error checking Status for ArangoDeployment: %s", d.Name))
		}

		if isUpToDate {
			logger.Info(fmt.Sprintf("ArangoDeployment: %s is %s", d.Name, v1.ConditionTypeUpToDate))
			return
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(WatchCheckInterval):
			logger.Info("ArangoDeployment: %s is not ready yet. Waiting...", d.Name)
			continue
		case <-time.After(watchTimeout):
			logger.Error("ArangoDeployment: %s is not %s yet - operation timed out!", d.Name, v1.ConditionTypeUpToDate)
			return
		}
	}
}
