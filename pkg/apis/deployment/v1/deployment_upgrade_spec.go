package v1

type DeploymentUpgradeSpec struct {
	// Flag specify if upgrade should be auto-injected, even if is not required (in case of stuck)
	AutoUpgrade bool `json:"autoUpgrade"`
}

func (d *DeploymentUpgradeSpec) Get() DeploymentUpgradeSpec {
	if d == nil {
		return DeploymentUpgradeSpec{}
	}

	return *d
}
