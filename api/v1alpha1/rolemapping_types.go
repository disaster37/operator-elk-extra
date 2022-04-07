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
	"encoding/json"

	olivere "github.com/olivere/elastic/v7"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// RoleMappingSpec defines the desired state of RoleMapping
// +k8s:openapi-gen=true
type RoleMappingSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	ElasticsearchRefSpec `json:"elasticsearchRef"`

	// Enabled permit to enable or disable the role mapping
	Enabled bool `json:"enabled"`

	// Roles is the list of role to map
	Roles []string `json:"roles"`

	// Rules is the mapping rules
	// JSON string
	Rules string `json:"rules"`

	// Metadata is the meta data
	// JSON string
	// +optional
	Metadata string `json:"metadata"`
}

// RoleMappingStatus defines the observed state of RoleMapping
type RoleMappingStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Conditions []metav1.Condition `json:"conditions"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// RoleMapping is the Schema for the rolemappings API
type RoleMapping struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RoleMappingSpec   `json:"spec,omitempty"`
	Status RoleMappingStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// RoleMappingList contains a list of RoleMapping
type RoleMappingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RoleMapping `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RoleMapping{}, &RoleMappingList{})
}

// GetObjectMeta permit to get the current ObjectMeta
func (h *RoleMapping) GetObjectMeta() metav1.ObjectMeta {
	return h.ObjectMeta
}

// GetStatus permit to get the current status
func (h *RoleMapping) GetStatus() any {
	return h.Status
}

func (h *RoleMapping) ToRoleMapping() (*olivere.XPackSecurityRoleMapping, error) {
	rm := &olivere.XPackSecurityRoleMapping{
		Enabled: h.Spec.Enabled,
		Roles:   h.Spec.Roles,
	}

	if h.Spec.Rules != "" {
		rules := make(map[string]any)
		if err := json.Unmarshal([]byte(h.Spec.Rules), &rules); err != nil {
			return nil, err
		}
		rm.Rules = rules
	}

	if h.Spec.Metadata != "" {
		meta := make(map[string]any)
		if err := json.Unmarshal([]byte(h.Spec.Metadata), &meta); err != nil {
			return nil, err
		}
		rm.Metadata = meta
	}

	return rm, nil
}
