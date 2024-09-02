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

func WithIntegrationEnvs(in Integration) ([]core.EnvVar, error) {
	if v, ok := in.(IntegrationEnvs); ok {
		return v.Envs()
	}

	return nil, nil
}

type IntegrationEnvs interface {
	Integration
	Envs() ([]core.EnvVar, error)
}

func WithIntegrationVolumes(in Integration) ([]core.Volume, []core.VolumeMount, error) {
	if v, ok := in.(IntegrationVolumes); ok {
		return v.Volumes()
	}

	return nil, nil, nil
}

type IntegrationVolumes interface {
	Integration
	Volumes() ([]core.Volume, []core.VolumeMount, error)
}

type Integration interface {
	Name() []string
	Args() (k8sutil.OptionPairs, error)
	Validate() error
}

func NewIntegration(image *schedulerContainerResourcesApi.Image, integration *schedulerApi.IntegrationSidecar, coreContainers []string, integrations ...Integration) (*schedulerApi.ProfileTemplate, error) {
	for _, integration := range integrations {
		if err := integration.Validate(); err != nil {
			name := strings.Join(integration.Name(), "/")

			return nil, errors.Wrapf(err, "Failure in %s", name)
		}
	}

	// Arguments

	exePath := k8sutil.BinaryPath()
	lifecycle, err := k8sutil.NewLifecycleFinalizersWithBinary(exePath)
	if err != nil {
		return nil, errors.Wrapf(err, "NewLifecycleFinalizers failed")
	}

	options := k8sutil.CreateOptionPairs(64)

	options.Addf("--services.address", "127.0.0.1:%d", integration.GetListenPort())
	options.Addf("--health.address", "0.0.0.0:%d", integration.GetControllerListenPort())

	// Volumes
	var volumes []core.Volume
	var volumeMounts []core.VolumeMount

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

	for _, i := range integrations {
		name := strings.Join(i.Name(), "/")

		if err := i.Validate(); err != nil {
			return nil, errors.Wrapf(err, "Failure in %s", name)
		}

		if args, err := i.Args(); err != nil {
			return nil, errors.Wrapf(err, "Failure in arguments %s", name)
		} else if len(args) > 0 {
			options.Merge(args)
		}

		if lvolumes, lvolumeMounts, err := WithIntegrationVolumes(i); err != nil {
			return nil, errors.Wrapf(err, "Failure in volumes %s", name)
		} else if len(lvolumes) > 0 || len(lvolumeMounts) > 0 {
			volumes = append(volumes, lvolumes...)
			volumeMounts = append(volumeMounts, lvolumeMounts...)
		}

		if lenvs, err := WithIntegrationEnvs(i); err != nil {
			return nil, errors.Wrapf(err, "Failure in envs %s", name)
		} else if len(lenvs) > 0 {
			envs = append(envs, lenvs...)
		}

		envs = append(envs, core.EnvVar{
			Name: fmt.Sprintf("INTEGRATION_SERVICE_%s", strings.Join(util.FormatList(i.Name(), func(a string) string {
				return strings.ToUpper(a)
			}), "_")),
			Value: fmt.Sprintf("127.0.0.1:%d", integration.GetListenPort()),
		})
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

		VolumeMounts: &schedulerContainerResourcesApi.VolumeMounts{
			VolumeMounts: volumeMounts,
		},
	}

	pt := schedulerApi.ProfileTemplate{
		Container: &schedulerApi.ProfileContainerTemplate{
			Containers: map[string]schedulerContainerApi.Container{
				ContainerName: util.TypeOrDefault(k8sutil.CreateDefaultContainerTemplate(image).With(&c).With(integration.GetContainer())),
			},
		},
		Pod: &schedulerPodApi.Pod{
			Metadata: &schedulerPodResourcesApi.Metadata{
				Annotations: map[string]string{},
			},
			Volumes: &schedulerPodResourcesApi.Volumes{
				Volumes: volumes,
			},
		},
	}

	for _, container := range coreContainers {
		pt.Pod.Metadata.Annotations[fmt.Sprintf("%s/%s", constants.AnnotationShutdownCoreContainer, container)] = constants.AnnotationShutdownCoreContainerModeWait
	}

	pt.Pod.Metadata.Annotations[fmt.Sprintf("%s/%s", constants.AnnotationShutdownContainer, ContainerName)] = ListenPortHealthName
	pt.Pod.Metadata.Annotations[constants.AnnotationShutdownManagedContainer] = "true"

	pt.Container.Containers.ExtendContainers(&schedulerContainerApi.Container{
		Environments: &schedulerContainerResourcesApi.Environments{
			Env: envs,
		},
	}, coreContainers...)

	return &pt, nil
}
