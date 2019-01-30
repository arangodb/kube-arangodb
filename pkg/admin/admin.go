//
// DISCLAIMER
//
// Copyright 2018 ArangoDB GmbH, Cologne, Germany
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

package admin

import (
	"context"
	"fmt"
	"reflect"
	"time"

	driver "github.com/arangodb/go-driver"
	api "github.com/arangodb/kube-arangodb/pkg/apis/admin/v1alpha"
	dapi "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/generated/clientset/versioned"
	client "github.com/arangodb/kube-arangodb/pkg/generated/clientset/versioned/typed/admin/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/util/arangod"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/rs/zerolog"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/record"
)

// Dependencies holds dependent services for a DatabaseAdmin
type Dependencies struct {
	Log                zerolog.Logger
	KubeCli            kubernetes.Interface
	DatabaseAdminCRCli versioned.Interface
	EventRecorder      record.EventRecorder
}

// DeploymentProvider provides clients and api objects for deployments
type DeploymentProvider interface {
	GetClient(ctx context.Context, deployment, namespace string) (driver.Client, error)
	GetAPIObject(deployment, namespace string) (*dapi.ArangoDeployment, error)
}

// DatabaseAdmin contains all information about the resources in the current cluster
type DatabaseAdmin struct {
	ResourceChannel chan interface{}
	EventChannel    chan watch.Event
	Databases       map[string]*Database
	Namespace       string
	Dependencies    Dependencies
	Resources       map[string]DatabaseAdminResource
}

type KubeClient client.DatabaseadminV1alphaInterface

type DeploymentLinkedResource interface {
	GetDeploymentName(DeploymentNameResolver) (string, error)
}

type ReconcileContext interface {
	// AddDeploymentFinalizer adds a finalizer to the deployment this resource belongs to
	AddDeploymentFinalizer(obj APIObjectResource)
	// RemoveDeploymentFinalizer
	RemoveDeploymentFinalizer(obj APIObjectResource)

	// RemoveFinalizer removes the operator finalizer from the resource
	RemoveFinalizer(obj APIObjectResource)
	// AddFinalizer adds the operator finalizer to the resource
	AddFinalizer(obj APIObjectResource)
	// HasFinalizer return true if the operator finalizer is set
	HasFinalizer(obj APIObjectResource) bool

	// RemoveResourceFinalizer removes a finalizer from the database admin resource
	// This requires an update of the resource
	RemoveResourceFinalizer(res, finalizer string) error
	// AddResourceFinalizer adds a finalizer to the database admin resource
	// This requires an update of the resource
	AddResourceFinalizer(res, finalizer string) error

	ReportError(obj APIObjectResource, reason string, err error)
	ReportWarning(obj APIObjectResource, reason, message string)
	ReportEvent(obj APIObjectResource, reason, message string)

	SetCondition(obj APIObjectResource, condition api.ConditionType, status v1.ConditionStatus, reason, message string)
	RemoveCondition(obj APIObjectResource, condition api.ConditionType)

	GetCreatedAt(obj APIObjectResource) *metav1.Time
	SetCreatedAtNow(obj APIObjectResource)

	GetKubeSecret(obj APIObjectResource, name string) (map[string][]byte, error)
	CreateKubeSecret(obj APIObjectResource, name string, data map[string][]byte) error

	GetResource(name string) (DatabaseAdminResource, error)
	GetDatabaseResource(name string) (*Database, error)

	GetArangoClient(ctx context.Context, obj APIObjectResource) (driver.Client, error)
	GetArangoDatabaseClient(ctx context.Context, resourceName string) (driver.Database, error)
}

type DeploymentNameResolver interface {
	DeploymentByDatabase(database string) (string, error)
	//DatabaseByCollection(collection string) string
}

type ModifyObjectContext interface {
	ValidationError(err error)
	ResetImmutableFields(fields []string)
}

type DatabaseAdminResource interface {
	DeploymentLinkedResource
	Reconcile(ctx context.Context, rctx ReconcileContext)
	GetAPIObject() ArangoResource
	GetNamespace() string
	GetName() string

	Update(kube KubeClient) error
	UpdateStatus(kube KubeClient) error
	Load(kube KubeClient) (runtime.Object, error)
}

type ArangoResource interface {
	GetStatus() *api.ResourceStatus
	GetMeta() *metav1.ObjectMeta
}

type ArangoResourceSpec interface {
	Validate() error
	SetDefaults()
	ResetImmutableFields()
}

