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

// ElasticsearchWatcherSpec defines the desired state of ElasticsearchWatcher
// +k8s:openapi-gen=true
type ElasticsearchWatcherSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	ElasticsearchRefSpec `json:"elasticsearchRef"`

	// JSON string
	Trigger string `json:"trigger"`

	// JSON string
	Input string `json:"input"`

	// JSON string
	Condition string `json:"condition"`

	// JSON string
	// +optional
	Transform string `json:"transform,omitempty"`

	// +optional
	ThrottlePeriod string `json:"throttle_period,omitempty"`

	// +optional
	ThrottlePeriodInMillis int64 `json:"throttle_period_in_millis,omitempty"`

	// JSON string
	Actions string `json:"actions"`

	// JSON string
	// +optional
	Metadata string `json:"metadata,omitempty"`
}

// ElasticsearchWatcherStatus defines the observed state of ElasticsearchWatcher
type ElasticsearchWatcherStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Conditions []metav1.Condition `json:"conditions"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ElasticsearchWatcher is the Schema for the elasticsearchwatchers API
type ElasticsearchWatcher struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ElasticsearchWatcherSpec   `json:"spec,omitempty"`
	Status ElasticsearchWatcherStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ElasticsearchWatcherList contains a list of ElasticsearchWatcher
type ElasticsearchWatcherList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ElasticsearchWatcher `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ElasticsearchWatcher{}, &ElasticsearchWatcherList{})
}

// GetObjectMeta permit to get the current ObjectMeta
func (h *ElasticsearchWatcher) GetObjectMeta() metav1.ObjectMeta {
	return h.ObjectMeta
}

// GetStatus permit to get the current status
func (h *ElasticsearchWatcher) GetStatus() any {
	return h.Status
}

func (h *ElasticsearchWatcher) ToWatch() (*olivere.XPackWatch, error) {
	watch := &olivere.XPackWatch{
		ThrottlePeriod:         h.Spec.ThrottlePeriod,
		ThrottlePeriodInMillis: h.Spec.ThrottlePeriodInMillis,
	}

	if h.Spec.Trigger != "" {
		trigger := make(map[string]map[string]any)
		if err := json.Unmarshal([]byte(h.Spec.Trigger), &trigger); err != nil {
			watch.Trigger = trigger
		}
	}

	if h.Spec.Input != "" {
		input := make(map[string]map[string]any)
		if err := json.Unmarshal([]byte(h.Spec.Input), &input); err != nil {
			watch.Input = input
		}
	}

	if h.Spec.Condition != "" {
		condition := make(map[string]map[string]any)
		if err := json.Unmarshal([]byte(h.Spec.Condition), &condition); err != nil {
			watch.Condition = condition
		}
	}

	if h.Spec.Transform != "" {
		transform := make(map[string]any)
		if err := json.Unmarshal([]byte(h.Spec.Transform), &transform); err != nil {
			watch.Transform = transform
		}
	}

	if h.Spec.Actions != "" {
		actions := make(map[string]map[string]any)
		if err := json.Unmarshal([]byte(h.Spec.Actions), &actions); err != nil {
			watch.Actions = actions
		}
	}

	if h.Spec.Metadata != "" {
		meta := make(map[string]any)
		if err := json.Unmarshal([]byte(h.Spec.Metadata), &meta); err != nil {
			watch.Metadata = meta
		}
	}

	return watch, nil
}
