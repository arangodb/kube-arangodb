//
// DISCLAIMER
//
// Copyright 2016-2025 ArangoDB GmbH, Cologne, Germany
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
	"context"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"net"
	"path"
	"path/filepath"
	"strconv"
	"sync"

	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/deployment/acs/sutil"
	"github.com/arangodb/kube-arangodb/pkg/deployment/features"
	"github.com/arangodb/kube-arangodb/pkg/deployment/member"
	"github.com/arangodb/kube-arangodb/pkg/deployment/pod"
	"github.com/arangodb/kube-arangodb/pkg/deployment/reconciler"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/assertion"
	utilConstants "github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	inspectorInterface "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/generic"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/interfaces"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/kerrors"
	ktls "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/tls"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/tolerations"
)

// createArangodArgsWithUpgrade creates command line arguments for an arangod server upgrade in the given group.
func createArangodArgsWithUpgrade(cachedStatus interfaces.Inspector, input pod.Input) ([]string, error) {
	return createArangodArgs(cachedStatus, input, pod.AutoUpgrade().Args(input)...)
}

// createArangodArgs creates command line arguments for an arangod server in the given group.
func createArangodArgs(cachedStatus interfaces.Inspector, input pod.Input, additionalOptions ...k8sutil.OptionPair) ([]string, error) {
	options := k8sutil.CreateOptionPairs(64)

	scheme := "tcp"
	if input.Deployment.IsSecure() {
		scheme = "ssl"
	}

	if input.GroupSpec.GetExternalPortEnabled() {
		options.Addf("--server.endpoint", "%s://%s:%d", scheme, input.Deployment.GetListenAddr(), input.GroupSpec.GetPort())
	}

	if port := input.GroupSpec.InternalPort; port != nil {
		internalScheme := "tcp"
		if input.Deployment.IsSecure() && input.GroupSpec.InternalPortProtocol.Get() == api.ServerGroupPortProtocolHTTPS {
			internalScheme = "ssl"
		}
		options.Addf("--server.endpoint", "%s://127.0.0.1:%d", internalScheme, *port)
	}

	// Authentication
	options.Merge(pod.JWT().Args(input))

	// Security
	options.Merge(pod.Security().Args(input))

	// Storage engine
	options.Add("--server.storage-engine", input.Deployment.GetStorageEngine().AsArangoArgument())

	// Logging
	options.Add("--log.level", "INFO")

	options.Append(additionalOptions...)

	// TLS
	options.Merge(pod.TLS().Args(input))

	// RocksDB
	options.Merge(pod.Encryption().Args(input))

	options.Add("--database.directory", shared.ArangodVolumeMountDir)
	options.Add("--log.output", "+")

	options.Merge(pod.SNI().Args(input))

	endpoint, err := pod.GenerateMemberEndpoint(cachedStatus, input.ApiObject, input.Deployment, input.Group, input.Member)
	if err != nil {
		return nil, err
	}
	endpoint = util.TypeOrDefault[string](input.Member.Endpoint, endpoint)

	myTCPURL := scheme + "://" + net.JoinHostPort(endpoint, strconv.Itoa(shared.ArangoPort))
	addAgentEndpoints := false
	switch input.Group {
	case api.ServerGroupAgents:
		options.Add("--agency.disaster-recovery-id", input.Member.ID)
		options.Add("--agency.activate", "true")
		options.Add("--agency.my-address", myTCPURL)
		options.Addf("--agency.size", "%d", input.Deployment.Agents.GetCount())
		options.Add("--agency.supervision", "true")
		options.Add("--foxx.queues", false)
		options.Add("--server.statistics", "false")
		for _, p := range input.Status.Members.Agents {
			if p.ID != input.Member.ID {
				dnsName, err := pod.GenerateMemberEndpoint(cachedStatus, input.ApiObject, input.Deployment, api.ServerGroupAgents, p)
				if err != nil {
					return nil, err
				}
				options.Addf("--agency.endpoint", "%s://%s", scheme, net.JoinHostPort(util.TypeOrDefault[string](p.Endpoint, dnsName), strconv.Itoa(shared.ArangoPort)))
			}
		}
	case api.ServerGroupDBServers:
		addAgentEndpoints = true
		options.Add("--cluster.my-address", myTCPURL)
		options.Add("--cluster.my-role", "PRIMARY")
		options.Add("--foxx.queues", false)
		options.Add("--server.statistics", "true")
		if IsServerProgressAvailable(input.Group, input.Image) {
			options.Add("--server.early-connections", "true")
		}
	case api.ServerGroupCoordinators:
		addAgentEndpoints = true
		options.Add("--cluster.my-address", myTCPURL)
		options.Add("--cluster.my-role", "COORDINATOR")
		options.Add("--foxx.queues", input.Deployment.Features.GetFoxxQueues())
		options.Add("--server.statistics", "true")
		if input.Deployment.ExternalAccess.HasAdvertisedEndpoint() {
			options.Add("--cluster.my-advertised-endpoint", input.Deployment.ExternalAccess.GetAdvertisedEndpoint())
		}
	case api.ServerGroupSingle:
		options.Add("--foxx.queues", input.Deployment.Features.GetFoxxQueues())
		options.Add("--server.statistics", "true")
		if input.Deployment.GetMode() == api.DeploymentModeActiveFailover {
			addAgentEndpoints = true
			options.Add("--replication.automatic-failover", "true")
			options.Add("--cluster.my-address", myTCPURL)
			options.Add("--cluster.my-role", "SINGLE")
			if input.Deployment.ExternalAccess.HasAdvertisedEndpoint() {
				options.Add("--cluster.my-advertised-endpoint", input.Deployment.ExternalAccess.GetAdvertisedEndpoint())
			}
		}
	}
	if addAgentEndpoints {
		for _, p := range input.Status.Members.Agents {
			dnsName, err := pod.GenerateMemberEndpoint(cachedStatus, input.ApiObject, input.Deployment, api.ServerGroupAgents, p)
			if err != nil {
				return nil, err
			}
			options.Addf("--cluster.agency-endpoint", "%s://%s", scheme, net.JoinHostPort(util.TypeOrDefault[string](p.Endpoint, dnsName), strconv.Itoa(shared.ArangoPort)))
		}
	}

	if features.EncryptionRotation().Enabled() {
		options.Add("--rocksdb.encryption-key-rotation", "true")
	}

	args := options.Copy().Sort().AsArgs()

	if a := input.GroupSpec.Args; len(a) > 0 {
		args = append(args, a...)
	}
	if a := input.ArangoMember.Spec.Overrides.GetArgs(); len(a) > 0 {
		args = append(args, a...)
	}

	return args, nil
}

