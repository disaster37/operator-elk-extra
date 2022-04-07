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
	componentFinalizer = "component.elk.k8s.webcenter.fr/finalizer"
	componentCondition = "UpdateIndexComponent"
)

// ElasticsearchComponentTemplateReconciler reconciles a ElasticsearchComponentTemplate object
type ElasticsearchComponentTemplateReconciler struct {
	Reconciler
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=elk.k8s.webcenter.fr,resources=elasticsearchcomponenttemplates,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=elk.k8s.webcenter.fr,resources=elasticsearchcomponenttemplates/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=elk.k8s.webcenter.fr,resources=elasticsearchcomponenttemplates/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ElasticsearchComponentTemplate object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *ElasticsearchComponentTemplateReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	reconciler, err := controller.NewStdReconciler(r.Client, componentFinalizer, r.reconciler, r.log, r.recorder, waitDurationWhenError)
	if err != nil {
		return ctrl.Result{}, err
	}

	component := &elkv1alpha1.ElasticsearchComponentTemplate{}
	data := map[string]any{}

	return reconciler.Reconcile(ctx, req, component, data)
}

// SetupWithManager sets up the controller with the Manager.
func (r *ElasticsearchComponentTemplateReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&elkv1alpha1.ElasticsearchComponentTemplate{}).
		Complete(r)
}

// Configure permit to init Elasticsearch handler
// It also permit to init condition
func (r *ElasticsearchComponentTemplateReconciler) Configure(ctx context.Context, req ctrl.Request, resource resource.Resource) (meta any, err error) {
	component := resource.(*elkv1alpha1.ElasticsearchComponentTemplate)

	// Init condition status if not exist
	if condition.FindStatusCondition(component.Status.Conditions, componentCondition) == nil {
		condition.SetStatusCondition(&component.Status.Conditions, v1.Condition{
			Type:   componentCondition,
			Status: v1.ConditionFalse,
			Reason: "Initialize",
		})
	}

	// Get elasticsearch handler / client
	meta, err = GetElasticsearchHandler(ctx, &component.Spec, r.Client, r.dinamicClient, req, r.log)
	if err != nil {
		r.recorder.Eventf(resource, core.EventTypeWarning, "Failed", "Unable to init elasticsearch handler: %s", err.Error())
		return nil, err
	}

	return meta, err
}

// Read permit to get current component template
func (r *ElasticsearchComponentTemplateReconciler) Read(ctx context.Context, resource resource.Resource, data map[string]any, meta any) (res ctrl.Result, err error) {
	component := resource.(*elkv1alpha1.ElasticsearchComponentTemplate)
	esHandler := meta.(elasticsearchhandler.ElasticsearchHandler)

	// Read component template from Elasticsearch
	currentComponent, err := esHandler.ComponentTemplateGet(component.Name)
	if err != nil {
		return res, errors.Wrap(err, "Unable to get component template from Elasticsearch")
	}

	data["component"] = currentComponent
	return res, nil
}

// Create add new component template
func (r *ElasticsearchComponentTemplateReconciler) Create(ctx context.Context, resource resource.Resource, data map[string]interface{}, meta interface{}) (res ctrl.Result, err error) {

	esHandler := meta.(elasticsearchhandler.ElasticsearchHandler)
	component := resource.(*elkv1alpha1.ElasticsearchComponentTemplate)

	// Create policy on Elasticsearch
	expectedComponent, err := component.ToComponentTemplate()
	if err != nil {
		return res, errors.Wrap(err, "Error when convert current component template to expected component template")
	}

	if err = esHandler.ComponentTemplateUpdate(component.Name, expectedComponent); err != nil {
		return res, errors.Wrap(err, "Error when update component template")
	}

	return res, nil
}

// Update permit to update component template from Elasticsearch
func (r *ElasticsearchComponentTemplateReconciler) Update(ctx context.Context, resource resource.Resource, data map[string]interface{}, meta interface{}) (res ctrl.Result, err error) {
	return r.Create(ctx, resource, data, meta)
}

// Delete permit to delete component from Elasticsearch
func (r *ElasticsearchComponentTemplateReconciler) Delete(ctx context.Context, resource resource.Resource, data map[string]interface{}, meta interface{}) (err error) {
	esHandler := meta.(elasticsearchhandler.ElasticsearchHandler)
	component := resource.(*elkv1alpha1.ElasticsearchComponentTemplate)

	if err = esHandler.ComponentTemplateDelete(component.Name); err != nil {
		return errors.Wrap(err, "Error when delete component template")
	}

	return nil

}

// Diff permit to check if diff between actual and expected component template exist
func (r *ElasticsearchComponentTemplateReconciler) Diff(resource resource.Resource, data map[string]interface{}, meta interface{}) (diff controller.Diff, err error) {
	esHandler := meta.(elasticsearchhandler.ElasticsearchHandler)
	component := resource.(*elkv1alpha1.ElasticsearchComponentTemplate)
	var currentComponent *olivere.IndicesGetComponentTemplateData
	var d any

	d, err = helper.Get(data, "component")
	if err != nil {
		return diff, err
	}
	currentComponent = d.(*olivere.IndicesGetComponentTemplateData)
	expectedComponent, err := component.ToComponentTemplate()
	if err != nil {
		return diff, err
	}

	diff = controller.Diff{
		NeedCreate: false,
		NeedUpdate: false,
	}

	if currentComponent == nil {
		diff.NeedCreate = true
		diff.Diff = "Component template not exist"
		return diff, nil
	}

	diffStr, err := esHandler.ComponentTemplateDiff(currentComponent, expectedComponent)
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
func (r *ElasticsearchComponentTemplateReconciler) OnError(ctx context.Context, resource resource.Resource, data map[string]any, meta any, err error) {
	component := resource.(*elkv1alpha1.ElasticsearchComponentTemplate)
	r.log.Error(err)
	r.recorder.Event(resource, core.EventTypeWarning, "Failed", err.Error())

	condition.SetStatusCondition(&component.Status.Conditions, v1.Condition{
		Type:    componentCondition,
		Status:  v1.ConditionFalse,
		Reason:  "Failed",
		Message: err.Error(),
	})
}

// OnSuccess permit to set status condition on the right state is everithink is good
func (r *ElasticsearchComponentTemplateReconciler) OnSuccess(ctx context.Context, resource resource.Resource, data map[string]any, meta any, diff controller.Diff) (err error) {
	component := resource.(*elkv1alpha1.ElasticsearchComponentTemplate)

	if diff.NeedCreate {
		condition.SetStatusCondition(&component.Status.Conditions, v1.Condition{
			Type:    componentCondition,
			Status:  v1.ConditionTrue,
			Reason:  "Success",
			Message: "Component template successfully created",
		})

		return nil
	}

	if diff.NeedUpdate {
		condition.SetStatusCondition(&component.Status.Conditions, v1.Condition{
			Type:    componentCondition,
			Status:  v1.ConditionTrue,
			Reason:  "Success",
			Message: "Component template successfully updated",
		})

		return nil
	}

	// Update condition status if needed
	if condition.IsStatusConditionPresentAndEqual(component.Status.Conditions, componentCondition, v1.ConditionFalse) {
		condition.SetStatusCondition(&component.Status.Conditions, v1.Condition{
			Type:    componentCondition,
			Reason:  "Success",
			Status:  v1.ConditionTrue,
			Message: "Component template already set",
		})

		r.recorder.Event(resource, core.EventTypeNormal, "Completed", "Component template already set")
	}

	return nil
}
