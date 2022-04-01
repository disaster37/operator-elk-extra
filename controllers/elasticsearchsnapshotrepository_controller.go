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
	olivere "github.com/olivere/elastic/v7"
	"github.com/pkg/errors"
)

const (
	repositoryFinalizer = "repository.elk.k8s.webcenter.fr/finalizer"
	repositoryCondition = "UpdateSnapshotRepository"
)

// ElasticsearchSnapshotRepositoryReconciler reconciles a ElasticsearchSnapshotRepository object
type ElasticsearchSnapshotRepositoryReconciler struct {
	Reconciler
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=elk.k8s.webcenter.fr,resources=elasticsearchsnapshotrepositories,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=elk.k8s.webcenter.fr,resources=elasticsearchsnapshotrepositories/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=elk.k8s.webcenter.fr,resources=elasticsearchsnapshotrepositories/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ElasticsearchSnapshotRepository object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *ElasticsearchSnapshotRepositoryReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	reconciler, err := controller.NewStdReconciler(r.Client, repositoryFinalizer, r.reconciler, r.log, r.recorder, waitDurationWhenError)
	if err != nil {
		return ctrl.Result{}, err
	}

	repository := &elkv1alpha1.ElasticsearchSnapshotRepository{}
	data := map[string]any{}

	return reconciler.Reconcile(ctx, req, repository, data)
}

// SetupWithManager sets up the controller with the Manager.
func (r *ElasticsearchSnapshotRepositoryReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&elkv1alpha1.ElasticsearchSnapshotRepository{}).
		Complete(r)
}

// Configure permit to init Elasticsearch handler
// It also permit to init condition
func (r *ElasticsearchSnapshotRepositoryReconciler) Configure(ctx context.Context, req ctrl.Request, resource resource.Resource) (meta any, err error) {
	repository := resource.(*elkv1alpha1.ElasticsearchSnapshotRepository)

	// Init condition status if not exist
	if condition.FindStatusCondition(repository.Status.Conditions, repositoryCondition) == nil {
		condition.SetStatusCondition(&repository.Status.Conditions, v1.Condition{
			Type:   repositoryCondition,
			Status: v1.ConditionFalse,
			Reason: "Initialize",
		})
	}

	// Get elasticsearch handler / client
	meta, err = GetElasticsearchHandler(ctx, &repository.Spec, r.Client, r.dinamicClient, req, r.log)
	if err != nil {
		r.recorder.Eventf(resource, core.EventTypeWarning, "Failed", "Unable to init elasticsearch handler: %s", err.Error())
		return nil, err
	}

	return meta, err
}

// Read permit to get current snapshot repository
func (r *ElasticsearchSnapshotRepositoryReconciler) Read(ctx context.Context, resource resource.Resource, data map[string]any, meta any) (res ctrl.Result, err error) {
	repository := resource.(*elkv1alpha1.ElasticsearchSnapshotRepository)
	esHandler := meta.(elasticsearchhandler.ElasticsearchHandler)

	// Read snapshot repository from Elasticsearch
	currentRepository, err := esHandler.SnapshotRepositoryGet(repository.Name)
	if err != nil {
		return res, errors.Wrap(err, "Unable to get snapshot repository from Elasticsearch")
	}

	data["repository"] = currentRepository
	return res, nil
}

// Create add new snapshot repository
func (r *ElasticsearchSnapshotRepositoryReconciler) Create(ctx context.Context, resource resource.Resource, data map[string]interface{}, meta interface{}) (res ctrl.Result, err error) {
	esHandler := meta.(elasticsearchhandler.ElasticsearchHandler)
	repository := resource.(*elkv1alpha1.ElasticsearchSnapshotRepository)

	settings := map[string]any{}
	if err = json.Unmarshal([]byte(repository.Spec.Settings), &settings); err != nil {
		return res, errors.Wrap(err, "Error when decode repository setting")
	}

	repoObj := &olivere.SnapshotRepositoryMetaData{
		Type:     repository.Spec.Type,
		Settings: settings,
	}

	// Create repository on Elasticsearch
	if err = esHandler.SnapshotRepositoryUpdate(repository.Name, repoObj); err != nil {
		return res, errors.Wrap(err, "Error when update snapshot repository")
	}

	return res, nil
}