// createArangodNumactl creates command line arguments for a numactl in the given group.
func createArangodNumactl(spec api.ServerGroupSpec) []string {
	if !spec.Numactl.IsEnabled() {
		return nil
	}

	args := spec.Numactl.GetArgs()

	if len(args) == 0 {
		return []string{
			spec.Numactl.GetPath(),
		}
	}

	r := make([]string, len(args)+1)

	r[0] = spec.Numactl.GetPath()

	copy(r[1:], args)

	return r
}

// withArangodNumactl creates command line arguments with a numactl in a given group.
func withArangodNumactl(spec api.ServerGroupSpec, args ...string) []string {
	nmctl := createArangodNumactl(spec)
	if len(nmctl) == 0 {
		return args
	}

	vs := make([]string, len(args)+len(nmctl))

	copy(vs, nmctl)
	copy(vs[len(nmctl):], args)

	return vs
}

// createArangoSyncArgs creates command line arguments for an arangosync server in the given group.
func createArangoSyncArgs(input pod.Input, additionalOptions ...k8sutil.OptionPair) []string {
	options := k8sutil.CreateOptionPairs(64)
	var runCmd string
	port := input.GroupSpec.GetPort()

	if input.Deployment.Sync.Monitoring.GetTokenSecretName() != "" {
		options.Addf("--monitoring.token", "$(%s)", utilConstants.EnvArangoSyncMonitoringToken)
	}
	masterSecretPath := filepath.Join(shared.MasterJWTSecretVolumeMountDir, utilConstants.SecretKeyToken)
	options.Add("--master.jwt-secret", masterSecretPath)

	var masterEndpoint []string
	switch input.Group {
	case api.ServerGroupSyncMasters:
		runCmd = "master"
		masterEndpoint = input.Deployment.Sync.ExternalAccess.ResolveMasterEndpoint(k8sutil.CreateSyncMasterClientServiceDNSNameWithDomain(input.ApiObject, input.Deployment.ClusterDomain), int(port))
		keyPath := filepath.Join(shared.TLSKeyfileVolumeMountDir, utilConstants.SecretTLSKeyfile)
		clientCAPath := filepath.Join(shared.ClientAuthCAVolumeMountDir, utilConstants.SecretCACertificate)
		options.Add("--server.keyfile", keyPath)
		options.Add("--server.client-cafile", clientCAPath)
		options.Add("--mq.type", "direct")
		if input.Deployment.IsAuthenticated() {
			clusterSecretPath := filepath.Join(shared.ClusterJWTSecretVolumeMountDir, utilConstants.SecretKeyToken)
			options.Add("--cluster.jwt-secret", clusterSecretPath)
		}
		dbServiceName := k8sutil.CreateDatabaseClientServiceName(input.ApiObject.GetName())
		scheme := "http"
		if input.Deployment.IsSecure() {
			scheme = "https"
		}
		options.Addf("--cluster.endpoint", "%s://%s:%d", scheme, dbServiceName, shared.ArangoPort)
	case api.ServerGroupSyncWorkers:
		runCmd = "worker"
		masterEndpointHost := k8sutil.CreateSyncMasterClientServiceName(input.ApiObject.GetName())
		masterEndpoint = []string{"https://" + net.JoinHostPort(masterEndpointHost, strconv.Itoa(shared.ArangoSyncMasterPort))}
	}
	for _, ep := range masterEndpoint {
		options.Add("--master.endpoint", ep)
	}
	serverEndpoint := "https://" + net.JoinHostPort(k8sutil.CreatePodDNSNameWithDomain(input.ApiObject, input.Deployment.ClusterDomain, input.Group.AsRole(), input.Member.ID), strconv.Itoa(int(port)))
	options.Add("--server.endpoint", serverEndpoint)
	options.Add("--server.port", strconv.Itoa(int(port)))

	options.Append(additionalOptions...)

	args := []string{
		"run",
		runCmd,
	}

	args = append(args, options.Copy().Sort().AsArgs()...)

	if a := input.GroupSpec.Args; len(a) > 0 {
		args = append(args, a...)
	}
	if a := input.ArangoMember.Spec.Overrides.GetArgs(); len(a) > 0 {
		args = append(args, a...)
	}

	return args
}