type APIObjectResource interface {
	DeploymentLinkedResource
	GetAPIObject() ArangoResource
	AsRuntimeObject() runtime.Object
	GetName() string
	AsOwner() metav1.OwnerReference
}

// NewDatabaseAdmin creates a new DatabaseAdmin
func NewDatabaseAdmin(Namespace string, deps Dependencies) *DatabaseAdmin {
	return &DatabaseAdmin{
		Databases:       make(map[string]*Database),
		Namespace:       Namespace,
		Dependencies:    deps,
		ResourceChannel: make(chan interface{}),
		EventChannel:    make(chan watch.Event),
		Resources:       make(map[string]DatabaseAdminResource),
	}
}

// GetAPIObject return the ArangoDeployment API object
func (da *DatabaseAdmin) GetAPIObject(deployment, namespace string) (*dapi.ArangoDeployment, error) {
	return da.Dependencies.DatabaseAdminCRCli.DatabaseV1alpha().ArangoDeployments(namespace).Get(deployment, metav1.GetOptions{})
}

// GetClient returns a database client for the given deployment
func (da *DatabaseAdmin) GetClient(ctx context.Context, deployment, namespace string) (driver.Client, error) {
	apiObject, err := da.GetAPIObject(deployment, namespace)
	if err == nil {
		return arangod.CreateArangodDatabaseClient(ctx, da.Dependencies.KubeCli.CoreV1(), apiObject, false)
	}

	return nil, err
}

// UpdateResource updates a resource
func (da *DatabaseAdmin) UpdateResource(r DatabaseAdminResource) {
	kube := da.Dependencies.DatabaseAdminCRCli.DatabaseadminV1alpha()

	if err := r.Update(kube); err != nil {
		da.Dependencies.Log.Error().Str("error", err.Error()).Msg("Failed to update resource")

		if k8sutil.IsConflict(err) {
			da.Dependencies.Log.Debug().Msg("Conflict - modify object")
			if api, err := r.Load(kube); err != nil {
				da.Dependencies.Log.Debug().Msg("Failed to load object")
			} else {
				da.ModifyObject(api)
			}
		}
	}
}

// UpdateResourceStatus updates the status of a resource
func (da *DatabaseAdmin) UpdateResourceStatus(r DatabaseAdminResource) {
	kube := da.Dependencies.DatabaseAdminCRCli.DatabaseadminV1alpha()

	if err := r.UpdateStatus(kube); err != nil {
		da.Dependencies.Log.Error().Str("error", err.Error()).Msg("Failed to update resource")

		if k8sutil.IsConflict(err) {
			da.Dependencies.Log.Debug().Msg("Conflict - modify object")
			if api, err := r.Load(kube); err != nil {
				da.Dependencies.Log.Debug().Msg("Failed to load object")
			} else {
				da.ModifyObject(api)
			}
		}
	}
}

// ReconcileResource reconciles the given resource
func (da *DatabaseAdmin) ReconcileResource(r DatabaseAdminResource) error {
	ctx := context.Background()

	//if db, ok := r.(*Database); ok {
	//	da.onDatabaseResourceUpdate(&db.apiObject)
	//}

	helper := ReconcileContextHelper{admin: da}

	//da.LoadResource(r)
	r.Reconcile(ctx, &helper)
	da.UpdateResource(r)
	// Check here if an update or updateStatus is required

	return nil
}

// CheckResources ensures all resources
func (da *DatabaseAdmin) CheckResources() {
	for name := range da.Resources {
		da.Dependencies.Log.Debug().Str("database", name).Msg("Reconciling")
		if err := da.ReconcileResource(da.Resources[name]); err != nil {
			da.Dependencies.Log.Error().Str("error", err.Error()).Msg("Failed to reconcile")
		}
	}
}

type ModifyObjectHandler struct {
	updateNeeded bool
}

func (da *DatabaseAdmin) ValidationError(err error) {
	da.Dependencies.Log.Error().Str("error", err.Error()).Msg("Failed to validate")
}

func (da *DatabaseAdmin) ResetImmutableFieldsError(res DatabaseAdminResource, fields []string) {
	// ReportError and replace the document with valid version
	da.Dependencies.Log.Error().Msgf("Reset immutable fields %v", fields)
}

func getObjectSpecValue(object interface{}) reflect.Value {
	return reflect.ValueOf(object).Elem().FieldByName("Spec")
}

func getAPIObjectSpecValue(object interface{}) reflect.Value {
	apip := reflect.ValueOf(object).MethodByName("GetAPIObject").Call([]reflect.Value{}) //.Elem().FieldByName("apiObject")
	if apip[0].IsValid() {
		return apip[0].Elem().Elem().FieldByName("Spec")
	}

	return reflect.Value{}
}

