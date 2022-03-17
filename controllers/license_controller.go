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
	"encoding/base64"
	"encoding/json"

	elkv1alpha1 "github.com/disaster37/operator-elk-extra/api/v1alpha1"
	"github.com/disaster37/operator-elk-extra/pkg/elasticsearchhandler"
	"github.com/disaster37/operator-sdk-extra/pkg/controller"
	"github.com/disaster37/operator-sdk-extra/pkg/resource"
	olivere "github.com/olivere/elastic/v7"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	core "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	licenseFinalizer = "license.elk.k8s.webcenter.fr/finalizer"
	licenseBasic     = "basic"
)

// LicenseReconciler reconciles a License object
type LicenseReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	recorder record.EventRecorder
	log      *logrus.Entry
}

//+kubebuilder:rbac:groups=elk.k8s.webcenter.fr,resources=licenses,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=elk.k8s.webcenter.fr,resources=licenses/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=elk.k8s.webcenter.fr,resources=licenses/finalizers,verbs=update
//+kubebuilder:rbac:groups=,resources=secrets,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the License object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *LicenseReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	reconciler, err := controller.NewStdReconciler(r.Client, licenseFinalizer, r, r.log, r.recorder, waitDurationWhenError)
	if err != nil {
		return ctrl.Result{}, err
	}

	license := &elkv1alpha1.License{}
	data := map[string]any{}
	meta := &elasticsearchhandler.ElasticsearchHandlerImpl{}

	return reconciler.Reconcile(ctx, req, license, data, meta)
}

// SetupWithManager sets up the controller with the Manager.
func (r *LicenseReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&elkv1alpha1.License{}).
		For(&core.Secret{}).
		Complete(r)
}

func (r *LicenseReconciler) SetLogger(log *logrus.Entry) {
	r.log = log
}

func (r *LicenseReconciler) SetRecorder(recorder record.EventRecorder) {
	r.recorder = recorder
}

// Read permit to get current License that fire the reconcile
// It also init elasticsearch handler and read the current license on Elasticsearch
func (r *LicenseReconciler) Read(ctx context.Context, req ctrl.Request, resource resource.Resource, data map[string]interface{}, meta interface{}) (res *ctrl.Result, err error) {
	// Get license
	license := resource.(*elkv1alpha1.License)
	if err := r.Get(ctx, req.NamespacedName, license); err != nil {
		if k8serrors.IsNotFound(err) {
			return &ctrl.Result{}, nil
		}
		r.log.Errorf("Error when get license: %s", err.Error())
		return nil, err
	}

	// Read license contend from secret
	secret := &core.Secret{}
	secretNS := types.NamespacedName{
		Namespace: req.NamespacedName.Namespace,
		Name:      license.Spec.SecretName,
	}
	if err = r.Get(ctx, secretNS, secret); err != nil {
		if k8serrors.IsNotFound(err) {
			r.log.Warnf("Secret %s not yet exist, try later", license.Spec.SecretName)
			r.recorder.Eventf(resource, core.EventTypeWarning, "Failed", "Secret %s not yet exist", license.Spec.SecretName)
			return &ctrl.Result{RequeueAfter: waitDurationWhenError}, nil
		}
		r.log.Errorf("Error when get resource: %s", err.Error())
		r.recorder.Eventf(resource, core.EventTypeWarning, "Failed", "Error when get secret %s: %s", license.Spec.SecretName, err.Error())
		return nil, err
	}
	licenseb64, ok := secret.Data["license"]
	if !ok {
		r.recorder.Eventf(resource, core.EventTypeWarning, "Failed", "Secret %s must have a license key", license.Spec.SecretName)
		return nil, errors.Errorf("Secret %s must have a license key", license.Spec.SecretName)
	}
	licenseData, err := base64.RawStdEncoding.DecodeString(string(licenseb64))
	if err != nil {
		r.recorder.Eventf(resource, core.EventTypeWarning, "Failed", "License contend is invalid: %s", err.Error())
		return nil, err
	}
	expectedLicense := &olivere.XPackInfoLicense{}
	if err = json.Unmarshal(licenseData, expectedLicense); err != nil {
		r.recorder.Eventf(resource, core.EventTypeWarning, "Failed", "License contend is invalid: %s", err.Error())
		return nil, err
	}
	data["expectedLicense"] = expectedLicense
	data["rawLicense"] = string(licenseData)

	// Get elasticsearch handler / client
	esHandler, res, err := GetElasticsearchHandler(ctx, &license.Spec, r.Client, req, r.log)
	if err != nil {
		r.recorder.Eventf(resource, core.EventTypeWarning, "Failed", "Unable to init elasticsearch handler: %s", err.Error())
		return nil, err
	}
	if res != nil {
		return res, nil
	}
	m := meta.(*elasticsearchhandler.ElasticsearchHandlerImpl)
	esHandlerP := esHandler.(*elasticsearchhandler.ElasticsearchHandlerImpl)
	*m = *esHandlerP

	// Read the current license from Elasticsearch
	licenseInfo, err := esHandler.LicenseGet()
	if err != nil {
		r.recorder.Eventf(resource, core.EventTypeWarning, "Failed", "Unable to get current license from Elasticsearch: %s", err.Error())
		return nil, err
	}
	data["currentLicense"] = licenseInfo
	return nil, nil
}