func createArangoGatewayArgs(input pod.Input, additionalOptions ...k8sutil.OptionPair) []string {
	options := k8sutil.CreateOptionPairs(64)
	if input.Deployment.Gateway.IsDynamic() {
		options.Add("--config-path", path.Join(utilConstants.MemberConfigVolumeMountDir, utilConstants.GatewayConfigFileName))
	} else {
		options.Add("--config-path", path.Join(utilConstants.GatewayVolumeMountDir, utilConstants.GatewayConfigFileName))
	}

	options.Append(additionalOptions...)

	args := options.Sort().AsSplittedArgs()

	if a := input.GroupSpec.Args; len(a) > 0 {
		args = append(args, a...)
	}
	if a := input.ArangoMember.Spec.Overrides.GetArgs(); len(a) > 0 {
		args = append(args, a...)
	}

	return args
}

// CreatePodTolerations creates a list of tolerations for a pod created for the given group.
func (r *Resources) CreatePodTolerations(group api.ServerGroup, groupSpec api.ServerGroupSpec) []core.Toleration {
	return tolerations.MergeTolerationsIfNotFound(CreatePodTolerations(r.context.GetMode(), group), groupSpec.GetTolerations())
}

func (r *Resources) RenderPodTemplateForMember(ctx context.Context, acs sutil.ACS, spec api.DeploymentSpec, status api.DeploymentStatus, memberID string, imageInfo api.ImageInfo) (*core.PodTemplateSpec, error) {
	if p, err := r.RenderPodForMember(ctx, acs, spec, status, memberID, imageInfo); err != nil {
		return nil, err
	} else {
		return &core.PodTemplateSpec{
			ObjectMeta: p.ObjectMeta,
			Spec:       p.Spec,
		}, nil
	}
}