func reflectObjectName(object interface{}) (string, bool) {
	ov := reflect.ValueOf(object)
	if m := ov.MethodByName("GetName"); m.IsValid() {
		rv := m.Call([]reflect.Value{})

		if len(rv) == 1 {
			return rv[0].String(), true
		}
	}

	return "", false
}

// ModifyObject update a existing object representation
func (da *DatabaseAdmin) ModifyObject(object runtime.Object) {
	da.Dependencies.Log.Debug().Msg("ModifyObject")

	// Try to obtain the meta data of the object
	if name, ok := reflectObjectName(object); ok {
		res, found := da.Resources[name]
		if !found {
			// This is an add now
			da.Dependencies.Log.Debug().Msg("Unknown, redirecting")
			da.AddObject(object)
			return
		}

		resv := reflect.ValueOf(res)
		old := getObjectSpecValue(res)
		if !old.IsValid() {
			da.Dependencies.Log.Debug().Msg("old is invalid")
			return
		}
		new := getObjectSpecValue(object)
		if !new.IsValid() {
			da.Dependencies.Log.Debug().Msg("new is invalid")
			return
		}

		// Check if new and old are the same type
		if new.Type() != old.Type() {
			da.Dependencies.Log.Error().Msgf("ModifyObject has different types: %v and %v", new.Type(), old.Type())
			return
		}

		new.Addr().MethodByName("SetDefaultsFrom").Call([]reflect.Value{old.Addr()})
		returnv := old.Addr().MethodByName("ResetImmutableFields").Call([]reflect.Value{new.Addr()})
		if len(returnv) == 0 {
			da.Dependencies.Log.Debug().Msg("bad return ResetImmutableFields")
			return
		}

		forceUpdate := false

		fields := returnv[0]
		if fields.Len() > 0 {
			reflect.ValueOf(da.ResetImmutableFieldsError).Call([]reflect.Value{resv, fields})
			forceUpdate = true
			//return
		}

		valid := new.Addr().MethodByName("Validate").Call([]reflect.Value{})
		if !valid[0].IsNil() {
			reflect.ValueOf(da.ValidationError).Call([]reflect.Value{resv, valid[0]})
			// reset the spec to old spec
			reflect.ValueOf(object).Elem().FieldByName("Spec").Set(old)
			forceUpdate = true
		}
		// Update!
		resv.MethodByName("SetAPIObject").Call([]reflect.Value{reflect.ValueOf(object).Elem()})

		if forceUpdate {
			da.UpdateResource(res)
		}

		da.ReconcileResource(da.Resources[name])
		return
	}

	da.Dependencies.Log.Error().Msg("Failed to modify object - not metav1.Object")
}

type NamedResource interface {
	GetName() string
}

// NewFromObject creates a new ArangoResource from a runtime.Object
func NewFromObject(object runtime.Object) (DatabaseAdminResource, error) {
	switch object.(type) {
	case *api.ArangoDatabase:
		return NewDatabaseFromObject(object)
	case *api.ArangoUser:
		return NewUserFromObject(object)
	case *api.ArangoCollection:
		return NewCollectionFromObject(object)
	case *api.ArangoGraph:
		return NewGraphFromObject(object)
	default:
		return nil, fmt.Errorf("Unknown object type")
	}
}

// AddObject adds a new object depending on the type
func (da *DatabaseAdmin) AddObject(object runtime.Object) {

	// Try to obtain the meta data of the object
	if name, ok := reflectObjectName(object); ok {

		log := da.Dependencies.Log.Debug().Str("name", name)
		// Check here if such an object is known to us
		if _, found := da.Resources[name]; found {
			// This is an update now
			da.ModifyObject(object)
		}

		// This is a new object
		obj, err := NewFromObject(object)
		if err != nil {
			log.Str("error", err.Error()).Msg("Failed to add resource")
			return
		}

		da.Resources[name] = obj
		da.ReconcileResource(obj)
		return
	}

	da.Dependencies.Log.Error().Msg("Failed to add object - not metav1.Object")
}

// DeleteObject deletes a object without any checks
func (da *DatabaseAdmin) DeleteObject(object runtime.Object) {
	// Try to obtain the meta data of the object
	if name, ok := reflectObjectName(object); ok {
		// Check here if such an object is known to us
		if _, found := da.Resources[name]; found {
			// Just delete it, our finalizer is gone
			delete(da.Resources, name)
		}
	}
}

