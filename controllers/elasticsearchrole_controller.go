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

	elkv1alpha1 "github.com/disaster37/operator-elk-extra/api/v1alpha1"
	"github.com/disaster37/operator-elk-extra/pkg/elasticsearchhandler"
	"github.com/disaster37/operator-sdk-extra/pkg/controller"
	"github.com/disaster37/operator-sdk-extra/pkg/helper"
	"github.com/disaster37/operator-sdk-extra/pkg/resource"
	olivere "github.com/olivere/elastic/v7"
	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	condition "k8s.io/apimachinery/pkg/api/meta"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	elasticsearchRoleFinalizer = "es-role.elk.k8s.webcenter.fr/finalizer"
	elasticsearchRoleCondition = "UpdateElasticsearchRole"
)

// ElasticsearchRoleReconciler reconciles a ElasticsearchRole object
type ElasticsearchRoleReconciler struct {
	Reconciler
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=elk.k8s.webcenter.fr,resources=elasticsearchroles,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=elk.k8s.webcenter.fr,resources=elasticsearchroles/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=elk.k8s.webcenter.fr,resources=elasticsearchroles/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ElasticsearchRole object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *ElasticsearchRoleReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	reconciler, err := controller.NewStdReconciler(r.Client, elasticsearchRoleFinalizer, r.reconciler, r.log, r.recorder, waitDurationWhenError)
	if err != nil {
		return ctrl.Result{}, err
	}

	role := &elkv1alpha1.ElasticsearchRole{}
	data := map[string]any{}

	return reconciler.Reconcile(ctx, req, role, data)
}

// SetupWithManager sets up the controller with the Manager.
func (r *ElasticsearchRoleReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&elkv1alpha1.ElasticsearchRole{}).
		Complete(r)
}

// Configure permit to init Elasticsearch handler
// It also permit to init condition
func (r *ElasticsearchRoleReconciler) Configure(ctx context.Context, req ctrl.Request, resource resource.Resource) (meta any, err error) {
	role := resource.(*elkv1alpha1.ElasticsearchRole)

	// Init condition status if not exist
	if condition.FindStatusCondition(role.Status.Conditions, elasticsearchRoleCondition) == nil {
		condition.SetStatusCondition(&role.Status.Conditions, v1.Condition{
			Type:   elasticsearchRoleCondition,
			Status: v1.ConditionFalse,
			Reason: "Initialize",
		})
	}

	// Get elasticsearch handler / client
	meta, err = GetElasticsearchHandler(ctx, &role.Spec, r.Client, r.dinamicClient, req, r.log)
	if err != nil {
		r.recorder.Eventf(resource, core.EventTypeWarning, "Failed", "Unable to init elasticsearch handler: %s", err.Error())
		return nil, err
	}

	return meta, err
}

// Read permit to get current elasticsearch role
func (r *ElasticsearchRoleReconciler) Read(ctx context.Context, resource resource.Resource, data map[string]any, meta any) (res ctrl.Result, err error) {
	role := resource.(*elkv1alpha1.ElasticsearchRole)
	esHandler := meta.(elasticsearchhandler.ElasticsearchHandler)

	// Read role from Elasticsearch
	currentRole, err := esHandler.RoleGet(role.Name)
	if err != nil {
		return res, errors.Wrap(err, "Unable to get role from Elasticsearch")
	}

	data["role"] = currentRole
	return res, nil
}

// Create add new elasticsearch role
func (r *ElasticsearchRoleReconciler) Create(ctx context.Context, resource resource.Resource, data map[string]interface{}, meta interface{}) (res ctrl.Result, err error) {
	esHandler := meta.(elasticsearchhandler.ElasticsearchHandler)
	role := resource.(*elkv1alpha1.ElasticsearchRole)

	// Create role on Elasticsearch
	expectedRole, err := role.ToRole()
	if err != nil {
		return res, errors.Wrap(err, "Error when convert to elasticsearch role")
	}
	if err = esHandler.RoleUpdate(role.Name, expectedRole); err != nil {
		return res, errors.Wrap(err, "Error when update elasticsearch role")
	}

	return res, nil
}

