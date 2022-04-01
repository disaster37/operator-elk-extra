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
	"encoding/json"

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
	"github.com/pkg/errors"
)

const (
	slmFinalizer = "slm.elk.k8s.webcenter.fr/finalizer"
	slmCondition = "UpdateSLMPolicy"
)

// ElasticsearchSLMReconciler reconciles a ElasticsearchSLM object
type ElasticsearchSLMReconciler struct {
	Reconciler
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=elk.k8s.webcenter.fr,resources=elasticsearchslms,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=elk.k8s.webcenter.fr,resources=elasticsearchslms/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=elk.k8s.webcenter.fr,resources=elasticsearchslms/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ElasticsearchSLM object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *ElasticsearchSLMReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	reconciler, err := controller.NewStdReconciler(r.Client, slmFinalizer, r.reconciler, r.log, r.recorder, waitDurationWhenError)
	if err != nil {
		return ctrl.Result{}, err
	}

	slm := &elkv1alpha1.ElasticsearchSLM{}
	data := map[string]any{}

	return reconciler.Reconcile(ctx, req, slm, data)
}

// SetupWithManager sets up the controller with the Manager.
func (r *ElasticsearchSLMReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&elkv1alpha1.ElasticsearchSLM{}).
		Complete(r)
}

// Configure permit to init Elasticsearch handler
// It also permit to init condition
func (r *ElasticsearchSLMReconciler) Configure(ctx context.Context, req ctrl.Request, resource resource.Resource) (meta any, err error) {
	slm := resource.(*elkv1alpha1.ElasticsearchSLM)

	// Init condition status if not exist
	if condition.FindStatusCondition(slm.Status.Conditions, slmCondition) == nil {
		condition.SetStatusCondition(&slm.Status.Conditions, v1.Condition{
			Type:   slmCondition,
			Status: v1.ConditionFalse,
			Reason: "Initialize",
		})
	}

	// Get elasticsearch handler / client
	meta, err = GetElasticsearchHandler(ctx, &slm.Spec, r.Client, r.dinamicClient, req, r.log)
	if err != nil {
		r.recorder.Eventf(resource, core.EventTypeWarning, "Failed", "Unable to init elasticsearch handler: %s", err.Error())
		return nil, err
	}

	return meta, err
}

// Read permit to get current SLM policy
func (r *ElasticsearchSLMReconciler) Read(ctx context.Context, resource resource.Resource, data map[string]any, meta any) (res ctrl.Result, err error) {
	slm := resource.(*elkv1alpha1.ElasticsearchSLM)
	esHandler := meta.(elasticsearchhandler.ElasticsearchHandler)

	// Read SLM policy from Elasticsearch
	slmPolicy, err := esHandler.SLMGet(slm.Name)
	if err != nil {
		return res, errors.Wrap(err, "Unable to get SLM policy from Elasticsearch")
	}

	data["policy"] = slmPolicy
	return res, nil
}

// Create add new SLM policy
func (r *ElasticsearchSLMReconciler) Create(ctx context.Context, resource resource.Resource, data map[string]interface{}, meta interface{}) (res ctrl.Result, err error) {

	esHandler := meta.(elasticsearchhandler.ElasticsearchHandler)
	slm := resource.(*elkv1alpha1.ElasticsearchSLM)
	policy := &elasticsearchhandler.SnapshotLifecyclePolicySpec{}

	// Before create policy, check if repository already exist
	repo, err := esHandler.SnapshotRepositoryGet(policy.Repository)
	if err != nil {
		return res, errors.Wrap(err, "Error when get snapshot repository to check if exist before create SLM policy")
	}
	if repo == nil {
		r.log.Warnf("Snapshot repository %s not yet exist, skip it", policy.Repository)
		r.recorder.Eventf(resource, core.EventTypeWarning, "Skip", "Snapshot repository %s not yet exist, wait it", policy.Repository)
		return ctrl.Result{RequeueAfter: waitDurationWhenError}, nil
	}

	// Create policy on Elasticsearch
	if err = json.Unmarshal([]byte(slm.Spec.Policy), &policy); err != nil {
		return res, errors.Wrap(err, "Error on Policy format")
	}
	if err = esHandler.SLMUpdate(slm.Name, policy); err != nil {
		return res, errors.Wrap(err, "Error when update policy")
	}

	return res, nil
}

