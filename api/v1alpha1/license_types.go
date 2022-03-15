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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const licenseFinalizer = "license.elk.k8s.webcenter.fr/finalizer"

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// LicenseSpec defines the desired state of License
// +k8s:openapi-gen=true
type LicenseSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// SecretName is the secret that contain the license
	SecretName string `json:"secretName,omitempty"`

	// ElasticsearchRef is the Elasticsearch reference to connect on it (managed by ECK)
	// +optional
	ElasticsearchRef *ElasticsearchRefSpec `json:"elasticsearchRef,omitempty"`

	//ElasticsearchExternalRef is the elasticsearch external reference to connect on it (not managed by ECK)
	// +optional
	ElasticsearchExternalRef *ElasticsearchExternalRefSpec `json:"elasticsearchExternalRef,omitempty"`
}

// LicenseStatus defines the observed state of License
type LicenseStatus struct {
	LicenseType string             `json:"licenseType,omitempty"`
	ExpireAt    string             `json:"expireAt,omitempty"`
	Conditions  []metav1.Condition `json:"conditions"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// License is the Schema for the licenses API
type License struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LicenseSpec   `json:"spec,omitempty"`
	Status LicenseStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// LicenseList contains a list of License
type LicenseList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []License `json:"items"`
}

func init() {
	SchemeBuilder.Register(&License{}, &LicenseList{})
}

// IsBeingDeleted returns true if a deletion timestamp is set
func (h *License) IsBeingDeleted() bool {
	return !h.ObjectMeta.DeletionTimestamp.IsZero()
}

// HasFinalizer returns true if the item has the specified finalizer
func (h *License) HasFinalizer() bool {
	return controllerutil.ContainsFinalizer(h, licenseFinalizer)
}

// IsSubmitted return true if service has been submitted to Centreon
func (h *License) IsSubmitted() bool {
	return h.Status.LicenseType != ""
}

// AddFinalizer adds the specified finalizer
func (h *License) AddFinalizer() {
	controllerutil.AddFinalizer(h, licenseFinalizer)
}

// RemoveFinalizer removes the specified finalizer
func (h *License) RemoveFinalizer() {
	controllerutil.RemoveFinalizer(h, licenseFinalizer)
}
