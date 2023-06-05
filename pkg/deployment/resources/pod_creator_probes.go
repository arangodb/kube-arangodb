//
// DISCLAIMER
//
// Copyright 2016-2023 ArangoDB GmbH, Cologne, Germany
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

package resources

import (
	"math"
	"time"

	core "k8s.io/api/core/v1"

	"github.com/arangodb/go-driver/jwt"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/features"
	"github.com/arangodb/kube-arangodb/pkg/deployment/pod"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/probes"
)

type Probe interface {
	Create() *core.Probe

	SetSpec(spec *api.ServerGroupProbeSpec)
}

type probeCheckBuilder struct {
	liveness, readiness, startup probeBuilder
}

type probeBuilder func(spec api.DeploymentSpec, group api.ServerGroup, _ api.ImageInfo) (Probe, error)

func nilProbeBuilder(_ api.DeploymentSpec, _ api.ServerGroup, _ api.ImageInfo) (Probe, error) {
	return nil, nil
}

func (r *Resources) getReadinessProbe(spec api.DeploymentSpec, group api.ServerGroup, imageInfo api.ImageInfo) (Probe, error) {
	if !r.isReadinessProbeEnabled(spec, group) {
		return nil, nil
	}

	builders := r.probeBuilders(imageInfo)

	builder, ok := builders[group]
	if !ok {
		return nil, nil
	}

	config, err := builder.readiness(spec, group, imageInfo)
	if err != nil {
		return nil, err
	}

	groupSpec := spec.GetServerGroupSpec(group)

	if !groupSpec.HasProbesSpec() {
		return config, nil
	}

	probeSpec := groupSpec.GetProbesSpec()

	config.SetSpec(probeSpec.ReadinessProbeSpec)

	return config, nil
}

func (r *Resources) getLivenessProbe(spec api.DeploymentSpec, group api.ServerGroup, imageInfo api.ImageInfo) (Probe, error) {
	if !r.isLivenessProbeEnabled(spec, group) {
		return nil, nil
	}

	builders := r.probeBuilders(imageInfo)

	builder, ok := builders[group]
	if !ok {
		return nil, nil
	}

	config, err := builder.liveness(spec, group, imageInfo)
	if err != nil {
		return nil, err
	}

	groupSpec := spec.GetServerGroupSpec(group)

	if !groupSpec.HasProbesSpec() {
		return config, nil
	}

	probeSpec := groupSpec.GetProbesSpec()

	config.SetSpec(probeSpec.LivenessProbeSpec)

	return config, nil
}

func (r *Resources) getStartupProbe(spec api.DeploymentSpec, group api.ServerGroup, imageInfo api.ImageInfo) (Probe, error) {
	if !r.isStartupProbeEnabled(spec, group) {
		return nil, nil
	}

	builders := r.probeBuilders(imageInfo)

	builder, ok := builders[group]
	if !ok {
		return nil, nil
	}

	config, err := builder.startup(spec, group, imageInfo)
	if err != nil {
		return nil, err
	}

	groupSpec := spec.GetServerGroupSpec(group)

	if !groupSpec.HasProbesSpec() {
		return config, nil
	}

	probeSpec := groupSpec.GetProbesSpec()

	config.SetSpec(probeSpec.StartupProbeSpec)

	return config, nil
}

func (r *Resources) isReadinessProbeEnabled(spec api.DeploymentSpec, group api.ServerGroup) bool {
	probe := pod.ReadinessSpec(group)

	groupSpec := spec.GetServerGroupSpec(group)

	if groupSpec.HasProbesSpec() {
		if p := groupSpec.GetProbesSpec().GetReadinessProbeDisabled(); p != nil {
			return !*p && probe.CanBeEnabled
		}
	}

	return probe.CanBeEnabled && probe.EnabledByDefault
}

func (r *Resources) isLivenessProbeEnabled(spec api.DeploymentSpec, group api.ServerGroup) bool {
	probe := pod.LivenessSpec(group)

	groupSpec := spec.GetServerGroupSpec(group)

	if groupSpec.HasProbesSpec() {
		if p := groupSpec.GetProbesSpec().LivenessProbeDisabled; p != nil {
			return !*p && probe.CanBeEnabled
		}
	}

	return probe.CanBeEnabled && probe.EnabledByDefault
}