func (r *Resources) RenderPodForMember(ctx context.Context, acs sutil.ACS, spec api.DeploymentSpec, status api.DeploymentStatus, memberID string, imageInfo api.ImageInfo) (*core.Pod, error) {
	log := r.log.Str("section", "member")
	apiObject := r.context.GetAPIObject()
	m, group, found := status.Members.ElementByID(memberID)
	if !found {
		return nil, errors.WithStack(errors.Errorf("Member '%s' not found", memberID))
	}
	groupSpec := spec.GetServerGroupSpec(group)

	if spec.Mode.Get() == api.DeploymentModeActiveFailover && !features.ActiveFailover().ImageSupported(&imageInfo) {
		return nil, errors.Errorf("ActiveFailover starting from ArangoDB 3.12 is not supported")
	}

	memberName := m.ArangoMemberName(r.context.GetAPIObject().GetName(), group)

	member, ok := acs.CurrentClusterCache().ArangoMember().V1().GetSimple(memberName)
	if !ok {
		return nil, errors.Errorf("ArangoMember %s not found", memberName)
	}

	cluster, ok := acs.Cluster(m.ClusterID)
	if !ok {
		return nil, errors.Errorf("Cluster is not found")
	}

	if !cluster.Ready() {
		return nil, errors.Errorf("Cluster is not ready")
	}

	cache := cluster.Cache()

	// Update pod name
	role := group.AsRole()
	roleAbbr := group.AsRoleAbbreviated()

	podName := k8sutil.CreatePodName(apiObject.GetName(), roleAbbr, m.ID, CreatePodSuffix(spec))

	input := pod.Input{
		ApiObject:    r.context.GetAPIObject(),
		Deployment:   spec,
		Status:       status,
		GroupSpec:    groupSpec,
		Group:        group,
		Image:        imageInfo,
		Member:       m,
		ArangoMember: *member,
		AutoUpgrade:  m.Conditions.IsTrue(api.ConditionTypeAutoUpgrade) || spec.Upgrade.Get().AutoUpgrade,
	}

	var podCreator interfaces.PodCreator
	switch group.Type() {
	case api.ServerGroupTypeArangoD:
		// Prepare arguments
		podCreator = &MemberArangoDPod{
			Input:        input,
			podName:      podName,
			resources:    r,
			context:      r.context,
			cachedStatus: cache,
		}
	case api.ServerGroupTypeArangoSync:
		// Check image
		if !imageInfo.Enterprise {
			log.Str("image", spec.GetImage()).Debug("Image is not an enterprise image")
			return nil, errors.WithStack(errors.Errorf("Image '%s' does not contain an Enterprise version of ArangoDB", spec.GetImage()))
		}
		// Check if the sync image is overwritten by the SyncSpec
		if spec.Sync.HasSyncImage() {
			input.Image.Image = spec.Sync.GetSyncImage()
		}

		podCreator = &MemberSyncPod{
			Input: input,

			podName:      podName,
			resources:    r,
			cachedStatus: cache,
		}
	case api.ServerGroupTypeGateway:
		input.Image.Image = r.context.GetOperatorImage()
		if spec.Gateway.GetImage() != "" {
			input.Image.Image = spec.Gateway.GetImage()
		}

		podCreator = &MemberGatewayPod{
			Input:        input,
			podName:      podName,
			resources:    r,
			context:      r.context,
			cachedStatus: cache,
		}
	default:
		return nil, assertion.InvalidGroupKey.Assert(true, "Unable to render pod for an unknown group: %s", group.AsRole())
	}

	pod, err := RenderArangoPod(ctx, cache, apiObject, role, m.ID, podName, podCreator)
	if err != nil {
		return nil, err
	}

	if features.RandomPodNames().Enabled() {
		// The server will generate the name with some additional suffix after `-`.
		pod.GenerateName = pod.Name + "-"
		pod.Name = ""
	}

	return pod, nil
}

func (r *Resources) SelectImage(spec api.DeploymentSpec, status api.DeploymentStatus) (api.ImageInfo, bool) {
	var imageInfo api.ImageInfo
	if current := status.CurrentImage; current != nil {
		// Use current image
		imageInfo = *current
	} else {
		// Find image ID
		info, imageFound := status.Images.GetByImage(spec.GetImage())
		if !imageFound {
			return api.ImageInfo{}, false
		}
		imageInfo = info
		// Save image as current image
		status.CurrentImage = &info
	}
	return imageInfo, true
}

func (r *Resources) SelectImageForMember(spec api.DeploymentSpec, status api.DeploymentStatus, member api.MemberStatus) (api.ImageInfo, bool) {
	if member.Image != nil {
		return *member.Image, true
	}

	return r.SelectImage(spec, status)
}

