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

// ElasticsearchIndexTemplateSpec defines the desired state of ElasticsearchIndexTemplate
// +k8s:openapi-gen=true
type ElasticsearchIndexTemplateSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	ElasticsearchRefSpec `json:"elasticsearchRef"`

	// IndexPatterns is the list of index to apply this template
	IndexPatterns []string `json:"index_patterns,omitempty"`

	//ComposedOf is the list of component templates
	// +optional
	ComposedOf []string `json:"composed_of,omitempty"`

	//Priority is the priority to apply this template
	// +optional
	Priority int `json:"priority,omitempty"`

	// The version
	// +optional
	Version int `json:"version,omitempty"`

	// Template is the template specification
	// +optional
	Template *ElasticsearchIndexTemplateData `json:"template,omitempty"`

	// Meta is extended info as JSON string
	// +optional
	Meta string `json:"_meta,omitempty"`

	// AllowAutoCreate permit to allow auto create index
	// +optional
	AllowAutoCreate bool `json:"allow_auto_create,omitempty"`
}

// ElasticsearchIndexTemplateData is the template specification
type ElasticsearchIndexTemplateData struct {

	// Settings is the template setting as JSON string
	// +optional
	Settings string `json:"settings,omitempty"`

	// Mappings is the template mapping as JSON string
	// +optional
	Mappings string `json:"mappings,omitempty"`

	// Aliases is the template alias as JSON string
	// +optional
	Aliases string `json:"aliases,omitempty"`
}

// ElasticsearchIndexTemplateStatus defines the observed state of ElasticsearchIndexTemplate
type ElasticsearchIndexTemplateStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Conditions []metav1.Condition `json:"conditions"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ElasticsearchIndexTemplate is the Schema for the elasticsearchindextemplates API
type ElasticsearchIndexTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ElasticsearchIndexTemplateSpec   `json:"spec,omitempty"`
	Status ElasticsearchIndexTemplateStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ElasticsearchIndexTemplateList contains a list of ElasticsearchIndexTemplate
type ElasticsearchIndexTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ElasticsearchIndexTemplate `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ElasticsearchIndexTemplate{}, &ElasticsearchIndexTemplateList{})
}

// GetObjectMeta permit to get the current ObjectMeta
func (h *ElasticsearchIndexTemplate) GetObjectMeta() metav1.ObjectMeta {
	return h.ObjectMeta
}

// GetStatus permit to get the current status
func (h *ElasticsearchIndexTemplate) GetStatus() any {
	return h.Status
}

func (h *ElasticsearchIndexTemplate) ToIndexTemplate() (*olivere.IndicesGetIndexTemplate, error) {
	template := &olivere.IndicesGetIndexTemplate{
		IndexPatterns:   h.Spec.IndexPatterns,
		ComposedOf:      h.Spec.ComposedOf,
		Priority:        h.Spec.Priority,
		Version:         h.Spec.Version,
		AllowAutoCreate: h.Spec.AllowAutoCreate,
	}

	if h.Spec.Template != nil {
		var settings, mappings, aliases map[string]any
		if h.Spec.Template.Settings != "" {
			settings = make(map[string]any)
			if err := json.Unmarshal([]byte(h.Spec.Template.Settings), &settings); err != nil {
				return nil, err
			}
		}
		if h.Spec.Template.Mappings != "" {
			mappings = make(map[string]any)
			if err := json.Unmarshal([]byte(h.Spec.Template.Mappings), &mappings); err != nil {
				return nil, err
			}
		}
		if h.Spec.Template.Aliases != "" {
			aliases = make(map[string]any)
			if err := json.Unmarshal([]byte(h.Spec.Template.Aliases), &aliases); err != nil {
				return nil, err
			}
		}
		template.Template = &olivere.IndicesGetIndexTemplateData{
			Settings: settings,
			Mappings: mappings,
			Aliases:  aliases,
		}
	}

	if h.Spec.Meta != "" {
		meta := make(map[string]any)
		if err := json.Unmarshal([]byte(h.Spec.Meta), &meta); err != nil {
			return nil, err
		}
		template.Meta = meta
	}

	return template, nil
}
