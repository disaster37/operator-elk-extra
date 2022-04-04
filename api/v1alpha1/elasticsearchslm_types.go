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
	"github.com/disaster37/operator-elk-extra/pkg/elasticsearchhandler"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ElasticsearchSLMSpec defines the desired state of ElasticsearchSLM
// +k8s:openapi-gen=true
type ElasticsearchSLMSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	ElasticsearchRefSpec `json:"elasticsearchRef"`

	Schedule   string                     `json:"schedule"`
	Name       string                     `json:"name"`
	Repository string                     `json:"repository"`
	Config     ElasticsearchSLMConfig     `json:"config"`
	Retention  *ElasticsearchSLMRetention `json:"retention,omitempty"`
}

// ElasticsearchSLMConfig is the config sub section
type ElasticsearchSLMConfig struct {
	ExpendWildcards    string            `json:"expand_wildcards,omitempty"`
	IgnoreUnavailable  bool              `json:"ignore_unavailable,omitempty"`
	IncludeGlobalState bool              `json:"include_global_state,omitempty"`
	Indices            []string          `json:"indices,omitempty"`
	FeatureStates      []string          `json:"feature_states,omitempty"`
	Metadata           map[string]string `json:"metadata,omitempty"`
	Partial            bool              `json:"partial,omitempty"`
}

// ElasticsearchSLMRetention is the retention sub section
type ElasticsearchSLMRetention struct {
	ExpireAfter string `json:"expire_after,omitempty"`
	MaxCount    int64  `json:"max_count,omitempty"`
	MinCount    int64  `json:"min_count,omitempty"`
}

// ElasticsearchSLMStatus defines the observed state of ElasticsearchSLM
type ElasticsearchSLMStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Conditions []metav1.Condition `json:"conditions"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ElasticsearchSLM is the Schema for the elasticsearchslms API
type ElasticsearchSLM struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ElasticsearchSLMSpec   `json:"spec,omitempty"`
	Status ElasticsearchSLMStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ElasticsearchSLMList contains a list of ElasticsearchSLM
type ElasticsearchSLMList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ElasticsearchSLM `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ElasticsearchSLM{}, &ElasticsearchSLMList{})
}

// GetObjectMeta permit to get the current ObjectMeta
func (h *ElasticsearchSLM) GetObjectMeta() metav1.ObjectMeta {
	return h.ObjectMeta
}

// GetStatus permit to get the current status
func (h *ElasticsearchSLM) GetStatus() any {
	return h.Status
}

func (h *ElasticsearchSLM) ToPolicy() *elasticsearchhandler.SnapshotLifecyclePolicySpec {
	policy := &elasticsearchhandler.SnapshotLifecyclePolicySpec{
		Schedule:   h.Spec.Schedule,
		Name:       h.Spec.Name,
		Repository: h.Spec.Repository,
		Config: elasticsearchhandler.ElasticsearchSLMConfig{
			ExpendWildcards:    h.Spec.Config.ExpendWildcards,
			IgnoreUnavailable:  h.Spec.Config.IgnoreUnavailable,
			IncludeGlobalState: h.Spec.Config.IncludeGlobalState,
			Indices:            h.Spec.Config.Indices,
			FeatureStates:      h.Spec.Config.FeatureStates,
			Metadata:           h.Spec.Config.Metadata,
			Partial:            h.Spec.Config.Partial,
		},
	}

	if h.Spec.Retention != nil {
		policy.Retention = &elasticsearchhandler.ElasticsearchSLMRetention{
			ExpireAfter: h.Spec.Retention.ExpireAfter,
			MaxCount:    h.Spec.Retention.MaxCount,
			MinCount:    h.Spec.Retention.MinCount,
		}
	}

	return policy
}
