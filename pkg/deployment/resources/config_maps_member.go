//
// DISCLAIMER
//
// Copyright 2024 ArangoDB GmbH, Cologne, Germany
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

	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/assertion"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	inspectorInterface "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector"
	configMapsV1 "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/configmap/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/kerrors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/patcher"
)

const (
	ConfigMapChecksumKey = "CHECKSUM"

	MemberConfigVolumeMountDir = "/etc/member/"
	MemberConfigVolumeName     = "member-config"

	MemberConfigChecksumENV = "MEMBER_CONFIG_CHECKSUM"
)

type memberConfigMapRenderer func(ctx context.Context, cachedStatus inspectorInterface.Inspector, member api.DeploymentStatusMemberElement) (map[string]string, error)

func (r *Resources) ensureMemberConfig(ctx context.Context, cachedStatus inspectorInterface.Inspector, configMaps configMapsV1.ModInterface) error {
	status := r.context.GetStatus()

	log := r.log.Str("section", "member-config-render")

	reconcileRequired := k8sutil.NewReconcile(cachedStatus)

	members := status.Members.AsList()

	if err := reconcileRequired.ParallelAll(len(members), func(id int) error {
		memberName := members[id].Member.ArangoMemberName(r.context.GetAPIObject().GetName(), members[id].Group)

		am, ok := cachedStatus.ArangoMember().V1().GetSimple(memberName)
		if !ok {
			return errors.Errorf("ArangoMember %s not found", memberName)
		}

		switch members[id].Group.Type() {
		case api.ServerGroupTypeGateway, api.ServerGroupTypeArangoSync, api.ServerGroupTypeArangoD:
			elements, err := r.renderMemberConfigElements(ctx, cachedStatus, members[id], r.ensureMemberConfigGatewayConfig)
			if err != nil {
				return err
			}

			if len(elements) == 0 {
				// CM should be gone
				if obj, ok := cachedStatus.ConfigMap().V1().GetSimple(memberName); !ok {
					return nil
				} else {
					if err := cachedStatus.ConfigMapsModInterface().V1().Delete(ctx, memberName, meta.DeleteOptions{
						Preconditions: meta.NewUIDPreconditions(string(obj.GetUID())),
					}); err != nil {
						if !kerrors.IsNotFound(err) {
							return err
						}
					}
				}
			} else {
				// We expect CM
				if obj, ok := cachedStatus.ConfigMap().V1().GetSimple(memberName); !ok {
					// Let's Create ConfigMap
					obj = &core.ConfigMap{
						ObjectMeta: meta.ObjectMeta{
							Name: memberName,
						},
						Data: elements,
					}

					err = globals.GetGlobalTimeouts().Kubernetes().RunWithTimeout(ctx, func(ctxChild context.Context) error {
						return k8sutil.CreateConfigMap(ctxChild, configMaps, obj, util.NewType(am.AsOwner()))
					})
					if kerrors.IsAlreadyExists(err) {
						// CM added while we tried it also
						return nil
					} else if err != nil {
						// Failed to create
						return errors.WithStack(err)
					}

					return errors.Reconcile()
				} else {
					// CM Exists, checks checksum - if key is not in the map we return empty string
					if currentSha, expectedSha := util.Optional(obj.Data, ConfigMapChecksumKey, ""), util.Optional(elements, ConfigMapChecksumKey, ""); currentSha != expectedSha || currentSha == "" {
						// We need to do the update
						if _, changed, err := patcher.Patcher[*core.ConfigMap](ctx, cachedStatus.ConfigMapsModInterface().V1(), obj, meta.PatchOptions{},
							patcher.PatchConfigMapData(elements)); err != nil {
							log.Err(err).Debug("Failed to patch GatewayConfig ConfigMap")
							return errors.WithStack(err)
						} else if changed {
							log.Str("service", obj.GetName()).Str("before", currentSha).Str("after", expectedSha).Info("Updated Member Config")
						}
					}
				}
			}
			return nil
		default:
			assertion.InvalidGroupKey.Assert(true, "Unable to create Member ConfigMap an unknown group: %s", members[id].Group.AsRole())
			return nil
		}
	}); err != nil {
		return errors.Section(err, "Member ConfigMap")
	}

	return nil
}

func (r *Resources) renderConfigMap(elements ...map[string]string) (map[string]string, error) {
	result := map[string]string{}

	for _, r := range elements {
		for k, v := range r {
			if _, ok := result[k]; ok {
				return nil, errors.Errorf("Key %s already defined", k)
			}

			result[k] = v
		}
	}

	if len(result) == 0 {
		return nil, nil
	}

	result[ConfigMapChecksumKey] = util.SHA256FromStringMap(result)

	return result, nil
}

func (r *Resources) renderMemberConfigElements(ctx context.Context, cachedStatus inspectorInterface.Inspector, member api.DeploymentStatusMemberElement, renders ...memberConfigMapRenderer) (map[string]string, error) {
	var elements = make([]map[string]string, len(renders))

	for _, r := range renders {
		if els, err := r(ctx, cachedStatus, member); err != nil {
			return nil, errors.Wrapf(err, "Unable to render CM for %s", member.Member.ID)
		} else {
			elements = append(elements, els)
		}
	}

	return r.renderConfigMap(elements...)
}