// Update permit to update role from Elasticsearch
func (r *ElasticsearchRoleReconciler) Update(ctx context.Context, resource resource.Resource, data map[string]interface{}, meta interface{}) (res ctrl.Result, err error) {
	return r.Create(ctx, resource, data, meta)
}

// Delete permit to delete role from Elasticsearch
func (r *ElasticsearchRoleReconciler) Delete(ctx context.Context, resource resource.Resource, data map[string]interface{}, meta interface{}) (err error) {
	esHandler := meta.(elasticsearchhandler.ElasticsearchHandler)
	role := resource.(*elkv1alpha1.ElasticsearchRole)

	if err = esHandler.RoleDelete(role.Name); err != nil {
		return errors.Wrap(err, "Error when delete elasticsearch role")
	}

	return nil

}

// Diff permit to check if diff between actual and expected role exist
func (r *ElasticsearchRoleReconciler) Diff(resource resource.Resource, data map[string]interface{}, meta interface{}) (diff controller.Diff, err error) {
	esHandler := meta.(elasticsearchhandler.ElasticsearchHandler)
	role := resource.(*elkv1alpha1.ElasticsearchRole)
	var currentRole *olivere.XPackSecurityRole
	var d any

	d, err = helper.Get(data, "role")
	if err != nil {
		return diff, err
	}
	currentRole = d.(*olivere.XPackSecurityRole)
	expectedRole, err := role.ToRole()
	if err != nil {
		return diff, err
	}

	diff = controller.Diff{
		NeedCreate: false,
		NeedUpdate: false,
	}

	if currentRole == nil {
		diff.NeedCreate = true
		diff.Diff = "Elasticsearch role not exist"
		return diff, nil
	}

	diffStr, err := esHandler.RoleDiff(currentRole, expectedRole)
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
func (r *ElasticsearchRoleReconciler) OnError(ctx context.Context, resource resource.Resource, data map[string]any, meta any, err error) {
	role := resource.(*elkv1alpha1.ElasticsearchRole)
	r.log.Error(err)
	r.recorder.Event(resource, core.EventTypeWarning, "Failed", err.Error())

	condition.SetStatusCondition(&role.Status.Conditions, v1.Condition{
		Type:    elasticsearchRoleCondition,
		Status:  v1.ConditionFalse,
		Reason:  "Failed",
		Message: err.Error(),
	})
}

// OnSuccess permit to set status condition on the right state is everithink is good
func (r *ElasticsearchRoleReconciler) OnSuccess(ctx context.Context, resource resource.Resource, data map[string]any, meta any, diff controller.Diff) (err error) {
	role := resource.(*elkv1alpha1.ElasticsearchRole)

	if diff.NeedCreate {
		condition.SetStatusCondition(&role.Status.Conditions, v1.Condition{
			Type:    elasticsearchRoleCondition,
			Status:  v1.ConditionTrue,
			Reason:  "Success",
			Message: "Role successfully created",
		})

		return nil
	}

	if diff.NeedUpdate {
		condition.SetStatusCondition(&role.Status.Conditions, v1.Condition{
			Type:    elasticsearchRoleCondition,
			Status:  v1.ConditionTrue,
			Reason:  "Success",
			Message: "Role successfully updated",
		})

		return nil
	}

	// Update condition status if needed
	if condition.IsStatusConditionPresentAndEqual(role.Status.Conditions, elasticsearchRoleCondition, v1.ConditionFalse) {
		condition.SetStatusCondition(&role.Status.Conditions, v1.Condition{
			Type:    elasticsearchRoleCondition,
			Reason:  "Success",
			Status:  v1.ConditionTrue,
			Message: "Role already set",
		})

		r.recorder.Event(resource, core.EventTypeNormal, "Completed", "Role already set")
	}

	return nil
}
