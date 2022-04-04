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

// ElasticsearchComponentTemplateSpec defines the desired state of ElasticsearchComponentTemplate
// +k8s:openapi-gen=true
type ElasticsearchComponentTemplateSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	ElasticsearchRefSpec `json:"elasticsearchRef,omitempty"`

	// Settings is the component setting
	Settings string `json:"settings,omitempty"`

	// Mappings is the component mapping
	Mappings string `json:"mappings,omitempty"`

	// Aliases is the component aliases
	Aliases string `json:"aliases,omitempty"`
}

// ElasticsearchComponentTemplateStatus defines the observed state of ElasticsearchComponentTemplate
type ElasticsearchComponentTemplateStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Conditions []metav1.Condition `json:"conditions"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ElasticsearchComponentTemplate is the Schema for the elasticsearchcomponenttemplates API
type ElasticsearchComponentTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ElasticsearchComponentTemplateSpec   `json:"spec,omitempty"`
	Status ElasticsearchComponentTemplateStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ElasticsearchComponentTemplateList contains a list of ElasticsearchComponentTemplate
type ElasticsearchComponentTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ElasticsearchComponentTemplate `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ElasticsearchComponentTemplate{}, &ElasticsearchComponentTemplateList{})
}

// GetObjectMeta permit to get the current ObjectMeta
func (h *ElasticsearchComponentTemplate) GetObjectMeta() metav1.ObjectMeta {
	return h.ObjectMeta
}

// GetStatus permit to get the current status
func (h *ElasticsearchComponentTemplate) GetStatus() any {
	return h.Status
}

// ToComponentTemplate permit to convert current spec to component template spec
func (h *ElasticsearchComponentTemplate) ToComponentTemplate() (*olivere.IndicesGetComponentTemplateData, error) {
	component := &olivere.IndicesGetComponentTemplateData{
		Settings: make(map[string]any),
		Mappings: make(map[string]any),
		Aliases:  make(map[string]any),
	}

	if h.Spec.Mappings != "" {
		if err := json.Unmarshal([]byte(h.Spec.Mappings), &component.Mappings); err != nil {
			return nil, err
		}
	}

	if h.Spec.Settings != "" {
		if err := json.Unmarshal([]byte(h.Spec.Settings), &component.Settings); err != nil {
			return nil, err
		}
	}

	if h.Spec.Aliases != "" {
		if err := json.Unmarshal([]byte(h.Spec.Aliases), &component.Aliases); err != nil {
			return nil, err
		}
	}

	return component, nil
}