// Create add new license or enable basic license
func (r *LicenseReconciler) Create(ctx context.Context, resource resource.Resource, data map[string]interface{}, meta interface{}) (res ctrl.Result, err error) {

	esHandler := meta.(elasticsearchhandler.ElasticsearchHandler)
	d, ok := data["expectedLicense"]
	if !ok {
		return res, errors.New("expectedLicense not provided")
	}
	expectedLicense := d.(*olivere.XPackInfoLicense)

	// Basic license
	if expectedLicense.Type == licenseBasic {
		if err = esHandler.LicenseEnableBasic(); err != nil {
			r.recorder.Eventf(resource, core.EventTypeWarning, "Failed", "Error when activate basic license: %s", err.Error())
			return res, err
		}
		r.log.Info("Successfully enable basic license")
		r.recorder.Event(resource, core.EventTypeNormal, "Completed", "Enable basic license")
		return
	}

	// Enterprise license
	d, ok = data["rawLicense"]
	if !ok {
		return res, errors.New("rawLicense not provided")
	}
	rawLicense := d.(string)
	if err = esHandler.LicenseUpdate(rawLicense); err != nil {
		r.recorder.Eventf(resource, core.EventTypeWarning, "Failed", "Error when add enterprise license: %s", err.Error())
		return res, err
	}
	r.log.Infof("Successfully enable %s license", expectedLicense.Type)
	r.recorder.Eventf(resource, core.EventTypeNormal, "Completed", "Enable %s license", expectedLicense.Type)
	return
}

// Update permit to update current license from Elasticsearch
func (r *LicenseReconciler) Update(ctx context.Context, resource resource.Resource, data map[string]interface{}, meta interface{}) (res ctrl.Result, err error) {
	return r.Create(ctx, resource, data, meta)
}

// Delete permit to delete current license from Elasticsearch
func (r *LicenseReconciler) Delete(ctx context.Context, resource resource.Resource, data map[string]interface{}, meta interface{}) (err error) {
	esHandler := meta.(elasticsearchhandler.ElasticsearchHandler)
	if err = esHandler.LicenseDelete(); err != nil {
		r.recorder.Eventf(resource, core.EventTypeWarning, "Failed", "Error when delete license: %s", err.Error())
		return err
	}

	r.log.Info("Successfully delete license")
	r.recorder.Event(resource, core.EventTypeNormal, "Completed", "Delete license")
	return
}

// Diff permit to check if diff between actual and expected license exist
func (r *LicenseReconciler) Diff(resource resource.Resource, data map[string]interface{}, meta interface{}) (diff controller.Diff, err error) {
	esHandler := meta.(elasticsearchhandler.ElasticsearchHandler)

	d, ok := data["expectedLicense"]
	if !ok {
		return diff, errors.New("expectedLicense not provided")
	}
	expectedLicense := d.(*olivere.XPackInfoLicense)

	d, ok = data["currentLicense"]
	if !ok {
		return diff, errors.New("currentLicense not provided")
	}
	currentLicense := d.(*olivere.XPackInfoLicense)

	diff = controller.Diff{
		NeedCreate: false,
		NeedUpdate: false,
	}
	if currentLicense == nil {
		diff.NeedCreate = true
		diff.Diff = "UID or license type mismatch"
		return diff, nil
	}

	if esHandler.LicenseDiff(currentLicense, expectedLicense) {
		diff.NeedUpdate = true
		return diff, nil
	}

	return
}
