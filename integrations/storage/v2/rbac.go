//
// DISCLAIMER
//
// Copyright 2026 ArangoDB GmbH, Cologne, Germany
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

package v2

import (
	"context"

	"google.golang.org/grpc"

	pbImplAuthorizationV1Shared "github.com/arangodb/kube-arangodb/integrations/authorization/v1/shared"
	pbStorageV2 "github.com/arangodb/kube-arangodb/integrations/storage/v2/definition"
	utilConstantsContext "github.com/arangodb/kube-arangodb/pkg/util/constants/context"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/svc/authenticator"
)

const (
	// ActionWriteObject is the IAM action evaluated before writing an object.
	ActionWriteObject = "storage:WriteObject"
	// ActionReadObject is the IAM action evaluated before reading an object.
	ActionReadObject = "storage:ReadObject"
	// ActionHeadObject is the IAM action evaluated before reading object metadata.
	ActionHeadObject = "storage:HeadObject"
	// ActionDeleteObject is the IAM action evaluated before deleting an object.
	ActionDeleteObject = "storage:DeleteObject"
	// ActionListObjects is the IAM action evaluated before listing objects.
	ActionListObjects = "storage:ListObjects"
	// ActionInit is the IAM action evaluated before initializing the storage backend.
	ActionInit = "storage:Init"
)

// rbac wraps a parent StorageV2 integration and enforces RBAC using the authorization evaluator. The
// parent Interface is embedded, so every method is promoted from it by default; rbac overrides only the
// methods that need an authorization check.
type rbac struct {
	Interface

	authz pbImplAuthorizationV1Shared.Evaluator
}

var _ Interface = &rbac{}

// WithRBAC wraps the given StorageV2 integration with RBAC enforcement backed by the authorization
// evaluator provided in the context (AuthZClientPlugin).
func WithRBAC(ctx context.Context, in Interface) (Interface, error) {
	authz, ok := utilConstantsContext.AuthZClientPlugin.Get(ctx)
	if !ok {
		return nil, errors.Errorf("Unable to get AuthZ Client Plugin")
	}

	return &rbac{
		Interface: in,
		authz:     authz,
	}, nil
}

// Register registers the RBAC wrapper (not the wrapped parent) as the gRPC server, so incoming calls
// dispatch to the wrapper's authorization-checked overrides.
func (r *rbac) Register(registrar *grpc.Server) {
	pbStorageV2.RegisterStorageV2Server(registrar, r)
}

func (r *rbac) evaluate(ctx context.Context, action, resource string) error {
	return authenticator.GetIdentity(ctx).EvaluatePermission(ctx, r.authz, action, resource)
}

// writeObjectStream replays the first received message (and its error) to the wrapped WriteObject handler,
// so the RBAC wrapper can inspect the path for the authorization check without consuming it.
type writeObjectStream struct {
	pbStorageV2.StorageV2_WriteObjectServer

	first    *pbStorageV2.StorageV2WriteObjectRequest
	firstErr error
	replayed bool
}

func (s *writeObjectStream) Recv() (*pbStorageV2.StorageV2WriteObjectRequest, error) {
	if !s.replayed {
		s.replayed = true
		return s.first, s.firstErr
	}

	return s.StorageV2_WriteObjectServer.Recv()
}

func (r *rbac) WriteObject(server pbStorageV2.StorageV2_WriteObjectServer) error {
	msg, err := server.Recv()
	if err == nil {
		if path := msg.GetPath().GetPath(); path != "" {
			if cerr := r.evaluate(server.Context(), ActionWriteObject, path); cerr != nil {
				return cerr
			}
		}
	}

	// Replay the peeked first message (and error) so the parent performs its own path validation and
	// write loop unchanged.
	return r.Interface.WriteObject(&writeObjectStream{
		StorageV2_WriteObjectServer: server,
		first:                       msg,
		firstErr:                    err,
	})
}

func (r *rbac) ReadObject(req *pbStorageV2.StorageV2ReadObjectRequest, server pbStorageV2.StorageV2_ReadObjectServer) error {
	if err := r.evaluate(server.Context(), ActionReadObject, req.GetPath().GetPath()); err != nil {
		return err
	}

	return r.Interface.ReadObject(req, server)
}

func (r *rbac) HeadObject(ctx context.Context, req *pbStorageV2.StorageV2HeadObjectRequest) (*pbStorageV2.StorageV2HeadObjectResponse, error) {
	if err := r.evaluate(ctx, ActionHeadObject, req.GetPath().GetPath()); err != nil {
		return nil, err
	}

	return r.Interface.HeadObject(ctx, req)
}

func (r *rbac) DeleteObject(ctx context.Context, req *pbStorageV2.StorageV2DeleteObjectRequest) (*pbStorageV2.StorageV2DeleteObjectResponse, error) {
	if err := r.evaluate(ctx, ActionDeleteObject, req.GetPath().GetPath()); err != nil {
		return nil, err
	}

	return r.Interface.DeleteObject(ctx, req)
}

func (r *rbac) ListObjects(req *pbStorageV2.StorageV2ListObjectsRequest, server pbStorageV2.StorageV2_ListObjectsServer) error {
	if err := r.evaluate(server.Context(), ActionListObjects, req.GetPath().GetPath()); err != nil {
		return err
	}

	return r.Interface.ListObjects(req, server)
}

func (r *rbac) Init(ctx context.Context, in *pbStorageV2.StorageV2InitRequest) (*pbStorageV2.StorageV2InitResponse, error) {
	if err := r.evaluate(ctx, ActionInit, ""); err != nil {
		return nil, err
	}

	return r.Interface.Init(ctx, in)
}
