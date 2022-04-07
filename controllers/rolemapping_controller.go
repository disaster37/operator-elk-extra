/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"

	core "k8s.io/api/core/v1"
	condition "k8s.io/apimachinery/pkg/api/meta"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	elkv1alpha1 "github.com/disaster37/operator-elk-extra/api/v1alpha1"
	"github.com/disaster37/operator-elk-extra/pkg/elasticsearchhandler"
	"github.com/disaster37/operator-sdk-extra/pkg/controller"
	"github.com/disaster37/operator-sdk-extra/pkg/helper"
	"github.com/disaster37/operator-sdk-extra/pkg/resource"
	olivere "github.com/olivere/elastic/v7"
	"github.com/pkg/errors"
)

const (
	roleMappingFinalizer = "role-mapping.elk.k8s.webcenter.fr/finalizer"
	roleMappingCondition = "UpdateRoleMapping"
)

// RoleMappingReconciler reconciles a RoleMapping object
type RoleMappingReconciler struct {
	Reconciler
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=elk.k8s.webcenter.fr,resources=rolemappings,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=elk.k8s.webcenter.fr,resources=rolemappings/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=elk.k8s.webcenter.fr,resources=rolemappings/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the RoleMapping object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *RoleMappingReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	reconciler, err := controller.NewStdReconciler(r.Client, roleMappingFinalizer, r.reconciler, r.log, r.recorder, waitDurationWhenError)
	if err != nil {
		return ctrl.Result{}, err
	}

	roleMapping := &elkv1alpha1.RoleMapping{}
	data := map[string]any{}

	return reconciler.Reconcile(ctx, req, roleMapping, data)
}

// SetupWithManager sets up the controller with the Manager.
func (r *RoleMappingReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&elkv1alpha1.RoleMapping{}).
		Complete(r)
}

// Configure permit to init Elasticsearch handler
// It also permit to init condition
func (r *RoleMappingReconciler) Configure(ctx context.Context, req ctrl.Request, resource resource.Resource) (meta any, err error) {
	roleMapping := resource.(*elkv1alpha1.RoleMapping)

	// Init condition status if not exist
	if condition.FindStatusCondition(roleMapping.Status.Conditions, roleMappingCondition) == nil {
		condition.SetStatusCondition(&roleMapping.Status.Conditions, v1.Condition{
			Type:   roleMappingCondition,
			Status: v1.ConditionFalse,
			Reason: "Initialize",
		})
	}

	// Get elasticsearch handler / client
	meta, err = GetElasticsearchHandler(ctx, &roleMapping.Spec, r.Client, r.dinamicClient, req, r.log)
	if err != nil {
		r.recorder.Eventf(resource, core.EventTypeWarning, "Failed", "Unable to init elasticsearch handler: %s", err.Error())
		return nil, err
	}

	return meta, err
}

// Read permit to get current role mapping
func (r *RoleMappingReconciler) Read(ctx context.Context, resource resource.Resource, data map[string]any, meta any) (res ctrl.Result, err error) {
	roleMapping := resource.(*elkv1alpha1.RoleMapping)
	esHandler := meta.(elasticsearchhandler.ElasticsearchHandler)

	// Read role mapping from Elasticsearch
	currentRoleMapping, err := esHandler.RoleMappingGet(roleMapping.Name)
	if err != nil {
		return res, errors.Wrap(err, "Unable to get role mapping from Elasticsearch")
	}

	data["roleMapping"] = currentRoleMapping
	return res, nil
}

// Create add role mapping
func (r *RoleMappingReconciler) Create(ctx context.Context, resource resource.Resource, data map[string]interface{}, meta interface{}) (res ctrl.Result, err error) {
	esHandler := meta.(elasticsearchhandler.ElasticsearchHandler)
	roleMapping := resource.(*elkv1alpha1.RoleMapping)

	// Create role mapping on Elasticsearch
	expectedRoleMapping, err := roleMapping.ToRoleMapping()
	if err != nil {
		return res, errors.Wrap(err, "Error when convert to role mapping")
	}
	if err = esHandler.RoleMappingUpdate(roleMapping.Name, expectedRoleMapping); err != nil {
		return res, errors.Wrap(err, "Error when update role mapping")
	}

	return res, nil
}

