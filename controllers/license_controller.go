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
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"

	elkv1alpha1 "github.com/disaster37/operator-elk-extra/api/v1alpha1"
	"github.com/disaster37/operator-elk-extra/pkg/elasticsearchhandler"
	"github.com/disaster37/operator-sdk-extra/pkg/controller"
	"github.com/disaster37/operator-sdk-extra/pkg/helper"
	"github.com/disaster37/operator-sdk-extra/pkg/resource"
	olivere "github.com/olivere/elastic/v7"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	core "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	condition "k8s.io/apimachinery/pkg/api/meta"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	licenseFinalizer  = "license.elk.k8s.webcenter.fr/finalizer"
	licenseAnnotation = "elk.k8s.webcenter.fr/license"
	licenseCondition  = "UpdateLicense"
)

// LicenseReconciler reconciles a License object
type LicenseReconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	recorder   record.EventRecorder
	log        *logrus.Entry
	reconciler controller.Reconciler
}

//+kubebuilder:rbac:groups=elk.k8s.webcenter.fr,resources=licenses,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=elk.k8s.webcenter.fr,resources=licenses/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=elk.k8s.webcenter.fr,resources=licenses/finalizers,verbs=update
//+kubebuilder:rbac:groups=,resources=secrets,verbs=get;list;watch;update;patch

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

	reconciler, err := controller.NewStdReconciler(r.Client, licenseFinalizer, r.reconciler, r.log, r.recorder, waitDurationWhenError)
	if err != nil {
		return ctrl.Result{}, err
	}

	license := &elkv1alpha1.License{}
	data := map[string]any{}

	return reconciler.Reconcile(ctx, req, license, data)
}

// SetupWithManager sets up the controller with the Manager.
func (r *LicenseReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&elkv1alpha1.License{}).
		Complete(r)
}

func (r *LicenseReconciler) SetLogger(log *logrus.Entry) {
	r.log = log
}

func (r *LicenseReconciler) SetRecorder(recorder record.EventRecorder) {
	r.recorder = recorder
}

// Configure permit to Elasticsearch handler
func (r *LicenseReconciler) Configure(ctx context.Context, req ctrl.Request, resource resource.Resource) (meta any, err error) {
	// Get elasticsearch handler / client
	license := resource.(*elkv1alpha1.License)
	meta, err = GetElasticsearchHandler(ctx, &license.Spec, r.Client, req, r.log)
	if err != nil {
		r.recorder.Eventf(resource, core.EventTypeWarning, "Failed", "Unable to init elasticsearch handler: %s", err.Error())
		return nil, err
	}

	return meta, err
}

// Read permit to get current License that fire the reconcile
// It also init elasticsearch handler and read the current license on Elasticsearch
func (r *LicenseReconciler) Read(ctx context.Context, resource resource.Resource, data map[string]any, meta any) (res ctrl.Result, err error) {
	license := resource.(*elkv1alpha1.License)
	esHandler := meta.(elasticsearchhandler.ElasticsearchHandler)

	// Init condition status if not exist
	if condition.FindStatusCondition(license.Status.Conditions, licenseCondition) == nil {
		condition.SetStatusCondition(&license.Status.Conditions, v1.Condition{
			Type:   licenseCondition,
			Status: v1.ConditionFalse,
			Reason: "Initialize",
		})
	}

	// Read license contend from secret if not basic
	if !license.Spec.Basic {
		secret := &core.Secret{}
		secretNS := types.NamespacedName{
			Namespace: license.Namespace,
			Name:      license.Spec.SecretName,
		}
		if err = r.Get(ctx, secretNS, secret); err != nil {
			if k8serrors.IsNotFound(err) {
				r.log.Warnf("Secret %s not yet exist, try later", license.Spec.SecretName)
				r.recorder.Eventf(resource, core.EventTypeWarning, "Failed", "Secret %s not yet exist", license.Spec.SecretName)
				return ctrl.Result{RequeueAfter: waitDurationWhenError}, nil
			}
			r.log.Errorf("Error when get resource: %s", err.Error())
			r.recorder.Eventf(resource, core.EventTypeWarning, "Failed", "Error when get secret %s: %s", license.Spec.SecretName, err.Error())
			return res, err
		}
		licenseB, ok := secret.Data["license"]
		if !ok {
			r.recorder.Eventf(resource, core.EventTypeWarning, "Failed", "Secret %s must have a license key", license.Spec.SecretName)
			return res, errors.Errorf("Secret %s must have a license key", license.Spec.SecretName)
		}
		if err != nil {
			r.recorder.Eventf(resource, core.EventTypeWarning, "Failed", "License contend is invalid: %s", err.Error())
			return res, err
		}
		expectedLicense := &olivere.XPackInfoServiceResponse{}
		if err = json.Unmarshal(licenseB, expectedLicense); err != nil {
			r.recorder.Eventf(resource, core.EventTypeWarning, "Failed", "License contend is invalid: %s", err.Error())
			return res, err
		}
		data["expectedLicense"] = &expectedLicense.License
		data["rawLicense"] = string(licenseB)

		// Add annotation on secret to track change
		if secret.Annotations == nil || secret.Annotations[licenseAnnotation] != license.Name {
			if secret.Annotations == nil {
				secret.Annotations = map[string]string{}
			}
			secret.Annotations[licenseAnnotation] = license.Name
			if err = r.Client.Update(ctx, secret); err != nil {
				r.recorder.Eventf(resource, core.EventTypeWarning, "Failed", "Error when add annotation on secret %s: %s", license.Spec.SecretName, err.Error())
				return res, err
			}

			r.recorder.Eventf(resource, core.EventTypeNormal, "Success", "Add annotation on secret %s", license.Spec.SecretName)
		}
	}

	// Read the current license from Elasticsearch
	licenseInfo, err := esHandler.LicenseGet()
	if err != nil {
		r.recorder.Eventf(resource, core.EventTypeWarning, "Failed", "Unable to get current license from Elasticsearch: %s", err.Error())
		return res, err
	}
	data["currentLicense"] = licenseInfo
	return res, nil
}