func (r *Resources) isStartupProbeEnabled(spec api.DeploymentSpec, group api.ServerGroup) bool {
	probe := pod.StartupSpec(group)

	groupSpec := spec.GetServerGroupSpec(group)

	if groupSpec.HasProbesSpec() {
		if p := groupSpec.GetProbesSpec().StartupProbeDisabled; p != nil {
			return !*p && probe.CanBeEnabled
		}
	}

	return probe.CanBeEnabled && probe.EnabledByDefault
}

func (r *Resources) probeBuilders(imageInfo api.ImageInfo) map[api.ServerGroup]probeCheckBuilder {
	return map[api.ServerGroup]probeCheckBuilder{
		api.ServerGroupSingle: {
			startup:   r.probeBuilderStartupCoreSelect(api.ServerGroupSingle, imageInfo),
			liveness:  r.probeBuilderLivenessCoreSelect(api.ServerGroupSingle, imageInfo),
			readiness: r.probeBuilderReadinessCoreSelect(),
		},
		api.ServerGroupAgents: {
			startup:   r.probeBuilderStartupCoreSelect(api.ServerGroupAgents, imageInfo),
			liveness:  r.probeBuilderLivenessCoreSelect(api.ServerGroupAgents, imageInfo),
			readiness: r.probeBuilderReadinessSimpleCoreSelect(),
		},
		api.ServerGroupDBServers: {
			startup:   r.probeBuilderStartupCoreSelect(api.ServerGroupDBServers, imageInfo),
			liveness:  r.probeBuilderLivenessCoreSelect(api.ServerGroupDBServers, imageInfo),
			readiness: r.probeBuilderReadinessSimpleCoreSelect(),
		},
		api.ServerGroupCoordinators: {
			startup:   r.probeBuilderStartupCoreSelect(api.ServerGroupCoordinators, imageInfo),
			liveness:  r.probeBuilderLivenessCoreSelect(api.ServerGroupCoordinators, imageInfo),
			readiness: r.probeBuilderReadinessCoreSelect(),
		},
		api.ServerGroupSyncMasters: {
			startup:   r.probeBuilderStartupSync,
			liveness:  r.probeBuilderLivenessSync,
			readiness: nilProbeBuilder,
		},
		api.ServerGroupSyncWorkers: {
			startup:   r.probeBuilderStartupSync,
			liveness:  r.probeBuilderLivenessSync,
			readiness: nilProbeBuilder,
		},
	}
}

func (r *Resources) probeCommand(spec api.DeploymentSpec, probeType api.ProbeType) []string {
	exePath := k8sutil.LifecycleBinary()
	args := []string{
		exePath,
		"lifecycle",
		"probe",
		string(probeType),
	}

	if spec.IsSecure() {
		args = append(args, "--ssl")
	}

	if spec.IsAuthenticated() {
		args = append(args, "--auth")
	}

	return args
}

func (r *Resources) probeBuilderLivenessCoreSelect(group api.ServerGroup, imageInfo api.ImageInfo) probeBuilder {
	if features.JWTRotation().Enabled() || IsServerProgressAvailable(group, imageInfo) {
		return r.probeBuilderLivenessCoreOperator
	}

	return r.probeBuilderLivenessCore
}

func (r *Resources) probeBuilderStartupCoreSelect(group api.ServerGroup, imageInfo api.ImageInfo) probeBuilder {
	if features.JWTRotation().Enabled() || IsServerProgressAvailable(group, imageInfo) {
		return r.probeBuilderStartupCoreOperator
	}

	return r.probeBuilderStartupCore
}

func (r *Resources) probeBuilderLivenessCoreOperator(spec api.DeploymentSpec, group api.ServerGroup,
	image api.ImageInfo) (Probe, error) {
	args := r.probeCommand(spec, api.ProbeTypeLiveness)

	cmdProbeConfig := &probes.CMDProbeConfig{
		Command: args,
	}
	if IsServerProgressAvailable(group, image) {
		cmdProbeConfig.FailureThreshold = math.MaxInt32
	}

	return cmdProbeConfig, nil
}