// Update permit to update role mapping from Elasticsearch
func (r *RoleMappingReconciler) Update(ctx context.Context, resource resource.Resource, data map[string]interface{}, meta interface{}) (res ctrl.Result, err error) {
	return r.Create(ctx, resource, data, meta)
}

// Delete permit to delete role mapping from Elasticsearch
func (r *RoleMappingReconciler) Delete(ctx context.Context, resource resource.Resource, data map[string]interface{}, meta interface{}) (err error) {
	esHandler := meta.(elasticsearchhandler.ElasticsearchHandler)
	roleMapping := resource.(*elkv1alpha1.RoleMapping)

	if err = esHandler.RoleMappingDelete(roleMapping.Name); err != nil {
		return errors.Wrap(err, "Error when delete role mapping")
	}

	return nil

}

// Diff permit to check if diff between actual and expected role mapping exist
func (r *RoleMappingReconciler) Diff(resource resource.Resource, data map[string]interface{}, meta interface{}) (diff controller.Diff, err error) {
	esHandler := meta.(elasticsearchhandler.ElasticsearchHandler)
	roleMapping := resource.(*elkv1alpha1.RoleMapping)
	var currentRoleMapping *olivere.XPackSecurityRoleMapping
	var d any

	d, err = helper.Get(data, "roleMapping")
	if err != nil {
		return diff, err
	}
	currentRoleMapping = d.(*olivere.XPackSecurityRoleMapping)
	expectedRoleMapping, err := roleMapping.ToRoleMapping()
	if err != nil {
		return diff, err
	}

	diff = controller.Diff{
		NeedCreate: false,
		NeedUpdate: false,
	}

	if currentRoleMapping == nil {
		diff.NeedCreate = true
		diff.Diff = "Role mapping not exist"
		return diff, nil
	}

	diffStr, err := esHandler.RoleMappingDiff(currentRoleMapping, expectedRoleMapping)
	if err != nil {
		return diff, err
	}

	if diffStr != "" {
		diff.NeedUpdate = true
		diff.Diff = diffStr
		return diff, nil
	}

	return
}

// OnError permit to set status condition on the right state and record error
func (r *RoleMappingReconciler) OnError(ctx context.Context, resource resource.Resource, data map[string]any, meta any, err error) {
	roleMapping := resource.(*elkv1alpha1.RoleMapping)
	r.log.Error(err)
	r.recorder.Event(resource, core.EventTypeWarning, "Failed", err.Error())

	condition.SetStatusCondition(&roleMapping.Status.Conditions, v1.Condition{
		Type:    roleMappingCondition,
		Status:  v1.ConditionFalse,
		Reason:  "Failed",
		Message: err.Error(),
	})
}

// OnSuccess permit to set status condition on the right state is everithink is good
func (r *RoleMappingReconciler) OnSuccess(ctx context.Context, resource resource.Resource, data map[string]any, meta any, diff controller.Diff) (err error) {
	roleMapping := resource.(*elkv1alpha1.RoleMapping)

	if diff.NeedCreate {
		condition.SetStatusCondition(&roleMapping.Status.Conditions, v1.Condition{
			Type:    roleMappingCondition,
			Status:  v1.ConditionTrue,
			Reason:  "Success",
			Message: "Role mapping successfully created",
		})

		return nil
	}

	if diff.NeedUpdate {
		condition.SetStatusCondition(&roleMapping.Status.Conditions, v1.Condition{
			Type:    roleMappingCondition,
			Status:  v1.ConditionTrue,
			Reason:  "Success",
			Message: "Role mapping successfully updated",
		})

		return nil
	}

	// Update condition status if needed
	if condition.IsStatusConditionPresentAndEqual(roleMapping.Status.Conditions, roleMappingCondition, v1.ConditionFalse) {
		condition.SetStatusCondition(&roleMapping.Status.Conditions, v1.Condition{
			Type:    roleMappingCondition,
			Reason:  "Success",
			Status:  v1.ConditionTrue,
			Message: "Role mapping already set",
		})

		r.recorder.Event(resource, core.EventTypeNormal, "Completed", "Role mapping already set")
	}

	return nil
}
