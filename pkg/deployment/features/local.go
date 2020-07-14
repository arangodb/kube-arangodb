package features

import (
	"fmt"
	"github.com/spf13/cobra"
	"sync"
)

var features map[string] Feature = map[string]Feature{}
var featuresLock sync.Mutex

func registerFeature(f Feature) {
	featuresLock.Lock()
	defer featuresLock.Unlock()

	if f == nil {
		panic("Feature cannot be nil")
	}

	if _, ok := features[f.Name()]; ok {
		panic("Feature already registered")
	}

	features[f.Name()] = f
}

var internalCMD = &cobra.Command{
	Use: "features",
	Short: "Describe all operator features",
	Run: cmdRun,
}

func Init(cmd *cobra.Command) {
	featuresLock.Lock()
	defer featuresLock.Unlock()

	cmd.AddCommand(internalCMD)

	f := cmd.Flags()

	for _, feature := range features {
		z := ""

		if v := feature.Version(); v != "" || feature.EnterpriseRequired() {
			if v != "" && feature.EnterpriseRequired() {
				z = fmt.Sprintf("%s - Required version %s and Enterprise Edition", feature.Description(), v)
			} else if v != "" {
				z = fmt.Sprintf("%s. Required version %s", feature.Description(), v)
			} else if feature.EnterpriseRequired() {
				z = fmt.Sprintf("%s - Required Enterprise Edition", feature.Description())
			} else {
				z = feature.Description()
			}
		}

		f.BoolVar(feature.EnabledPointer(), fmt.Sprintf("deployment.feature.%s", feature.Name()), feature.EnabledByDefault(), z)
	}
}

func cmdRun(cmd *cobra.Command, args []string) {
	featuresLock.Lock()
	defer featuresLock.Unlock()

	for _, feature := range features {
		println(fmt.Sprintf("Feature: %s", feature.Name()))

		println()
	}
}