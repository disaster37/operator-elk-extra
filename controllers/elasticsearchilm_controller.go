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
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	elkv1alpha1 "github.com/disaster37/operator-elk-extra/api/v1alpha1"
	"github.com/disaster37/operator-elk-extra/pkg/elasticsearchhandler"
	"github.com/disaster37/operator-sdk-extra/pkg/controller"
	"github.com/disaster37/operator-sdk-extra/pkg/helper"
	"github.com/disaster37/operator-sdk-extra/pkg/resource"
	"github.com/sirupsen/logrus"
)

const (
	ilmFinalizer = "ilm.elk.k8s.webcenter.fr/finalizer"
	ilmCondition = "UpdateILMPolicy"
)

// ElasticsearchILMReconciler reconciles a ElasticsearchILM object
type ElasticsearchILMReconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	recorder   record.EventRecorder
	log        *logrus.Entry
	reconciler controller.Reconciler
}

//+kubebuilder:rbac:groups=elk.k8s.webcenter.fr,resources=elasticsearchilms,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=elk.k8s.webcenter.fr,resources=elasticsearchilms/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=elk.k8s.webcenter.fr,resources=elasticsearchilms/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ElasticsearchILM object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *ElasticsearchILMReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	reconciler, err := controller.NewStdReconciler(r.Client, ilmFinalizer, r.reconciler, r.log, r.recorder, waitDurationWhenError)
	if err != nil {
		return ctrl.Result{}, err
	}

	ilm := &elkv1alpha1.ElasticsearchILM{}
	data := map[string]any{}

	return reconciler.Reconcile(ctx, req, ilm, data)
}

// SetupWithManager sets up the controller with the Manager.
func (r *ElasticsearchILMReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&elkv1alpha1.ElasticsearchILM{}).
		Complete(r)
}

func (r *ElasticsearchILMReconciler) SetLogger(log *logrus.Entry) {
	r.log = log
}

func (r *ElasticsearchILMReconciler) SetRecorder(recorder record.EventRecorder) {
	r.recorder = recorder
}

// Configure permit to init Elasticsearch handler
func (r *ElasticsearchILMReconciler) Configure(ctx context.Context, req ctrl.Request, resource resource.Resource) (meta any, err error) {
	// Get elasticsearch handler / client
	ilm := resource.(*elkv1alpha1.ElasticsearchILM)
	meta, err = GetElasticsearchHandler(ctx, &ilm.Spec, r.Client, req, r.log)
	if err != nil {
		r.recorder.Eventf(resource, core.EventTypeWarning, "Failed", "Unable to init elasticsearch handler: %s", err.Error())
		return nil, err
	}

	return meta, err
}

// Read permit to get current ILM policy
func (r *ElasticsearchILMReconciler) Read(ctx context.Context, resource resource.Resource, data map[string]any, meta any) (res ctrl.Result, err error) {
	ilm := resource.(*elkv1alpha1.ElasticsearchILM)
	esHandler := meta.(elasticsearchhandler.ElasticsearchHandler)

	// Init condition status if not exist
	if condition.FindStatusCondition(ilm.Status.Conditions, ilmCondition) == nil {
		condition.SetStatusCondition(&ilm.Status.Conditions, v1.Condition{
			Type:   ilmCondition,
			Status: v1.ConditionFalse,
			Reason: "Initialize",
		})
	}

	// Read ILM policy from Elasticsearch
	ilmPolicy, err := esHandler.ILMGet(ilm.Name)
	if err != nil {
		r.recorder.Eventf(resource, core.EventTypeWarning, "Failed", "Unable to get ILM policy from Elasticsearch: %s", err.Error())
		return res, err
	}
	data["policy"] = ilmPolicy
	return res, nil
}

// Create add new ILM policy
func (r *ElasticsearchILMReconciler) Create(ctx context.Context, resource resource.Resource, data map[string]interface{}, meta interface{}) (res ctrl.Result, err error) {

	esHandler := meta.(elasticsearchhandler.ElasticsearchHandler)
	ilm := resource.(*elkv1alpha1.ElasticsearchILM)
	policy := make(map[string]any)

	// handler condition status if error
	defer func() {
		if err != nil {
			condition.SetStatusCondition(&ilm.Status.Conditions, v1.Condition{
				Type:    ilmCondition,
				Status:  v1.ConditionFalse,
				Reason:  "Failed",
				Message: err.Error(),
			})
		}
	}()

	// Create policy on Elasticsearch
	if err = json.Unmarshal([]byte(ilm.Spec.Policy), &policy); err != nil {
		return res, err
	}
	if err = esHandler.ILMUpdate(ilm.Name, policy); err != nil {
		return res, err
	}

	// Update status
	condition.SetStatusCondition(&ilm.Status.Conditions, v1.Condition{
		Type:    ilmCondition,
		Status:  v1.ConditionTrue,
		Reason:  "Success",
		Message: "ILM successfully updated",
	})

	return res, nil
}

// Update permit to update ILM from Elasticsearch
func (r *ElasticsearchILMReconciler) Update(ctx context.Context, resource resource.Resource, data map[string]interface{}, meta interface{}) (res ctrl.Result, err error) {
	return r.Create(ctx, resource, data, meta)
}

// Delete permit to delete ILM from Elasticsearch
func (r *ElasticsearchILMReconciler) Delete(ctx context.Context, resource resource.Resource, data map[string]interface{}, meta interface{}) (err error) {
	esHandler := meta.(elasticsearchhandler.ElasticsearchHandler)
	ilm := resource.(*elkv1alpha1.ElasticsearchILM)

	if err = esHandler.ILMDelete(ilm.Name); err != nil {
		return err
	}

	return nil

}

// Diff permit to check if diff between actual and expected ILM exist
func (r *ElasticsearchILMReconciler) Diff(resource resource.Resource, data map[string]interface{}, meta interface{}) (diff controller.Diff, err error) {
	esHandler := meta.(elasticsearchhandler.ElasticsearchHandler)
	ilm := resource.(*elkv1alpha1.ElasticsearchILM)
	expectedPolicy := make(map[string]any)
	var currentPolicy map[string]any
	var d any

	d, err = helper.Get(data, "policy")
	if err != nil {
		return diff, err
	}
	currentPolicy = d.(map[string]any)

	diff = controller.Diff{
		NeedCreate: false,
		NeedUpdate: false,
	}

	if currentPolicy == nil {
		diff.NeedCreate = true
		diff.Diff = "ILM policy not exist"
		return diff, nil
	}

	diffStr, err := esHandler.ILMDiff(currentPolicy, expectedPolicy)
	if err != nil {
		return diff, err
	}

	if diffStr != "" {
		diff.NeedUpdate = true
		diff.Diff = diffStr
		return diff, nil
	}

	// Update condition status if needed
	if condition.IsStatusConditionPresentAndEqual(ilm.Status.Conditions, ilmCondition, v1.ConditionFalse) {
		condition.SetStatusCondition(&ilm.Status.Conditions, v1.Condition{
			Type:   ilmCondition,
			Reason: "Success",
			Status: v1.ConditionTrue,
		})

		r.recorder.Event(resource, core.EventTypeNormal, "Completed", "ILM already set")
	}

	return
}
