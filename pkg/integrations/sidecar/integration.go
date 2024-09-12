//
// DISCLAIMER
//
// Copyright 2023-2024 ArangoDB GmbH, Cologne, Germany

package sidecar

import (
	"fmt"
	"strings"

	core "k8s.io/api/core/v1"

	schedulerApi "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1beta1"
	schedulerContainerApi "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1beta1/container"
	schedulerContainerResourcesApi "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1beta1/container/resources"
	schedulerIntegrationApi "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1beta1/integration"
	schedulerPodApi "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1beta1/pod"
	schedulerPodResourcesApi "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1beta1/pod/resources"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

const (
	ContainerName        = "integration"
	ListenPortName       = "integration"
	ListenPortHealthName = "health"
)

type Integration interface {
	Name() []string
	Envs() ([]core.EnvVar, error)
	GlobalEnvs() ([]core.EnvVar, error)
	Volumes() ([]core.Volume, []core.VolumeMount, error)
	Validate() error
}

func NewShutdownAnnotations(coreContainers []string) *schedulerApi.ProfileTemplate {
	pt := schedulerApi.ProfileTemplate{
		Pod: &schedulerPodApi.Pod{
			Metadata: &schedulerPodResourcesApi.Metadata{
				Annotations: map[string]string{},
			},
		},
	}

	for _, container := range coreContainers {
		pt.Pod.Metadata.Annotations[fmt.Sprintf("%s/%s", constants.AnnotationShutdownCoreContainer, container)] = constants.AnnotationShutdownCoreContainerModeWait
	}

	return &pt
}

func NewIntegrationEnablement(integrations ...Integration) (*schedulerApi.ProfileTemplate, error) {
	var envs, gEnvs []core.EnvVar
	var volumes []core.Volume
	var volumeMounts []core.VolumeMount

	for _, integration := range integrations {
		name := strings.Join(integration.Name(), "/")

		if err := integration.Validate(); err != nil {
			return nil, errors.Wrapf(err, "Failure in %s", name)
		}

		if lvolumes, lvolumeMounts, err := integration.Volumes(); err != nil {
			return nil, errors.Wrapf(err, "Failure in volumes %s", name)
		} else if len(lvolumes) > 0 || len(lvolumeMounts) > 0 {
			volumes = append(volumes, lvolumes...)
			volumeMounts = append(volumeMounts, lvolumeMounts...)
		}

		if lenvs, err := integration.Envs(); err != nil {
			return nil, errors.Wrapf(err, "Failure in envs %s", name)
		} else if len(lenvs) > 0 {
			envs = append(envs, lenvs...)
		}

		if lgenvs, err := integration.GlobalEnvs(); err != nil {
			return nil, errors.Wrapf(err, "Failure in global envs %s", name)
		} else if len(lgenvs) > 0 {
			gEnvs = append(gEnvs, lgenvs...)
		}
	}

	if len(envs) == 0 && len(gEnvs) == 0 {
		return nil, nil
	}

	return &schedulerApi.ProfileTemplate{
		Pod: &schedulerPodApi.Pod{
			Volumes: &schedulerPodResourcesApi.Volumes{
				Volumes: volumes,
			},
		},
		Container: &schedulerApi.ProfileContainerTemplate{
			Containers: map[string]schedulerContainerApi.Container{
				ContainerName: {
					Environments: &schedulerContainerResourcesApi.Environments{
						Env: envs,
					},
					VolumeMounts: &schedulerContainerResourcesApi.VolumeMounts{
						VolumeMounts: volumeMounts,
					},
				},
			},
			All: &schedulerContainerApi.Generic{
				Environments: &schedulerContainerResourcesApi.Environments{
					Env: gEnvs,
				},
			},
		},
	}, nil
}

func NewIntegration(image *schedulerContainerResourcesApi.Image, integration *schedulerIntegrationApi.Sidecar) (*schedulerApi.ProfileTemplate, error) {
	// Arguments

	exePath := k8sutil.BinaryPath()
	lifecycle, err := k8sutil.NewLifecycleFinalizersWithBinary(exePath)
	if err != nil {
		return nil, errors.Wrapf(err, "NewLifecycleFinalizers failed")
	}

	options := k8sutil.CreateOptionPairs(64)

	options.Addf("--services.address", "127.0.0.1:%d", integration.GetListenPort())
	options.Addf("--health.address", "0.0.0.0:%d", integration.GetControllerListenPort())

	// Envs

	var envs = []core.EnvVar{
		{
			Name:  "INTEGRATION_API_ADDRESS",
			Value: fmt.Sprintf("127.0.0.1:%d", integration.GetListenPort()),
		},
		{
			Name:  "INTEGRATION_SERVICE_ADDRESS",
			Value: fmt.Sprintf("127.0.0.1:%d", integration.GetListenPort()),
		},
	}

	c := schedulerContainerApi.Container{
		Core: &schedulerContainerResourcesApi.Core{
			Command: append([]string{exePath, "integration"}, options.Sort().AsArgs()...),
		},
		Environments: &schedulerContainerResourcesApi.Environments{
			Env: k8sutil.GetLifecycleEnv(),
		},
		Networking: &schedulerContainerResourcesApi.Networking{
			Ports: []core.ContainerPort{
				{
					Name:          ListenPortName,
					ContainerPort: int32(integration.GetListenPort()),
					Protocol:      core.ProtocolTCP,
				},
				{
					Name:          ListenPortHealthName,
					ContainerPort: int32(integration.GetControllerListenPort()),
					Protocol:      core.ProtocolTCP,
				},
			},
		},
		Image: image,

		Lifecycle: &schedulerContainerResourcesApi.Lifecycle{
			Lifecycle: lifecycle,
		},

		Probes: &schedulerContainerResourcesApi.Probes{
			ReadinessProbe: &core.Probe{
				ProbeHandler: core.ProbeHandler{
					GRPC: &core.GRPCAction{
						Port: int32(integration.GetControllerListenPort()),
					},
				},
				InitialDelaySeconds: 1,  // Wait 1s before first probe
				TimeoutSeconds:      2,  // Timeout of each probe is 2s
				PeriodSeconds:       30, // Interval between probes is 30s
				SuccessThreshold:    1,  // Single probe is enough to indicate success
				FailureThreshold:    2,  // Need 2 failed probes to consider a failed state
			},
		},
	}

	pt := schedulerApi.ProfileTemplate{
		Container: &schedulerApi.ProfileContainerTemplate{
			All: &schedulerContainerApi.Generic{
				Environments: &schedulerContainerResourcesApi.Environments{
					Env: envs,
				},
			},
			Containers: map[string]schedulerContainerApi.Container{
				ContainerName: util.TypeOrDefault(k8sutil.CreateDefaultContainerTemplate(image).With(&c).With(integration.GetContainer())),
			},
		},
		Pod: &schedulerPodApi.Pod{
			Metadata: &schedulerPodResourcesApi.Metadata{
				Annotations: map[string]string{},
			},
		},
	}

	pt.Pod.Metadata.Annotations[fmt.Sprintf("%s/%s", constants.AnnotationShutdownContainer, ContainerName)] = ListenPortHealthName
	pt.Pod.Metadata.Annotations[constants.AnnotationShutdownManagedContainer] = "true"

	pt.Container.All.Environments = &schedulerContainerResourcesApi.Environments{
		Env: envs,
	}

	return &pt, nil
}