// Update permit to update SLM from Elasticsearch
func (r *ElasticsearchSLMReconciler) Update(ctx context.Context, resource resource.Resource, data map[string]interface{}, meta interface{}) (res ctrl.Result, err error) {
	return r.Create(ctx, resource, data, meta)
}

// Delete permit to delete SLM from Elasticsearch
func (r *ElasticsearchSLMReconciler) Delete(ctx context.Context, resource resource.Resource, data map[string]interface{}, meta interface{}) (err error) {
	esHandler := meta.(elasticsearchhandler.ElasticsearchHandler)
	slm := resource.(*elkv1alpha1.ElasticsearchSLM)

	if err = esHandler.SLMDelete(slm.Name); err != nil {
		return errors.Wrap(err, "Error when delete policy")
	}

	return nil

}

// Diff permit to check if diff between actual and expected SLM exist
func (r *ElasticsearchSLMReconciler) Diff(resource resource.Resource, data map[string]interface{}, meta interface{}) (diff controller.Diff, err error) {
	esHandler := meta.(elasticsearchhandler.ElasticsearchHandler)
	slm := resource.(*elkv1alpha1.ElasticsearchSLM)
	expectedPolicy := &elasticsearchhandler.SnapshotLifecyclePolicySpec{}
	var currentPolicy *elasticsearchhandler.SnapshotLifecyclePolicySpec
	var d any

	d, err = helper.Get(data, "policy")
	if err != nil {
		return diff, err
	}
	currentPolicy = d.(*elasticsearchhandler.SnapshotLifecyclePolicySpec)
	if err = json.Unmarshal([]byte(slm.Spec.Policy), expectedPolicy); err != nil {
		return diff, err
	}

	diff = controller.Diff{
		NeedCreate: false,
		NeedUpdate: false,
	}

	if currentPolicy == nil {
		diff.NeedCreate = true
		diff.Diff = "SLM policy not exist"
		return diff, nil
	}

	diffStr, err := esHandler.SLMDiff(currentPolicy, expectedPolicy)
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
func (r *ElasticsearchSLMReconciler) OnError(ctx context.Context, resource resource.Resource, data map[string]any, meta any, err error) {
	slm := resource.(*elkv1alpha1.ElasticsearchSLM)
	r.log.Error(err)
	r.recorder.Event(resource, core.EventTypeWarning, "Failed", err.Error())

	condition.SetStatusCondition(&slm.Status.Conditions, v1.Condition{
		Type:    slmCondition,
		Status:  v1.ConditionFalse,
		Reason:  "Failed",
		Message: err.Error(),
	})
}

// OnSuccess permit to set status condition on the right state is everithink is good
func (r *ElasticsearchSLMReconciler) OnSuccess(ctx context.Context, resource resource.Resource, data map[string]any, meta any, diff controller.Diff) (err error) {
	slm := resource.(*elkv1alpha1.ElasticsearchSLM)

	if diff.NeedCreate {
		condition.SetStatusCondition(&slm.Status.Conditions, v1.Condition{
			Type:    slmCondition,
			Status:  v1.ConditionTrue,
			Reason:  "Success",
			Message: "SLM policy successfully created",
		})

		return nil
	}

	if diff.NeedUpdate {
		condition.SetStatusCondition(&slm.Status.Conditions, v1.Condition{
			Type:    slmCondition,
			Status:  v1.ConditionTrue,
			Reason:  "Success",
			Message: "SLM policy successfully updated",
		})

		return nil
	}

	// Update condition status if needed
	if condition.IsStatusConditionPresentAndEqual(slm.Status.Conditions, slmCondition, v1.ConditionFalse) {
		condition.SetStatusCondition(&slm.Status.Conditions, v1.Condition{
			Type:    slmCondition,
			Reason:  "Success",
			Status:  v1.ConditionTrue,
			Message: "SLM policy already set",
		})

		r.recorder.Event(resource, core.EventTypeNormal, "Completed", "SLM policy already set")
	}

	return nil
}