// createPodForMember creates all Pods listed in member status
func (r *Resources) createPodForMember(ctx context.Context, cachedStatus inspectorInterface.Inspector, spec api.DeploymentSpec, arangoMember *api.ArangoMember, memberID string, imageNotFoundOnce *sync.Once) error {
	log := r.log.Str("section", "member")
	status := r.context.GetStatus()

	// Select image
	imageInfo, imageFound := r.SelectImage(spec, status)
	if !imageFound {
		imageNotFoundOnce.Do(func() {
			log.Str("image", spec.GetImage()).Debug("Image ID is not known yet for image")
		})
		return nil
	}

	template := arangoMember.Status.Template

	if template == nil {
		// Template not yet propagated
		return errors.Errorf("Template not yet propagated")
	}

	if status.CurrentImage == nil {
		status.CurrentImage = &imageInfo
	}

	m, group, found := status.Members.ElementByID(memberID)
	if m.Image == nil {
		m.Image = status.CurrentImage

		if err := status.Members.Update(m, group); err != nil {
			return errors.WithStack(err)
		}
	}

	imageInfo = *m.Image

	apiObject := r.context.GetAPIObject()

	if !found {
		return errors.WithStack(errors.Errorf("Member '%s' not found", memberID))
	}

	// Update pod name
	role := group.AsRole()

	newPhase := api.MemberPhaseCreated
	// Create pod
	switch group.Type() {
	case api.ServerGroupTypeArangoD:
		// Prepare arguments
		autoUpgrade := m.Conditions.IsTrue(api.ConditionTypeAutoUpgrade)
		if autoUpgrade {
			newPhase = api.MemberPhaseUpgrading
		}

		ctxChild, cancel := globals.GetGlobalTimeouts().Kubernetes().WithTimeout(ctx)
		defer cancel()

		podName, uid, err := CreateArangoPod(ctxChild, cachedStatus.PodsModInterface().V1(), apiObject, spec, group, CreatePodFromTemplate(template.PodSpec))
		if err != nil {
			if uerr := r.context.WithMemberStatusUpdateErr(ctx, m.ID, group, updateMemberPhase(api.MemberPhaseCreationFailed)); uerr != nil {
				return errors.WithStack(uerr)
			}
			return errors.WithStack(err)
		}

		var pod api.MemberPodStatus

		pod.Name = podName
		pod.UID = uid
		pod.SpecVersion = template.PodSpecChecksum

		m.Pod = &pod
		m.Pod.Propagate(&m)

		m.ArangoVersion = m.Image.ArangoDBVersion
		m.ImageID = m.Image.ImageID

		// reset old sidecar values to nil
		m.SideCarSpecs = nil

		log.Str("pod-name", pod.Name).Debug("Created pod")
		if m.Image == nil {
			log.Str("pod-name", pod.Name).Debug("Created pod with default image")
		} else {
			log.Str("pod-name", pod.Name).Debug("Created pod with predefined image")
		}
	case api.ServerGroupTypeArangoSync:
		// Check monitoring token secret
		if group == api.ServerGroupSyncMasters {
			// Create TLS secret
			tlsKeyfileSecretName := k8sutil.CreateTLSKeyfileSecretName(apiObject.GetName(), role, m.ID)

			names, err := ktls.GetSyncAltNames(apiObject, spec, spec.Sync.TLS, group, m)
			if err != nil {
				return errors.WithStack(errors.Wrapf(err, "Failed to render alt names"))
			}

			owner := arangoMember.AsOwner()
			_, err = createTLSServerCertificate(ctx, log, cachedStatus, cachedStatus.SecretsModInterface().V1(), names, spec.Sync.TLS, tlsKeyfileSecretName, &owner)
			if err != nil && !kerrors.IsAlreadyExists(err) {
				return errors.WithStack(errors.Wrapf(err, "Failed to create TLS keyfile secret"))
			}
		}

		ctxChild, cancel := globals.GetGlobalTimeouts().Kubernetes().WithTimeout(ctx)
		defer cancel()
		podName, uid, err := CreateArangoPod(ctxChild, cachedStatus.PodsModInterface().V1(), apiObject, spec, group, CreatePodFromTemplate(template.PodSpec))
		if err != nil {
			if uerr := r.context.WithMemberStatusUpdateErr(ctx, m.ID, group, updateMemberPhase(api.MemberPhaseCreationFailed)); uerr != nil {
				return errors.WithStack(uerr)
			}
			return errors.WithStack(err)
		}

		var pod api.MemberPodStatus

		pod.Name = podName
		pod.UID = uid
		pod.SpecVersion = template.PodSpecChecksum

		m.Pod = &pod
		m.Pod.Propagate(&m)

		log.Str("pod-name", pod.Name).Debug("Created pod")
	case api.ServerGroupTypeGateway:
		ctxChild, cancel := globals.GetGlobalTimeouts().Kubernetes().WithTimeout(ctx)
		defer cancel()
		podName, uid, err := CreateArangoPod(ctxChild, cachedStatus.PodsModInterface().V1(), apiObject, spec, group, CreatePodFromTemplate(template.PodSpec))
		if err != nil {
			if uErr := r.context.WithMemberStatusUpdateErr(ctx, m.ID, group, updateMemberPhase(api.MemberPhaseCreationFailed)); uErr != nil {
				return errors.WithStack(uErr)
			}
			return errors.WithStack(err)
		}

		var pod api.MemberPodStatus

		pod.Name = podName
		pod.UID = uid
		pod.SpecVersion = template.PodSpecChecksum

		m.Pod = &pod
		m.Pod.Propagate(&m)

		log.Str("pod-name", pod.Name).Debug("Created Gateway pod")
	default:
		return assertion.InvalidGroupKey.Assert(true, "Unable to create pod for an unknown group: %s", group.AsRole())
	}

	member.GetPhaseExecutor().Execute(r.context.GetAPIObject(), spec, group, &m, api.Action{}, newPhase)

	if top := status.Topology; top.Enabled() {
		if m.Topology != nil && m.Topology.ID == top.ID {
			if top.IsTopologyEvenlyDistributed(group) {
				m.Conditions.Update(api.ConditionTypeTopologyAware, true, "Topology Aware", "Topology Aware")
			} else {
				m.Conditions.Update(api.ConditionTypeTopologyAware, false, "Topology Aware", "Topology invalid")
			}
		} else {
			m.Conditions.Update(api.ConditionTypeTopologyAware, false, "Topology spec missing", "Topology spec missing")
		}
	}

	log.Str("pod", m.Pod.GetName()).Info("Updating member")
	if err := status.Members.Update(m, group); err != nil {
		return errors.WithStack(err)
	}
	if err := r.context.UpdateStatus(ctx, status); err != nil {
		return errors.WithStack(err)
	}
	// Create event
	r.context.CreateEvent(k8sutil.NewPodCreatedEvent(m.Pod.GetName(), role, apiObject))
	cachedStatus.GetThrottles().Pod().Invalidate()

	return nil
}

