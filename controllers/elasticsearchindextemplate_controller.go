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
	indexTemplateFinalizer = "index-template.elk.k8s.webcenter.fr/finalizer"
	indexTemplateCondition = "UpdateIndexTemplate"
)

// ElasticsearchIndexTemplateReconciler reconciles a ElasticsearchIndexTemplate object
type ElasticsearchIndexTemplateReconciler struct {
	Reconciler
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=elk.k8s.webcenter.fr,resources=elasticsearchindextemplates,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=elk.k8s.webcenter.fr,resources=elasticsearchindextemplates/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=elk.k8s.webcenter.fr,resources=elasticsearchindextemplates/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ElasticsearchIndexTemplate object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *ElasticsearchIndexTemplateReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	reconciler, err := controller.NewStdReconciler(r.Client, indexTemplateFinalizer, r.reconciler, r.log, r.recorder, waitDurationWhenError)
	if err != nil {
		return ctrl.Result{}, err
	}

	template := &elkv1alpha1.ElasticsearchIndexTemplate{}
	data := map[string]any{}

	return reconciler.Reconcile(ctx, req, template, data)
}

// SetupWithManager sets up the controller with the Manager.
func (r *ElasticsearchIndexTemplateReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&elkv1alpha1.ElasticsearchIndexTemplate{}).
		Complete(r)
}

// Configure permit to init Elasticsearch handler
// It also permit to init condition
func (r *ElasticsearchIndexTemplateReconciler) Configure(ctx context.Context, req ctrl.Request, resource resource.Resource) (meta any, err error) {
	template := resource.(*elkv1alpha1.ElasticsearchIndexTemplate)

	// Init condition status if not exist
	if condition.FindStatusCondition(template.Status.Conditions, indexTemplateCondition) == nil {
		condition.SetStatusCondition(&template.Status.Conditions, v1.Condition{
			Type:   indexTemplateCondition,
			Status: v1.ConditionFalse,
			Reason: "Initialize",
		})
	}

	// Get elasticsearch handler / client
	meta, err = GetElasticsearchHandler(ctx, &template.Spec, r.Client, r.dinamicClient, req, r.log)
	if err != nil {
		r.recorder.Eventf(resource, core.EventTypeWarning, "Failed", "Unable to init elasticsearch handler: %s", err.Error())
		return nil, err
	}

	return meta, err
}

// Read permit to get current index template
func (r *ElasticsearchIndexTemplateReconciler) Read(ctx context.Context, resource resource.Resource, data map[string]any, meta any) (res ctrl.Result, err error) {
	template := resource.(*elkv1alpha1.ElasticsearchIndexTemplate)
	esHandler := meta.(elasticsearchhandler.ElasticsearchHandler)

	// Read index template from Elasticsearch
	currentTemplate, err := esHandler.IndexTemplateGet(template.Name)
	if err != nil {
		return res, errors.Wrap(err, "Unable to get index template from Elasticsearch")
	}

	data["template"] = currentTemplate
	return res, nil
}

// Create add new index template
func (r *ElasticsearchIndexTemplateReconciler) Create(ctx context.Context, resource resource.Resource, data map[string]interface{}, meta interface{}) (res ctrl.Result, err error) {

	esHandler := meta.(elasticsearchhandler.ElasticsearchHandler)
	template := resource.(*elkv1alpha1.ElasticsearchIndexTemplate)

	// Create index template on Elasticsearch
	expectedTemplate, err := template.ToIndexTemplate()
	if err != nil {
		return res, errors.Wrap(err, "Error when convert to index template")
	}
	if err = esHandler.IndexTemplateUpdate(template.Name, expectedTemplate); err != nil {
		return res, errors.Wrap(err, "Error when update index template")
	}

	return res, nil
}

// Update permit to update index template from Elasticsearch
func (r *ElasticsearchIndexTemplateReconciler) Update(ctx context.Context, resource resource.Resource, data map[string]interface{}, meta interface{}) (res ctrl.Result, err error) {
	return r.Create(ctx, resource, data, meta)
}

// Delete permit to delete index template from Elasticsearch
func (r *ElasticsearchIndexTemplateReconciler) Delete(ctx context.Context, resource resource.Resource, data map[string]interface{}, meta interface{}) (err error) {
	esHandler := meta.(elasticsearchhandler.ElasticsearchHandler)
	template := resource.(*elkv1alpha1.ElasticsearchIndexTemplate)

	if err = esHandler.IndexTemplateDelete(template.Name); err != nil {
		return errors.Wrap(err, "Error when delete index template")
	}

	return nil

}

// Diff permit to check if diff between actual and expected index template exist
func (r *ElasticsearchIndexTemplateReconciler) Diff(resource resource.Resource, data map[string]interface{}, meta interface{}) (diff controller.Diff, err error) {
	esHandler := meta.(elasticsearchhandler.ElasticsearchHandler)
	template := resource.(*elkv1alpha1.ElasticsearchIndexTemplate)
	var currentTemplate *olivere.IndicesGetIndexTemplate
	var d any

	d, err = helper.Get(data, "template")
	if err != nil {
		return diff, err
	}
	currentTemplate = d.(*olivere.IndicesGetIndexTemplate)
	expectedTemplate, err := template.ToIndexTemplate()
	if err != nil {
		return diff, err
	}

	diff = controller.Diff{
		NeedCreate: false,
		NeedUpdate: false,
	}

	if currentTemplate == nil {
		diff.NeedCreate = true
		diff.Diff = "Index template not exist"
		return diff, nil
	}

	diffStr, err := esHandler.IndexTemplateDiff(currentTemplate, expectedTemplate)
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
func (r *ElasticsearchIndexTemplateReconciler) OnError(ctx context.Context, resource resource.Resource, data map[string]any, meta any, err error) {
	template := resource.(*elkv1alpha1.ElasticsearchIndexTemplate)
	r.log.Error(err)
	r.recorder.Event(resource, core.EventTypeWarning, "Failed", err.Error())

	condition.SetStatusCondition(&template.Status.Conditions, v1.Condition{
		Type:    indexTemplateCondition,
		Status:  v1.ConditionFalse,
		Reason:  "Failed",
		Message: err.Error(),
	})
}

// OnSuccess permit to set status condition on the right state is everithink is good
func (r *ElasticsearchIndexTemplateReconciler) OnSuccess(ctx context.Context, resource resource.Resource, data map[string]any, meta any, diff controller.Diff) (err error) {
	template := resource.(*elkv1alpha1.ElasticsearchIndexTemplate)

	if diff.NeedCreate {
		condition.SetStatusCondition(&template.Status.Conditions, v1.Condition{
			Type:    indexTemplateCondition,
			Status:  v1.ConditionTrue,
			Reason:  "Success",
			Message: "Index template successfully created",
		})

		return nil
	}

	if diff.NeedUpdate {
		condition.SetStatusCondition(&template.Status.Conditions, v1.Condition{
			Type:    indexTemplateCondition,
			Status:  v1.ConditionTrue,
			Reason:  "Success",
			Message: "Index template successfully updated",
		})

		return nil
	}

	// Update condition status if needed
	if condition.IsStatusConditionPresentAndEqual(template.Status.Conditions, indexTemplateCondition, v1.ConditionFalse) {
		condition.SetStatusCondition(&template.Status.Conditions, v1.Condition{
			Type:    indexTemplateCondition,
			Reason:  "Success",
			Status:  v1.ConditionTrue,
			Message: "Index template already set",
		})

		r.recorder.Event(resource, core.EventTypeNormal, "Completed", "Index template already set")
	}

	return nil
}