func (r *Resources) probeBuilderStartupCoreOperator(spec api.DeploymentSpec, group api.ServerGroup,
	image api.ImageInfo) (Probe, error) {
	args := r.probeCommand(spec, api.ProbeTypeStartUp)

	retries, periodSeconds := getProbeRetries(group)
	if IsServerProgressAvailable(group, image) {
		retries = math.MaxInt32
	}

	return &probes.CMDProbeConfig{
		Command:             args,
		FailureThreshold:    retries,
		PeriodSeconds:       periodSeconds,
		InitialDelaySeconds: 1,
	}, nil
}

func (r *Resources) probeBuilderLivenessCore(spec api.DeploymentSpec, group api.ServerGroup, _ api.ImageInfo) (Probe, error) {
	authorization := ""
	if spec.IsAuthenticated() {
		secretData, err := r.getJWTSecret(spec)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		authorization, err = jwt.CreateArangodJwtAuthorizationHeaderAllowedPaths(secretData, "kube-arangodb", []string{"/_api/version"})
		if err != nil {
			return nil, errors.WithStack(err)
		}
	}
	return &probes.HTTPProbeConfig{
		LocalPath:     "/_api/version",
		Secure:        spec.IsSecure(),
		Authorization: authorization,
	}, nil
}

func (r *Resources) probeBuilderStartupCore(spec api.DeploymentSpec, group api.ServerGroup, _ api.ImageInfo) (Probe, error) {
	retries, periodSeconds := getProbeRetries(group)

	authorization := ""
	if spec.IsAuthenticated() {
		secretData, err := r.getJWTSecret(spec)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		authorization, err = jwt.CreateArangodJwtAuthorizationHeaderAllowedPaths(secretData, "kube-arangodb", []string{"/_api/version"})
		if err != nil {
			return nil, errors.WithStack(err)
		}
	}
	return &probes.HTTPProbeConfig{
		LocalPath:           "/_api/version",
		Secure:              spec.IsSecure(),
		Authorization:       authorization,
		FailureThreshold:    retries,
		PeriodSeconds:       periodSeconds,
		InitialDelaySeconds: 1,
	}, nil
}

func (r *Resources) probeBuilderReadinessSimpleCoreSelect() probeBuilder {
	if features.JWTRotation().Enabled() {
		return r.probeBuilderReadinessSimpleCoreOperator
	}

	return r.probeBuilderReadinessSimpleCore
}

func (r *Resources) probeBuilderReadinessSimpleCoreOperator(spec api.DeploymentSpec, group api.ServerGroup,
	image api.ImageInfo) (Probe, error) {
	p, err := r.probeBuilderReadinessCoreOperator(spec, group, image)
	if err != nil {
		return nil, err
	}

	if p == nil {
		return nil, nil
	}

	p.SetSpec(&api.ServerGroupProbeSpec{
		InitialDelaySeconds: util.NewType[int32](15),
		PeriodSeconds:       util.NewType[int32](10),
	})

	return p, nil
}

func (r *Resources) probeBuilderReadinessSimpleCore(spec api.DeploymentSpec, group api.ServerGroup,
	image api.ImageInfo) (Probe, error) {
	p, err := r.probeBuilderReadinessCore(spec, group, image)
	if err != nil {
		return nil, err
	}

	if p == nil {
		return nil, nil
	}

	p.SetSpec(&api.ServerGroupProbeSpec{
		InitialDelaySeconds: util.NewType[int32](15),
		PeriodSeconds:       util.NewType[int32](10),
	})

	return p, nil
}

func (r *Resources) probeBuilderReadinessCoreSelect() probeBuilder {
	if features.JWTRotation().Enabled() {
		return r.probeBuilderReadinessCoreOperator
	}

	return r.probeBuilderReadinessCore
}

func (r *Resources) probeBuilderReadinessCoreOperator(spec api.DeploymentSpec, group api.ServerGroup, _ api.ImageInfo) (Probe, error) {
	args := r.probeCommand(spec, api.ProbeTypeReadiness)

	return &probes.CMDProbeConfig{
		Command:             args,
		InitialDelaySeconds: 2,
		PeriodSeconds:       2,
	}, nil
}