// RenderArangoPod renders new ArangoD Pod
func RenderArangoPod(ctx context.Context, cachedStatus inspectorInterface.Inspector, deployment k8sutil.APIObject,
	role, id, podName string, podCreator interfaces.PodCreator) (*core.Pod, error) {

	// Validate if the pod can be created.
	if err := podCreator.Validate(cachedStatus); err != nil {
		return nil, errors.Wrapf(err, "Validation of pods resources failed")
	}

	// Prepare basic pod.
	p := k8sutil.NewPod(deployment.GetName(), role, id, podName, podCreator)

	for k, v := range podCreator.Annotations() {
		if p.Annotations == nil {
			p.Annotations = map[string]string{}
		}

		p.Annotations[k] = v
	}

	for k, v := range podCreator.Labels() {
		if p.Labels == nil {
			p.Labels = map[string]string{}
		}

		p.Labels[k] = v
	}

	if err := podCreator.Init(ctx, cachedStatus, &p); err != nil {
		return nil, err
	}

	if initContainers, err := podCreator.GetInitContainers(cachedStatus); err != nil {
		return nil, errors.WithStack(err)
	} else if initContainers != nil {
		p.Spec.InitContainers = append(p.Spec.InitContainers, initContainers...)
	}

	p.Spec.Volumes = podCreator.GetVolumes()
	c, err := k8sutil.NewContainer(podCreator.GetContainerCreator())
	if err != nil {
		return nil, errors.WithStack(err)
	}

	p.Spec.Containers = append(p.Spec.Containers, c)
	if err := podCreator.GetSidecars(&p); err != nil {
		return nil, err
	}

	if err := podCreator.ApplyPodSpec(&p.Spec); err != nil {
		return nil, err
	}

	// Add affinity
	p.Spec.Affinity = &core.Affinity{
		NodeAffinity:    podCreator.GetNodeAffinity(),
		PodAntiAffinity: podCreator.GetPodAntiAffinity(),
		PodAffinity:     podCreator.GetPodAffinity(),
	}

	if profiles, err := podCreator.Profiles(); err != nil {
		return nil, err
	} else if len(profiles) > 0 {
		if err := profiles.RenderOnTemplate(&p); err != nil {
			return nil, err
		}
	}

	return &core.Pod{
		ObjectMeta: p.ObjectMeta,
		Spec:       p.Spec,
	}, nil
}