// Update permit to update snappshot repository from Elasticsearch
func (r *ElasticsearchSnapshotRepositoryReconciler) Update(ctx context.Context, resource resource.Resource, data map[string]interface{}, meta interface{}) (res ctrl.Result, err error) {
	return r.Create(ctx, resource, data, meta)
}

// Delete permit to delete snapshot repository from Elasticsearch
func (r *ElasticsearchSnapshotRepositoryReconciler) Delete(ctx context.Context, resource resource.Resource, data map[string]interface{}, meta interface{}) (err error) {
	esHandler := meta.(elasticsearchhandler.ElasticsearchHandler)
	repository := resource.(*elkv1alpha1.ElasticsearchSnapshotRepository)

	if err = esHandler.SnapshotRepositoryDelete(repository.Name); err != nil {
		return errors.Wrap(err, "Error when delete snaoshot repository")
	}

	return nil

}

// Diff permit to check if diff between actual and expected snapshot repository
func (r *ElasticsearchSnapshotRepositoryReconciler) Diff(resource resource.Resource, data map[string]interface{}, meta interface{}) (diff controller.Diff, err error) {
	esHandler := meta.(elasticsearchhandler.ElasticsearchHandler)
	repository := resource.(*elkv1alpha1.ElasticsearchSnapshotRepository)
	var currentRepository *olivere.SnapshotRepositoryMetaData
	var d any

	settings := map[string]any{}
	if err = json.Unmarshal([]byte(repository.Spec.Settings), &settings); err != nil {
		return diff, errors.Wrap(err, "Error when decode repository setting")
	}
	expectedRepository := &olivere.SnapshotRepositoryMetaData{
		Type:     repository.Spec.Type,
		Settings: settings,
	}

	d, err = helper.Get(data, "repository")
	if err != nil {
		return diff, err
	}
	currentRepository = d.(*olivere.SnapshotRepositoryMetaData)

	diff = controller.Diff{
		NeedCreate: false,
		NeedUpdate: false,
	}

	if currentRepository == nil {
		diff.NeedCreate = true
		diff.Diff = "Snapshot repository not exist"
		return diff, nil
	}

	diffStr, err := esHandler.SnapshotRepositoryDiff(currentRepository, expectedRepository)
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
func (r *ElasticsearchSnapshotRepositoryReconciler) OnError(ctx context.Context, resource resource.Resource, data map[string]any, meta any, err error) {
	repository := resource.(*elkv1alpha1.ElasticsearchSnapshotRepository)
	r.log.Error(err)
	r.recorder.Event(resource, core.EventTypeWarning, "Failed", err.Error())

	condition.SetStatusCondition(&repository.Status.Conditions, v1.Condition{
		Type:    repositoryCondition,
		Status:  v1.ConditionFalse,
		Reason:  "Failed",
		Message: err.Error(),
	})
}

// OnSuccess permit to set status condition on the right state is everithink is good
func (r *ElasticsearchSnapshotRepositoryReconciler) OnSuccess(ctx context.Context, resource resource.Resource, data map[string]any, meta any, diff controller.Diff) (err error) {
	repository := resource.(*elkv1alpha1.ElasticsearchSnapshotRepository)

	if diff.NeedCreate {
		condition.SetStatusCondition(&repository.Status.Conditions, v1.Condition{
			Type:    repositoryCondition,
			Status:  v1.ConditionTrue,
			Reason:  "Success",
			Message: "Snapshot repository successfully created",
		})

		return nil
	}

	if diff.NeedUpdate {
		condition.SetStatusCondition(&repository.Status.Conditions, v1.Condition{
			Type:    repositoryCondition,
			Status:  v1.ConditionTrue,
			Reason:  "Success",
			Message: "Snapshot repository policy successfully updated",
		})

		return nil
	}

	// Update condition status if needed
	if condition.IsStatusConditionPresentAndEqual(repository.Status.Conditions, repositoryCondition, v1.ConditionFalse) {
		condition.SetStatusCondition(&repository.Status.Conditions, v1.Condition{
			Type:    repositoryCondition,
			Reason:  "Success",
			Status:  v1.ConditionTrue,
			Message: "Snapshot repository already set",
		})

		r.recorder.Event(resource, core.EventTypeNormal, "Completed", "Snapshot repository already set")
	}

	return nil
}
