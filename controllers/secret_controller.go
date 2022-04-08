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
	"fmt"

	elkv1alpha1 "github.com/disaster37/operator-elk-extra/api/v1alpha1"
	"github.com/sirupsen/logrus"
	core "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

type Secret struct {
	core.Secret
}

// GetObjectMeta permit to get the current ObjectMeta
func (h *Secret) GetObjectMeta() metav1.ObjectMeta {
	return h.ObjectMeta
}

// GetStatus permit to get the current status
func (h *Secret) GetStatus() any {
	return nil
}

// SecretReconciler reconciles a Secret object
type SecretReconciler struct {
	Reconciler
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups="",resources=secrets/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=elk.k8s.webcenter.fr,resources=licenses,verbs=get;list;watch
//+kubebuilder:rbac:groups=elk.k8s.webcenter.fr,resources=licenses/status,verbs=get;update;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Secret object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *SecretReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	r.log = r.log.WithFields(logrus.Fields{
		"name":      req.Name,
		"namespace": req.Namespace,
	})
	r.log.Infof("---> Starting reconcile loop")
	defer r.log.Info("---> Finish reconcile loop for")

	// Get current resource
	secret := &core.Secret{}
	if err := r.Get(ctx, req.NamespacedName, secret); err != nil {
		if k8serrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
	}

	// Get license object
	licenseName := secret.Annotations[licenseAnnotation]
	if licenseName == "" {
		return ctrl.Result{}, nil
	}
	license := &elkv1alpha1.License{}
	nsLicense := types.NamespacedName{
		Namespace: req.Namespace,
		Name:      licenseName,
	}
	if err := r.Get(ctx, nsLicense, license); err != nil {
		if k8serrors.IsNotFound(err) {
			r.log.Infof("License %s not found, skip it", licenseName)
			return ctrl.Result{}, nil
		}
	}

	// Update hash status on license if diff
	licenseB, ok := secret.Data["license"]
	if !ok {
		r.log.Warnf("Secret %s not contain license key, skip it", req.Name)
		return ctrl.Result{}, nil
	}
	hashLicense := fmt.Sprintf("%x", sha256.Sum256(licenseB))
	if hashLicense != license.Status.LicenseHash {
		license.Status.LicenseHash = hashLicense
		if err := r.Client.Status().Update(ctx, license); err != nil {
			r.log.Errorf("Error when update license status: %s", err.Error())
			return ctrl.Result{Requeue: true}, nil
		}
		r.log.Info("Change license hash successfully to force reconcile")
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SecretReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&core.Secret{}).
		WithEventFilter(viewSecretAnnotationPredicate()).
		Complete(r)
}

func viewSecretAnnotationPredicate() predicate.Predicate {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {

			return isLicenseAnnotation(e.ObjectNew.GetAnnotations())
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return isLicenseAnnotation(e.Object.GetAnnotations())
		},
		CreateFunc: func(e event.CreateEvent) bool {
			return isLicenseAnnotation(e.Object.GetAnnotations())
		},
		GenericFunc: func(e event.GenericEvent) bool {
			return isLicenseAnnotation(e.Object.GetAnnotations())
		},
	}
}

func isLicenseAnnotation(annotations map[string]string) bool {
	if annotations == nil {
		return false
	}
	_, ok := annotations[licenseAnnotation]
	if !ok {
		return false
	}
	return true
}
