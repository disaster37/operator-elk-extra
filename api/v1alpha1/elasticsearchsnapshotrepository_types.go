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
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ElasticsearchSnapshotRepositorySpec defines the desired state of ElasticsearchSnapshotRepository
// +k8s:openapi-gen=true
type ElasticsearchSnapshotRepositorySpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	ElasticsearchRefSpec `json:"elasticsearchRef"`

	// Type the Snapshot repository type
	Type string `json:"type"`

	// The config of snapshot repository
	// +optional
	Settings string `json:"settings,omitempty"`
}

// ElasticsearchSnapshotRepositoryStatus defines the observed state of ElasticsearchSnapshotRepository
type ElasticsearchSnapshotRepositoryStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Conditions []metav1.Condition `json:"conditions"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ElasticsearchSnapshotRepository is the Schema for the elasticsearchsnapshotrepositories API
type ElasticsearchSnapshotRepository struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ElasticsearchSnapshotRepositorySpec   `json:"spec,omitempty"`
	Status ElasticsearchSnapshotRepositoryStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ElasticsearchSnapshotRepositoryList contains a list of ElasticsearchSnapshotRepository
type ElasticsearchSnapshotRepositoryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ElasticsearchSnapshotRepository `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ElasticsearchSnapshotRepository{}, &ElasticsearchSnapshotRepositoryList{})
}

// GetObjectMeta permit to get the current ObjectMeta
func (h *ElasticsearchSnapshotRepository) GetObjectMeta() metav1.ObjectMeta {
	return h.ObjectMeta
}

// GetStatus permit to get the current status
func (h *ElasticsearchSnapshotRepository) GetStatus() any {
	return h.Status
}