func (da *DatabaseAdmin) HandleWatchEvent(ev watch.Event) {
	switch ev.Type {
	case watch.Added:
		// A new object was added
		da.AddObject(ev.Object)
		break
	case watch.Modified:
		// A "known" resource was modified
		da.ModifyObject(ev.Object)
		break
	case watch.Deleted:
		// A "known" resource was deleted
		da.DeleteObject(ev.Object)
		break
	}
}

// Run runs the database admin
func (da *DatabaseAdmin) Run(stop <-chan struct{}) {
	for {
		select {
		case ev := <-da.EventChannel:
			da.HandleWatchEvent(ev)
			break
		case <-stop:
			close(da.ResourceChannel)
			return
		case <-time.After(5 * time.Second):
			da.Dependencies.Log.Debug().Msg("Hello there! Inspecting your deployments...")
			da.CheckResources()
			da.Dependencies.Log.Debug().Msg("Finished inspection.")
		}
	}
}

func (rch *ReconcileContextHelper) RemoveDeploymentFinalizer(obj APIObjectResource) {
	//deploymentName := obj.GetDeploymentName(da)
	rch.admin.Dependencies.Log.Debug().Msgf("RemoveDeploymentFinalizer(%s)", obj.GetName())
	// do stuff here
}

func (rch *ReconcileContextHelper) AddDeploymentFinalizer(obj APIObjectResource) {
	//deploymentName := obj.GetDeploymentName(da)
	rch.admin.Dependencies.Log.Debug().Msgf("AddDeploymentFinalizer(%s)", obj.GetName())
	// do stuff here
}

type ReconcileContextHelper struct {
	admin          *DatabaseAdmin
	updateRequired bool
}

func (rch *ReconcileContextHelper) RemoveFinalizer(obj APIObjectResource) {
	rch.admin.Dependencies.Log.Debug().Msgf("RemoveFinalizer(%s)", obj.GetName())
	meta := obj.GetAPIObject().GetMeta()
	for i, other := range meta.Finalizers {
		if other == rch.GetFinalizerName() {
			meta.Finalizers = append(meta.Finalizers[:i], meta.Finalizers[i+1:]...)
			return
		}
	}
}

func (rch *ReconcileContextHelper) AddFinalizer(obj APIObjectResource) {
	meta := obj.GetAPIObject().GetMeta()
	meta.Finalizers = append(meta.Finalizers, rch.GetFinalizerName())
}

func (rch *ReconcileContextHelper) GetFinalizerName() string {
	return "arango-database-admin"
}

func (rch *ReconcileContextHelper) HasFinalizer(obj APIObjectResource) bool {
	for _, f := range obj.GetAPIObject().GetMeta().Finalizers {
		if f == rch.GetFinalizerName() {
			return true
		}
	}
	return false
}

func (rch *ReconcileContextHelper) ReportError(obj APIObjectResource, reason string, err error) {
	rch.admin.Dependencies.Log.Debug().Str("reason", reason).Str("message", err.Error()).Msgf("ReportError(%s)", obj.GetName())
	rch.admin.Dependencies.EventRecorder.Event(obj.AsRuntimeObject(), v1.EventTypeWarning, reason, err.Error())
}

func (rch *ReconcileContextHelper) ReportWarning(obj APIObjectResource, reason, message string) {
	rch.admin.Dependencies.Log.Debug().Str("reason", reason).Str("message", message).Msgf("ReportWarning(%s)", obj.GetName())
	rch.admin.Dependencies.EventRecorder.Event(obj.AsRuntimeObject(), v1.EventTypeWarning, reason, message)
}

func (rch *ReconcileContextHelper) ReportEvent(obj APIObjectResource, reason, message string) {
	rch.admin.Dependencies.Log.Info().Str("reason", reason).Str("message", message).Msgf("ReportWarning(%s)", obj.GetName())
	rch.admin.Dependencies.EventRecorder.Event(obj.AsRuntimeObject(), v1.EventTypeNormal, reason, message)
}

func (rch *ReconcileContextHelper) SetCondition(obj APIObjectResource, condition api.ConditionType, status v1.ConditionStatus, reason, message string) {
	if obj.GetAPIObject().GetStatus().Conditions.SetCondition(condition, status, reason, message) {
		rch.updateRequired = true
	}
}

func (rch *ReconcileContextHelper) RemoveCondition(obj APIObjectResource, condition api.ConditionType) {
	obj.GetAPIObject().GetStatus().Conditions.RemoveCondition(condition)
	rch.updateRequired = true
}

