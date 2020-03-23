//
// DISCLAIMER
//
// Copyright 2020 ArangoDB GmbH, Cologne, Germany
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
// Author Adam Janikowski
//

package resources

import (
	"github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/jwt"
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/pod"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

type probeCheckBuilder struct {
	liveness, readiness probeBuilder
}

type probeBuilder func(spec api.DeploymentSpec, group api.ServerGroup, version driver.Version) (*k8sutil.HTTPProbeConfig, error)

func nilProbeBuilder(spec api.DeploymentSpec, group api.ServerGroup, version driver.Version) (*k8sutil.HTTPProbeConfig, error) {
	return nil, nil
}

func (r *Resources) getReadinessProbe(spec api.DeploymentSpec, group api.ServerGroup, version driver.Version) (*k8sutil.HTTPProbeConfig, error) {
	if !r.isReadinessProbeEnabled(spec, group, version) {
		return nil, nil
	}

	builders := r.probeBuilders()

	builder, ok := builders[group]
	if !ok {
		return nil, nil
	}

	config, err := builder.readiness(spec, group, version)
	if err != nil {
		return nil, err
	}

	groupSpec := spec.GetServerGroupSpec(group)

	if !groupSpec.HasProbesSpec() {
		return config, nil
	}

	probeSpec := groupSpec.GetProbesSpec()

	config.InitialDelaySeconds = probeSpec.ReadinessProbeSpec.GetInitialDelaySeconds(config.InitialDelaySeconds)
	config.PeriodSeconds = probeSpec.ReadinessProbeSpec.GetPeriodSeconds(config.PeriodSeconds)
	config.TimeoutSeconds = probeSpec.ReadinessProbeSpec.GetTimeoutSeconds(config.TimeoutSeconds)
	config.SuccessThreshold = probeSpec.ReadinessProbeSpec.GetSuccessThreshold(config.SuccessThreshold)
	config.FailureThreshold = probeSpec.ReadinessProbeSpec.GetFailureThreshold(config.FailureThreshold)

	return config, nil
}

func (r *Resources) getLivenessProbe(spec api.DeploymentSpec, group api.ServerGroup, version driver.Version) (*k8sutil.HTTPProbeConfig, error) {
	if !r.isLivenessProbeEnabled(spec, group, version) {
		return nil, nil
	}

	builders := r.probeBuilders()

	builder, ok := builders[group]
	if !ok {
		return nil, nil
	}

	config, err := builder.liveness(spec, group, version)
	if err != nil {
		return nil, err
	}

	groupSpec := spec.GetServerGroupSpec(group)

	if !groupSpec.HasProbesSpec() {
		return config, nil
	}

	probeSpec := groupSpec.GetProbesSpec()

	config.InitialDelaySeconds = probeSpec.LivenessProbeSpec.GetInitialDelaySeconds(config.InitialDelaySeconds)
	config.PeriodSeconds = probeSpec.LivenessProbeSpec.GetPeriodSeconds(config.PeriodSeconds)
	config.TimeoutSeconds = probeSpec.LivenessProbeSpec.GetTimeoutSeconds(config.TimeoutSeconds)
	config.SuccessThreshold = probeSpec.LivenessProbeSpec.GetSuccessThreshold(config.SuccessThreshold)
	config.FailureThreshold = probeSpec.LivenessProbeSpec.GetFailureThreshold(config.FailureThreshold)

	return config, nil
}

func (r *Resources) isReadinessProbeEnabled(spec api.DeploymentSpec, group api.ServerGroup, version driver.Version) bool {
	probe := pod.ReadinessSpec(group)

	groupSpec := spec.GetServerGroupSpec(group)

	if groupSpec.HasProbesSpec() {
		if p := groupSpec.GetProbesSpec().GetReadinessProbeDisabled(); p != nil {
			return !*p && probe.CanBeEnabled
		}
	}

	return probe.CanBeEnabled && probe.EnabledByDefault
}

func (r *Resources) isLivenessProbeEnabled(spec api.DeploymentSpec, group api.ServerGroup, version driver.Version) bool {
	probe := pod.LivenessSpec(group)

	groupSpec := spec.GetServerGroupSpec(group)

	if groupSpec.HasProbesSpec() {
		if p := groupSpec.GetProbesSpec().LivenessProbeDisabled; p != nil {
			return !*p && probe.CanBeEnabled
		}
	}

	return probe.CanBeEnabled && probe.EnabledByDefault
}

func (r *Resources) probeBuilders() map[api.ServerGroup]probeCheckBuilder {
	return map[api.ServerGroup]probeCheckBuilder{
		api.ServerGroupSingle: {
			liveness:  r.probeBuilderLivenessCore,
			readiness: r.probeBuilderReadinessCore,
		},
		api.ServerGroupAgents: {
			liveness:  r.probeBuilderLivenessCore,
			readiness: r.probeBuilderReadinessSimpleCore,
		},
		api.ServerGroupDBServers: {
			liveness:  r.probeBuilderLivenessCore,
			readiness: r.probeBuilderReadinessSimpleCore,
		},
		api.ServerGroupCoordinators: {
			liveness:  r.probeBuilderLivenessCore,
			readiness: r.probeBuilderReadinessCore,
		},
		api.ServerGroupSyncMasters: {
			liveness:  r.probeBuilderLivenessSync,
			readiness: nilProbeBuilder,
		},
		api.ServerGroupSyncWorkers: {
			liveness:  r.probeBuilderLivenessSync,
			readiness: nilProbeBuilder,
		},
	}
}

func (r *Resources) probeBuilderLivenessCore(spec api.DeploymentSpec, group api.ServerGroup, version driver.Version) (*k8sutil.HTTPProbeConfig, error) {
	authorization := ""
	if spec.IsAuthenticated() {
		secretData, err := r.getJWTSecret(spec)
		if err != nil {
			return nil, maskAny(err)
		}
		authorization, err = jwt.CreateArangodJwtAuthorizationHeaderAllowedPaths(secretData, "kube-arangodb", []string{"/_api/version"})
		if err != nil {
			return nil, maskAny(err)
		}
	}
	return &k8sutil.HTTPProbeConfig{
		LocalPath:     "/_api/version",
		Secure:        spec.IsSecure(),
		Authorization: authorization,
	}, nil
}

func (r *Resources) probeBuilderReadinessSimpleCore(spec api.DeploymentSpec, group api.ServerGroup, version driver.Version) (*k8sutil.HTTPProbeConfig, error) {
	p, err := r.probeBuilderLivenessCore(spec, group, version)
	if err != nil {
		return nil, err
	}

	if p == nil {
		return nil, nil
	}

	p.InitialDelaySeconds = 15
	p.PeriodSeconds = 10

	return p, nil
}

func (r *Resources) probeBuilderReadinessCore(spec api.DeploymentSpec, group api.ServerGroup, version driver.Version) (*k8sutil.HTTPProbeConfig, error) {
	localPath := "/_api/version"
	switch spec.GetMode() {
	case api.DeploymentModeActiveFailover:
		localPath = "/_admin/echo"
	}

	// /_admin/server/availability is the way to go, it is available since 3.3.9
	if version.CompareTo("3.3.9") >= 0 {
		localPath = "/_admin/server/availability"
	}

	authorization := ""
	if spec.IsAuthenticated() {
		secretData, err := r.getJWTSecret(spec)
		if err != nil {
			return nil, maskAny(err)
		}
		authorization, err = jwt.CreateArangodJwtAuthorizationHeaderAllowedPaths(secretData, "kube-arangodb", []string{localPath})
		if err != nil {
			return nil, maskAny(err)
		}
	}
	probeCfg := &k8sutil.HTTPProbeConfig{
		LocalPath:           localPath,
		Secure:              spec.IsSecure(),
		Authorization:       authorization,
		InitialDelaySeconds: 2,
		PeriodSeconds:       2,
	}

	return probeCfg, nil
}

func (r *Resources) probeBuilderLivenessSync(spec api.DeploymentSpec, group api.ServerGroup, version driver.Version) (*k8sutil.HTTPProbeConfig, error) {
	authorization := ""
	port := k8sutil.ArangoSyncMasterPort
	if group == api.ServerGroupSyncWorkers {
		port = k8sutil.ArangoSyncWorkerPort
	}
	if spec.Sync.Monitoring.GetTokenSecretName() != "" {
		// Use monitoring token
		token, err := r.getSyncMonitoringToken(spec)
		if err != nil {
			return nil, maskAny(err)
		}
		authorization = "bearer " + token
	} else if group == api.ServerGroupSyncMasters {
		// Fall back to JWT secret
		secretData, err := r.getSyncJWTSecret(spec)
		if err != nil {
			return nil, maskAny(err)
		}
		authorization, err = jwt.CreateArangodJwtAuthorizationHeaderAllowedPaths(secretData, "kube-arangodb", []string{"/_api/version"})
		if err != nil {
			return nil, maskAny(err)
		}
	} else {
		// Don't have a probe
		return nil, nil
	}
	return &k8sutil.HTTPProbeConfig{
		LocalPath:     "/_api/version",
		Secure:        spec.IsSecure(),
		Authorization: authorization,
		Port:          port,
	}, nil
}
