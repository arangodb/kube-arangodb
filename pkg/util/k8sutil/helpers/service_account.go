//
// DISCLAIMER
//
// Copyright 2023 ArangoDB GmbH, Cologne, Germany
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

package helpers

import (
	"context"
	"fmt"
	"strings"

	"github.com/dchest/uniuri"
	core "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	sharedApi "github.com/arangodb/kube-arangodb/pkg/apis/shared/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/patch"
	operator "github.com/arangodb/kube-arangodb/pkg/operatorV2"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/kerrors"
)

func EnsureServiceAccount(ctx context.Context, client kubernetes.Interface, owner meta.OwnerReference, obj *sharedApi.ServiceAccount, name, namespace string, role, clusterRole []rbac.PolicyRule) (bool, error) {
	if obj == nil {
		return false, errors.Newf("Object reference cannot be nil")
	}

	// Check if secret exists
	if obj.Object != nil {
		if sa, err := util.WithKubernetesContextTimeoutP2A2(ctx, client.CoreV1().ServiceAccounts(namespace).Get, obj.Object.GetName(), meta.GetOptions{}); err != nil {
			if !kerrors.IsNotFound(err) {
				return false, err
			}

			obj.Object = nil

			return true, operator.Reconcile("SA is missing")
		} else {
			if !obj.Object.Equals(sa) {
				// Invalid object
				if err := util.WithKubernetesContextTimeoutP1A2(ctx, client.CoreV1().ServiceAccounts(namespace).Delete, name, meta.DeleteOptions{
					Preconditions: meta.NewUIDPreconditions(string(obj.Object.GetUID())),
				}); err != nil {
					if !kerrors.IsNotFound(err) && !kerrors.IsConflict(err) {
						return false, err
					}
				}

				obj.Object = nil

				return true, operator.Reconcile("Removing SA")
			}
		}
	}

	if obj.Object == nil {
		if sa, err := util.WithKubernetesContextTimeoutP2A2(ctx, client.CoreV1().ServiceAccounts(namespace).Create, &core.ServiceAccount{
			ObjectMeta: meta.ObjectMeta{
				OwnerReferences: []meta.OwnerReference{owner},

				Name:      fmt.Sprintf("%s-%s", name, strings.ToLower(uniuri.NewLen(6))),
				Namespace: namespace,
			},
		}, meta.CreateOptions{}); err != nil {
			return false, err
		} else {
			obj.Object = util.NewType(sharedApi.NewObject(sa))
			return true, operator.Reconcile("Created SA")
		}
	}

	// ROLE

	if len(role) == 0 {
		// Ensure role and binding is missing
		if obj.Namespaced != nil {
			if obj.Namespaced.Binding != nil {
				// Remove binding
				if err := util.WithKubernetesContextTimeoutP1A2(ctx, client.RbacV1().RoleBindings(namespace).Delete, obj.Namespaced.Binding.GetName(), meta.DeleteOptions{
					Preconditions: meta.NewUIDPreconditions(string(obj.Namespaced.Binding.GetUID())),
				}); err != nil {
					if !kerrors.IsNotFound(err) && !kerrors.IsConflict(err) {
						return false, err
					}
				}

				obj.Namespaced.Binding = nil
				return true, operator.Reconcile("Removing RoleBinding")
			}

			if obj.Namespaced.Role != nil {
				// Remove binding
				if err := util.WithKubernetesContextTimeoutP1A2(ctx, client.RbacV1().Roles(namespace).Delete, obj.Namespaced.Role.GetName(), meta.DeleteOptions{
					Preconditions: meta.NewUIDPreconditions(string(obj.Namespaced.Role.GetUID())),
				}); err != nil {
					if !kerrors.IsNotFound(err) && !kerrors.IsConflict(err) {
						return false, err
					}
				}

				obj.Namespaced.Role = nil
				return true, operator.Reconcile("Removing Role")
			}

			obj.Namespaced = nil
			return true, operator.Reconcile("Removing Namespaced Handler")
		}
	} else {
		// Create if required
		if obj.Namespaced == nil || obj.Namespaced.Role == nil {
			if r, err := util.WithKubernetesContextTimeoutP2A2(ctx, client.RbacV1().Roles(namespace).Create, &rbac.Role{
				ObjectMeta: meta.ObjectMeta{
					OwnerReferences: []meta.OwnerReference{owner},

					Name:      fmt.Sprintf("%s-%s", name, strings.ToLower(uniuri.NewLen(6))),
					Namespace: namespace,
				},
				Rules: role,
			}, meta.CreateOptions{}); err != nil {
				return false, err
			} else {
				if obj.Namespaced == nil {
					obj.Namespaced = &sharedApi.ServiceAccountRole{}
				}
				obj.Namespaced.Role = util.NewType(sharedApi.NewObject(r))
				return true, operator.Reconcile("Created Role")
			}
		}

		if obj.Namespaced == nil || obj.Namespaced.Binding == nil {
			if r, err := util.WithKubernetesContextTimeoutP2A2(ctx, client.RbacV1().RoleBindings(namespace).Create, &rbac.RoleBinding{
				ObjectMeta: meta.ObjectMeta{
					OwnerReferences: []meta.OwnerReference{owner},

					Name:      fmt.Sprintf("%s-%s", name, strings.ToLower(uniuri.NewLen(6))),
					Namespace: namespace,
				},
				RoleRef: rbac.RoleRef{
					APIGroup: "rbac.authorization.k8s.io",
					Kind:     "Role",
					Name:     obj.Namespaced.Role.GetName(),
				},
				Subjects: []rbac.Subject{
					{
						Kind:      "ServiceAccount",
						APIGroup:  "",
						Name:      obj.Object.GetName(),
						Namespace: namespace,
					},
				},
			}, meta.CreateOptions{}); err != nil {
				return false, err
			} else {
				if obj.Namespaced == nil {
					obj.Namespaced = &sharedApi.ServiceAccountRole{}
				}
				obj.Namespaced.Binding = util.NewType(sharedApi.NewObject(r))
				return true, operator.Reconcile("Created RoleBinding")
			}
		}

		// Both object are nil, lets validate aspects
		if r, err := util.WithKubernetesContextTimeoutP2A2(ctx, client.RbacV1().Roles(namespace).Get, obj.Namespaced.Role.GetName(), meta.GetOptions{}); err != nil {
			if !kerrors.IsNotFound(err) {
				return false, err
			}

			obj.Namespaced.Role = nil
			return true, operator.Reconcile("Missing Role")
		} else {
			if !obj.Namespaced.Role.Equals(r) {
				if err := util.WithKubernetesContextTimeoutP1A2(ctx, client.RbacV1().Roles(namespace).Delete, obj.Namespaced.Role.GetName(), meta.DeleteOptions{
					Preconditions: meta.NewUIDPreconditions(string(obj.Namespaced.Role.GetUID())),
				}); err != nil {
					if !kerrors.IsNotFound(err) && !kerrors.IsConflict(err) {
						return false, err
					}
				}

				obj.Namespaced.Role = nil
				return true, operator.Reconcile("Recreate Role")
			}

			if !equality.Semantic.DeepEqual(r.Rules, role) {
				// There is change in the roles
				if _, err := util.WithKubernetesPatch[*rbac.Role](ctx, obj.Namespaced.Role.GetName(), client.RbacV1().Roles(namespace), patch.ItemReplace(patch.NewPath("rules"), role)); err != nil {
					if !kerrors.IsNotFound(err) {
						return false, err
					}
				}

				return false, operator.Reconcile("Refresh Role")
			}
		}

		if r, err := util.WithKubernetesContextTimeoutP2A2(ctx, client.RbacV1().RoleBindings(namespace).Get, obj.Namespaced.Binding.GetName(), meta.GetOptions{}); err != nil {
			if !kerrors.IsNotFound(err) {
				return false, err
			}

			obj.Namespaced.Role = nil
			return true, operator.Reconcile("Missing RoleBinding")
		} else {
			if !obj.Namespaced.Binding.Equals(r) {
				if err := util.WithKubernetesContextTimeoutP1A2(ctx, client.RbacV1().RoleBindings(namespace).Delete, obj.Namespaced.Binding.GetName(), meta.DeleteOptions{
					Preconditions: meta.NewUIDPreconditions(string(obj.Namespaced.Role.GetUID())),
				}); err != nil {
					if !kerrors.IsNotFound(err) && !kerrors.IsConflict(err) {
						return false, err
					}
				}

				obj.Namespaced.Role = nil
				return true, operator.Reconcile("Recreate RoleBinding")
			}
		}
	}

	// CLUSTER ROLE

	if len(clusterRole) == 0 {
		// Ensure role and binding is missing
		if obj.Cluster != nil {
			if obj.Cluster.Binding != nil {
				// Remove binding
				if err := util.WithKubernetesContextTimeoutP1A2(ctx, client.RbacV1().ClusterRoleBindings().Delete, obj.Cluster.Binding.GetName(), meta.DeleteOptions{
					Preconditions: meta.NewUIDPreconditions(string(obj.Cluster.Binding.GetUID())),
				}); err != nil {
					if !kerrors.IsNotFound(err) && !kerrors.IsConflict(err) {
						return false, err
					}
				}

				obj.Cluster.Binding = nil
				return true, operator.Reconcile("Removing ClusterRoleBinding")
			}

			if obj.Cluster.Role != nil {
				// Remove binding
				if err := util.WithKubernetesContextTimeoutP1A2(ctx, client.RbacV1().ClusterRoles().Delete, obj.Cluster.Role.GetName(), meta.DeleteOptions{
					Preconditions: meta.NewUIDPreconditions(string(obj.Cluster.Role.GetUID())),
				}); err != nil {
					if !kerrors.IsNotFound(err) && !kerrors.IsConflict(err) {
						return false, err
					}
				}

				obj.Cluster.Role = nil
				return true, operator.Reconcile("Removing ClusterRole")
			}

			obj.Cluster = nil
			return true, operator.Reconcile("Removing Cluster Handler")
		}
	} else {
		// Create if required
		if obj.Cluster == nil || obj.Cluster.Role == nil {
			if r, err := util.WithKubernetesContextTimeoutP2A2(ctx, client.RbacV1().ClusterRoles().Create, &rbac.ClusterRole{
				ObjectMeta: meta.ObjectMeta{
					OwnerReferences: []meta.OwnerReference{owner},

					Name: fmt.Sprintf("%s-%s", name, strings.ToLower(uniuri.NewLen(6))),
				},
				Rules: clusterRole,
			}, meta.CreateOptions{}); err != nil {
				return false, err
			} else {
				if obj.Cluster == nil {
					obj.Cluster = &sharedApi.ServiceAccountRole{}
				}
				obj.Cluster.Role = util.NewType(sharedApi.NewObject(r))
				return true, operator.Reconcile("Created ClusterRole")
			}
		}

		if obj.Cluster == nil || obj.Cluster.Binding == nil {
			if r, err := util.WithKubernetesContextTimeoutP2A2(ctx, client.RbacV1().ClusterRoleBindings().Create, &rbac.ClusterRoleBinding{
				ObjectMeta: meta.ObjectMeta{
					OwnerReferences: []meta.OwnerReference{owner},

					Name: fmt.Sprintf("%s-%s", name, strings.ToLower(uniuri.NewLen(6))),
				},
				RoleRef: rbac.RoleRef{
					APIGroup: "rbac.authorization.k8s.io",
					Kind:     "ClusterRole",
					Name:     obj.Cluster.Role.GetName(),
				},
				Subjects: []rbac.Subject{
					{
						Kind:      "ServiceAccount",
						APIGroup:  "",
						Name:      obj.Object.GetName(),
						Namespace: namespace,
					},
				},
			}, meta.CreateOptions{}); err != nil {
				return false, err
			} else {
				if obj.Cluster == nil {
					obj.Cluster = &sharedApi.ServiceAccountRole{}
				}
				obj.Cluster.Binding = util.NewType(sharedApi.NewObject(r))
				return true, operator.Reconcile("Created ClusterRoleBinding")
			}
		}

		// Both object are nil, lets validate aspects
		if r, err := util.WithKubernetesContextTimeoutP2A2(ctx, client.RbacV1().ClusterRoles().Get, obj.Cluster.Role.GetName(), meta.GetOptions{}); err != nil {
			if !kerrors.IsNotFound(err) {
				return false, err
			}

			obj.Cluster.Role = nil
			return true, operator.Reconcile("Missing ClusterRole")
		} else {
			if !obj.Cluster.Role.Equals(r) {
				if err := util.WithKubernetesContextTimeoutP1A2(ctx, client.RbacV1().ClusterRoles().Delete, obj.Cluster.Role.GetName(), meta.DeleteOptions{
					Preconditions: meta.NewUIDPreconditions(string(obj.Cluster.Role.GetUID())),
				}); err != nil {
					if !kerrors.IsNotFound(err) && !kerrors.IsConflict(err) {
						return false, err
					}
				}

				obj.Cluster.Role = nil
				return true, operator.Reconcile("Recreate ClusterRole")
			}

			if !equality.Semantic.DeepEqual(r.Rules, clusterRole) {
				// There is change in the roles
				if _, err := util.WithKubernetesPatch[*rbac.ClusterRole](ctx, obj.Cluster.Role.GetName(), client.RbacV1().ClusterRoles(), patch.ItemReplace(patch.NewPath("rules"), clusterRole)); err != nil {
					if !kerrors.IsNotFound(err) {
						return false, err
					}
				}

				return false, operator.Reconcile("Refresh ClusterRole")
			}
		}

		if r, err := util.WithKubernetesContextTimeoutP2A2(ctx, client.RbacV1().ClusterRoleBindings().Get, obj.Cluster.Binding.GetName(), meta.GetOptions{}); err != nil {
			if !kerrors.IsNotFound(err) {
				return false, err
			}

			obj.Cluster.Role = nil
			return true, operator.Reconcile("Missing ClusterRoleBinding")
		} else {
			if !obj.Cluster.Binding.Equals(r) {
				if err := util.WithKubernetesContextTimeoutP1A2(ctx, client.RbacV1().ClusterRoleBindings().Delete, obj.Cluster.Binding.GetName(), meta.DeleteOptions{
					Preconditions: meta.NewUIDPreconditions(string(obj.Cluster.Role.GetUID())),
				}); err != nil {
					if !kerrors.IsNotFound(err) && !kerrors.IsConflict(err) {
						return false, err
					}
				}

				obj.Cluster.Role = nil
				return true, operator.Reconcile("Recreate ClusterRoleBinding")
			}
		}
	}

	return false, nil
}
