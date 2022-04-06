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
	watchFinalizer = "watch.elk.k8s.webcenter.fr/finalizer"
	watchCondition = "UpdateWatch"
)

// ElasticsearchWatcherReconciler reconciles a ElasticsearchWatcher object
type ElasticsearchWatcherReconciler struct {
	Reconciler
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=elk.k8s.webcenter.fr,resources=elasticsearchwatchers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=elk.k8s.webcenter.fr,resources=elasticsearchwatchers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=elk.k8s.webcenter.fr,resources=elasticsearchwatchers/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ElasticsearchWatcher object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *ElasticsearchWatcherReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	reconciler, err := controller.NewStdReconciler(r.Client, watchFinalizer, r.reconciler, r.log, r.recorder, waitDurationWhenError)
	if err != nil {
		return ctrl.Result{}, err
	}

	watch := &elkv1alpha1.ElasticsearchWatcher{}
	data := map[string]any{}

	return reconciler.Reconcile(ctx, req, watch, data)
}

// SetupWithManager sets up the controller with the Manager.
func (r *ElasticsearchWatcherReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&elkv1alpha1.ElasticsearchWatcher{}).
		Complete(r)
}

// Configure permit to init Elasticsearch handler
// It also permit to init condition
func (r *ElasticsearchWatcherReconciler) Configure(ctx context.Context, req ctrl.Request, resource resource.Resource) (meta any, err error) {
	watch := resource.(*elkv1alpha1.ElasticsearchWatcher)

	// Init condition status if not exist
	if condition.FindStatusCondition(watch.Status.Conditions, watchCondition) == nil {
		condition.SetStatusCondition(&watch.Status.Conditions, v1.Condition{
			Type:   watchCondition,
			Status: v1.ConditionFalse,
			Reason: "Initialize",
		})
	}

	// Get elasticsearch handler / client
	meta, err = GetElasticsearchHandler(ctx, &watch.Spec, r.Client, r.dinamicClient, req, r.log)
	if err != nil {
		r.recorder.Eventf(resource, core.EventTypeWarning, "Failed", "Unable to init elasticsearch handler: %s", err.Error())
		return nil, err
	}

	return meta, err
}

// Read permit to get current watch
func (r *ElasticsearchWatcherReconciler) Read(ctx context.Context, resource resource.Resource, data map[string]any, meta any) (res ctrl.Result, err error) {
	watch := resource.(*elkv1alpha1.ElasticsearchWatcher)
	esHandler := meta.(elasticsearchhandler.ElasticsearchHandler)

	// Read watch from Elasticsearch
	currentWatch, err := esHandler.WatchGet(watch.Name)
	if err != nil {
		return res, errors.Wrap(err, "Unable to get watch from Elasticsearch")
	}

	data["watch"] = currentWatch
	return res, nil
}

// Create add new watch
func (r *ElasticsearchWatcherReconciler) Create(ctx context.Context, resource resource.Resource, data map[string]interface{}, meta interface{}) (res ctrl.Result, err error) {

	esHandler := meta.(elasticsearchhandler.ElasticsearchHandler)
	watch := resource.(*elkv1alpha1.ElasticsearchWatcher)

	// Create watch on Elasticsearch
	expectedWatch, err := watch.ToWatch()
	if err != nil {
		return res, errors.Wrap(err, "Error when convert to watch")
	}
	if err = esHandler.WatchUpdate(watch.Name, expectedWatch); err != nil {
		return res, errors.Wrap(err, "Error when update watch")
	}

	return res, nil
}

// Update permit to update watch from Elasticsearch
func (r *ElasticsearchWatcherReconciler) Update(ctx context.Context, resource resource.Resource, data map[string]interface{}, meta interface{}) (res ctrl.Result, err error) {
	return r.Create(ctx, resource, data, meta)
}

// Delete permit to delete watch from Elasticsearch
func (r *ElasticsearchWatcherReconciler) Delete(ctx context.Context, resource resource.Resource, data map[string]interface{}, meta interface{}) (err error) {
	esHandler := meta.(elasticsearchhandler.ElasticsearchHandler)
	watch := resource.(*elkv1alpha1.ElasticsearchWatcher)

	if err = esHandler.WatchDelete(watch.Name); err != nil {
		return errors.Wrap(err, "Error when delete watch")
	}

	return nil

}

// Diff permit to check if diff between actual and expected watch exist
func (r *ElasticsearchWatcherReconciler) Diff(resource resource.Resource, data map[string]interface{}, meta interface{}) (diff controller.Diff, err error) {
	esHandler := meta.(elasticsearchhandler.ElasticsearchHandler)
	watch := resource.(*elkv1alpha1.ElasticsearchWatcher)
	var currentWatch *olivere.XPackWatch
	var d any

	d, err = helper.Get(data, "watch")
	if err != nil {
		return diff, err
	}
	currentWatch = d.(*olivere.XPackWatch)
	expectedWatch, err := watch.ToWatch()
	if err != nil {
		return diff, err
	}

	diff = controller.Diff{
		NeedCreate: false,
		NeedUpdate: false,
	}

	if currentWatch == nil {
		diff.NeedCreate = true
		diff.Diff = "Watch not exist"
		return diff, nil
	}

	diffStr, err := esHandler.WatchDiff(currentWatch, expectedWatch)
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
func (r *ElasticsearchWatcherReconciler) OnError(ctx context.Context, resource resource.Resource, data map[string]any, meta any, err error) {
	watch := resource.(*elkv1alpha1.ElasticsearchWatcher)
	r.log.Error(err)
	r.recorder.Event(resource, core.EventTypeWarning, "Failed", err.Error())

	condition.SetStatusCondition(&watch.Status.Conditions, v1.Condition{
		Type:    watchCondition,
		Status:  v1.ConditionFalse,
		Reason:  "Failed",
		Message: err.Error(),
	})
}

// OnSuccess permit to set status condition on the right state is everithink is good
func (r *ElasticsearchWatcherReconciler) OnSuccess(ctx context.Context, resource resource.Resource, data map[string]any, meta any, diff controller.Diff) (err error) {
	watch := resource.(*elkv1alpha1.ElasticsearchWatcher)

	if diff.NeedCreate {
		condition.SetStatusCondition(&watch.Status.Conditions, v1.Condition{
			Type:    watchCondition,
			Status:  v1.ConditionTrue,
			Reason:  "Success",
			Message: "Watch successfully created",
		})

		return nil
	}

	if diff.NeedUpdate {
		condition.SetStatusCondition(&watch.Status.Conditions, v1.Condition{
			Type:    watchCondition,
			Status:  v1.ConditionTrue,
			Reason:  "Success",
			Message: "Watch successfully updated",
		})

		return nil
	}

	// Update condition status if needed
	if condition.IsStatusConditionPresentAndEqual(watch.Status.Conditions, watchCondition, v1.ConditionFalse) {
		condition.SetStatusCondition(&watch.Status.Conditions, v1.Condition{
			Type:    watchCondition,
			Reason:  "Success",
			Status:  v1.ConditionTrue,
			Message: "Watch already set",
		})

		r.recorder.Event(resource, core.EventTypeNormal, "Completed", "Watch already set")
	}

	return nil
}