// Create add new license or enable basic license
func (r *LicenseReconciler) Create(ctx context.Context, resource resource.Resource, data map[string]interface{}, meta interface{}) (res ctrl.Result, err error) {

	esHandler := meta.(elasticsearchhandler.ElasticsearchHandler)
	license := resource.(*elkv1alpha1.License)
	var d any

	// handler condition status if error
	defer func() {
		if err != nil {
			condition.SetStatusCondition(&license.Status.Conditions, v1.Condition{
				Type:    licenseCondition,
				Status:  v1.ConditionFalse,
				Reason:  "Failed",
				Message: err.Error(),
			})
		}
	}()

	// Basic license
	if license.Spec.Basic {
		if err = esHandler.LicenseEnableBasic(); err != nil {
			r.recorder.Eventf(resource, core.EventTypeWarning, "Failed", "Error when activate basic license: %s", err.Error())
			return res, err
		}
		r.log.Info("Successfully enable basic license")
		r.recorder.Event(resource, core.EventTypeNormal, "Completed", "Enable basic license")
		license.Status.LicenseType = "basic"
		license.Status.ExpireAt = ""
		license.Status.LicenseHash = ""

	} else {
		// Enterprise license
		d, err = helper.Get(data, "expectedLicense")
		if err != nil {
			return res, err
		}
		expectedLicense := d.(*olivere.XPackInfoLicense)
		d, err = helper.Get(data, "rawLicense")
		if err != nil {
			return res, err
		}
		rawLicense := d.(string)
		if err = esHandler.LicenseUpdate(rawLicense); err != nil {
			r.recorder.Eventf(resource, core.EventTypeWarning, "Failed", "Error when add enterprise license: %s", err.Error())
			return res, err
		}
		r.log.Infof("Successfully enable %s license", expectedLicense.Type)
		r.recorder.Eventf(resource, core.EventTypeNormal, "Completed", "Enable %s license", expectedLicense.Type)

		license.Status.ExpireAt = time.UnixMilli(int64(expectedLicense.ExpiryMilis)).Format(time.RFC3339)
		license.Status.LicenseHash = fmt.Sprintf("%x", sha256.Sum256([]byte(rawLicense)))
		license.Status.LicenseType = expectedLicense.Type
	}

	// Update status
	condition.SetStatusCondition(&license.Status.Conditions, v1.Condition{
		Type:    licenseCondition,
		Status:  v1.ConditionTrue,
		Reason:  "Success",
		Message: fmt.Sprintf("License of type %s successfully updated", license.Status.LicenseType),
	})

	return res, nil
}

// Update permit to update current license from Elasticsearch
func (r *LicenseReconciler) Update(ctx context.Context, resource resource.Resource, data map[string]interface{}, meta interface{}) (res ctrl.Result, err error) {
	return r.Create(ctx, resource, data, meta)
}

// Delete permit to delete current license from Elasticsearch
func (r *LicenseReconciler) Delete(ctx context.Context, resource resource.Resource, data map[string]interface{}, meta interface{}) (err error) {
	esHandler := meta.(elasticsearchhandler.ElasticsearchHandler)
	license := resource.(*elkv1alpha1.License)

	// Not delete License
	// If enterprise license, it must enable basic license instead
	if !license.Spec.Basic {
		if err = esHandler.LicenseEnableBasic(); err != nil {
			r.recorder.Eventf(resource, core.EventTypeWarning, "Failed", "Error when downgrade to basic license: %s", err.Error())
			return err
		}
		r.log.Info("Successfully downgrade to basic license")
		r.recorder.Event(resource, core.EventTypeNormal, "Completed", "Downgrade to basic license")
	}

	return nil

}

// Diff permit to check if diff between actual and expected license exist
func (r *LicenseReconciler) Diff(resource resource.Resource, data map[string]interface{}, meta interface{}) (diff controller.Diff, err error) {
	esHandler := meta.(elasticsearchhandler.ElasticsearchHandler)
	license := resource.(*elkv1alpha1.License)

	var expectedLicense *olivere.XPackInfoLicense
	var d any

	if license.Spec.Basic {
		expectedLicense = &olivere.XPackInfoLicense{
			Type: "basic",
		}
	} else {
		d, err = helper.Get(data, "expectedLicense")
		if err != nil {
			return diff, err
		}
		expectedLicense = d.(*olivere.XPackInfoLicense)

	}

	d, err = helper.Get(data, "currentLicense")
	if err != nil {
		return diff, err
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

	// Update condition status if needed
	if condition.IsStatusConditionPresentAndEqual(license.Status.Conditions, licenseCondition, v1.ConditionFalse) {
		condition.SetStatusCondition(&license.Status.Conditions, v1.Condition{
			Type:   licenseCondition,
			Reason: "Success",
			Status: v1.ConditionTrue,
		})

		r.recorder.Event(resource, core.EventTypeNormal, "Completed", "License already set")
	}

	return
}