// CreateArangoPod creates a new Pod with container provided by parameter 'containerCreator'
// If the pod already exists, nil is returned.
// If another error occurs, that error is returned.
func CreateArangoPod(ctx context.Context, c generic.ModClient[*core.Pod], deployment k8sutil.APIObject,
	deploymentSpec api.DeploymentSpec, group api.ServerGroup, pod *core.Pod) (string, types.UID, error) {
	podName, uid, err := k8sutil.CreatePod(ctx, c, pod, deployment.GetNamespace(), deployment.AsOwner())
	if err != nil {
		return "", "", errors.WithStack(err)
	}

	return podName, uid, nil
}

func CreatePodFromTemplate(p *core.PodTemplateSpec) *core.Pod {
	return &core.Pod{
		ObjectMeta: p.ObjectMeta,
		Spec:       p.Spec,
	}
}

func ChecksumArangoPod(groupSpec api.ServerGroupSpec, pod *core.Pod) (string, error) {
	shaPod := pod.DeepCopy()
	switch groupSpec.InitContainers.GetMode().Get() {
	case api.ServerGroupInitContainerUpdateMode:
		shaPod.Spec.InitContainers = groupSpec.InitContainers.GetContainers()
	default:
		shaPod.Spec.InitContainers = nil
	}

	data, err := json.Marshal(shaPod.Spec)
	if err != nil {
		return "", err
	}

	return util.SHA256(data), nil
}

// EnsurePods creates all Pods listed in member status
func (r *Resources) EnsurePods(ctx context.Context, cachedStatus inspectorInterface.Inspector) error {
	iterator := r.context.GetServerGroupIterator()
	deploymentStatus := r.context.GetStatus()
	imageNotFoundOnce := &sync.Once{}

	log := r.log.Str("section", "member")

	if err := iterator.ForeachServerGroupAccepted(func(group api.ServerGroup, groupSpec api.ServerGroupSpec, status *api.MemberStatusList) error {
		for _, m := range *status {
			if m.Phase != api.MemberPhasePending {
				continue
			}

			member, ok := cachedStatus.ArangoMember().V1().GetSimple(m.ArangoMemberName(r.context.GetName(), group))
			if !ok {
				// ArangoMember not found, skip
				continue
			}

			if member.Status.Template == nil {
				log.Warn("Missing Template")
				// Template is missing, nothing to do
				continue
			}

			log.Warn("Ensuring pod")

			if err := cachedStatus.Refresh(ctx); err != nil {
				return err
			}

			spec := r.context.GetSpec()
			if err := r.createPodForMember(ctx, cachedStatus, spec, member, m.ID, imageNotFoundOnce); err != nil {
				log.Err(err).Warn("Ensuring pod failed")
				return errors.WithStack(err)
			}

			if err := cachedStatus.Refresh(ctx); err != nil {
				return err
			}
		}
		return nil
	}, &deploymentStatus); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// CreatePodSuffix creates additional string to glue it to the POD name.
// The suffix is calculated according to the given spec, so it is easily to recognize by name if the pods have the same spec.
// The additional `postSuffix` can be provided. It can be used to distinguish restarts of POD.
func CreatePodSuffix(spec api.DeploymentSpec) string {
	if features.ShortPodNames().Enabled() || features.RandomPodNames().Enabled() {
		return ""
	}

	raw, _ := json.Marshal(spec)
	hash := sha1.Sum(raw)
	return fmt.Sprintf("%0x", hash)[:6]
}

func updateMemberPhase(phase api.MemberPhase) reconciler.DeploymentMemberStatusUpdateErrFunc {
	return func(s *api.MemberStatus) (bool, error) {
		if s == nil {
			return false, nil
		}

		if s.Phase == phase {
			return false, nil
		}

		s.Phase = phase

		return true, nil
	}
}