func (r *Resources) probeBuilderReadinessCore(spec api.DeploymentSpec, _ api.ServerGroup, _ api.ImageInfo) (Probe, error) {
	// /_admin/server/availability is the way to go, it is available since 3.3.9
	localPath := "/_admin/server/availability"
	if features.FailoverLeadership().Enabled() && r.context.GetMode() == api.DeploymentModeActiveFailover {
		localPath = "/_api/version"
	}

	authorization := ""
	if spec.IsAuthenticated() {
		secretData, err := r.getJWTSecret(spec)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		authorization, err = jwt.CreateArangodJwtAuthorizationHeaderAllowedPaths(secretData, "kube-arangodb", []string{localPath})
		if err != nil {
			return nil, errors.WithStack(err)
		}
	}
	probeCfg := &probes.HTTPProbeConfig{
		LocalPath:           localPath,
		Secure:              spec.IsSecure(),
		Authorization:       authorization,
		InitialDelaySeconds: 2,
		PeriodSeconds:       2,
	}

	return probeCfg, nil
}

func (r *Resources) probeBuilderLivenessSync(spec api.DeploymentSpec, group api.ServerGroup, _ api.ImageInfo) (Probe, error) {
	authorization := ""

	if spec.Sync.Monitoring.GetTokenSecretName() != "" {
		// Use monitoring token
		token, err := r.getSyncMonitoringToken(spec)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		authorization = "bearer " + token
	} else if group == api.ServerGroupSyncMasters {
		// Fall back to JWT secret
		secretData, err := r.getSyncJWTSecret(spec)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		authorization, err = jwt.CreateArangodJwtAuthorizationHeaderAllowedPaths(secretData, "kube-arangodb", []string{"/_api/version"})
		if err != nil {
			return nil, errors.WithStack(err)
		}
	} else {
		// Don't have a probe
		return nil, nil
	}
	return &probes.HTTPProbeConfig{
		LocalPath:     "/_api/version",
		Secure:        spec.Sync.TLS.IsSecure(),
		Authorization: authorization,
	}, nil
}

func (r *Resources) probeBuilderStartupSync(spec api.DeploymentSpec, group api.ServerGroup, _ api.ImageInfo) (Probe, error) {
	authorization := ""

	if spec.Sync.Monitoring.GetTokenSecretName() != "" {
		// Use monitoring token
		token, err := r.getSyncMonitoringToken(spec)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		authorization = "bearer " + token
	} else if group == api.ServerGroupSyncMasters {
		// Fall back to JWT secret
		secretData, err := r.getSyncJWTSecret(spec)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		authorization, err = jwt.CreateArangodJwtAuthorizationHeaderAllowedPaths(secretData, "kube-arangodb", []string{"/_api/version"})
		if err != nil {
			return nil, errors.WithStack(err)
		}
	} else {
		// Don't have a probe
		return nil, nil
	}
	return &probes.HTTPProbeConfig{
		LocalPath:     "/_api/version",
		Secure:        spec.Sync.TLS.IsSecure(),
		Authorization: authorization,
	}, nil
}

// getProbeRetries returns how many attempts should be performed and what is the period in seconds between these attempts.
func getProbeRetries(group api.ServerGroup) (int32, int32) {
	// Set default values.
	period, howLong := 5*time.Second, 300*time.Second

	if group == api.ServerGroupDBServers {
		// Wait 6 hours (in seconds) for WAL replay.
		howLong = 6 * time.Hour
	} else if group == api.ServerGroupCoordinators {
		// Coordinator should wait for agents, but agents could take more time to spin up.
		howLong = time.Hour
	}

	return int32(howLong / period), int32(period / time.Second)
}

// IsServerProgressAvailable returns true if server progress is available.
func IsServerProgressAvailable(group api.ServerGroup, imageInfo api.ImageInfo) bool {
	return group == api.ServerGroupDBServers &&
		features.Version310().ImageSupported(&imageInfo)
}