// GetCreatedAt returns the created timestamp
func (rch *ReconcileContextHelper) GetCreatedAt(obj APIObjectResource) *metav1.Time {
	return obj.GetAPIObject().GetStatus().CreatedAt
}

// SetCreatedAtNow sets the created time to now
func (rch *ReconcileContextHelper) SetCreatedAtNow(obj APIObjectResource) {
	obj.GetAPIObject().GetStatus().CreatedAt = &metav1.Time{Time: time.Now()}
	rch.updateRequired = true
}

func (rch *ReconcileContextHelper) GetArangoClient(ctx context.Context, obj APIObjectResource) (driver.Client, error) {
	deployment, err := obj.GetDeploymentName(rch.admin)
	if err != nil {
		return nil, err
	}
	return rch.admin.GetClient(ctx, deployment, rch.admin.Namespace)
}

func (rch *ReconcileContextHelper) GetArangoDatabaseClient(ctx context.Context, databaseResourceName string) (driver.Database, error) {
	db, err := rch.GetDatabaseResource(databaseResourceName)
	if err != nil {
		return nil, err
	}
	client, err := rch.GetArangoClient(ctx, db)
	if err != nil {
		return nil, err
	}
	dbc, err := client.Database(ctx, db.Spec.GetName())
	if err != nil {
		return nil, err
	}
	return dbc, nil
}

func (rch *ReconcileContextHelper) GetKubeSecret(obj APIObjectResource, name string) (map[string][]byte, error) {
	secretsCli := rch.admin.Dependencies.KubeCli.CoreV1().Secrets(rch.admin.Namespace)
	secret, err := secretsCli.Get(name, metav1.GetOptions{})
	if err == nil {
		return secret.Data, nil
	}
	return nil, err
}

func (rch *ReconcileContextHelper) CreateKubeSecret(obj APIObjectResource, name string, data map[string][]byte) error {
	secretsCli := rch.admin.Dependencies.KubeCli.CoreV1().Secrets(rch.admin.Namespace)

	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:            name,
			OwnerReferences: []metav1.OwnerReference{obj.AsOwner()},
		},
		Data: data,
	}

	_, err := secretsCli.Create(secret)
	if err != nil {
		return err
	}
	return err
}

// GetResource returns the resource with given name, if found
func (rch *ReconcileContextHelper) GetResource(name string) (DatabaseAdminResource, error) {
	if res, ok := rch.admin.Resources[name]; ok {
		return res, nil
	}

	return nil, fmt.Errorf("Resource %s not found", name)
}

// GetDatabaseResource returns the database resource of the given name
// or an error if there is no such database resource
func (rch *ReconcileContextHelper) GetDatabaseResource(name string) (*Database, error) {
	res, err := rch.GetResource(name)
	if err != nil {
		return nil, err
	}
	if db, ok := res.(*Database); ok {
		return db, nil
	}
	return nil, fmt.Errorf("resource %s is not a database resource", name)
}

func (da *DatabaseAdmin) DeploymentByDatabase(database string) (string, error) {
	if res, ok := da.Resources[database]; ok {
		if db, ok := res.(*Database); ok {
			return db.GetDeploymentName(da)
		}
	}

	return "", fmt.Errorf("database resource %s unknown", database)
}

// RemoveResourceFinalizer removes a finalizer from the database admin resource
// This requires an update of the resource
func (rch *ReconcileContextHelper) RemoveResourceFinalizer(resname, finalizer string) error {
	res, err := rch.GetResource(resname)
	if err != nil {
		return err
	}
	meta := res.GetAPIObject().GetMeta()
	final := meta.GetFinalizers()
	new := make([]string, 0, len(final))
	for _, x := range final {
		if x != finalizer {
			new = append(new, x)
		}
	}

	if len(new) == len(final) {
		return nil // nothing changed
	}
	meta.SetFinalizers(new)
	rch.admin.UpdateResource(res)
	return nil
}

// AddResourceFinalizer adds a finalizer to the database admin resource
// This requires an update of the resource
func (rch *ReconcileContextHelper) AddResourceFinalizer(resname, finalizer string) error {
	res, err := rch.GetResource(resname)
	if err != nil {
		return err
	}
	meta := res.GetAPIObject().GetMeta()
	final := meta.GetFinalizers()
	for _, x := range final {
		if x == finalizer {
			return nil // Already set
		}
	}
	meta.SetFinalizers(append(final, finalizer))
	rch.admin.UpdateResource(res)
	return nil
}
