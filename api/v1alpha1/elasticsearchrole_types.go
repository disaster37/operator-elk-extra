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

	"github.com/disaster37/operator-elk-extra/pkg/elasticsearchhandler"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ElasticsearchRoleSpec defines the desired state of ElasticsearchRole
// +k8s:openapi-gen=true
type ElasticsearchRoleSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	ElasticsearchRefSpec `json:"elasticsearchRef"`

	// Cluster is a list of cluster privileges
	// +optional
	Cluster []string `json:"cluster,omitempty"`

	// Indices is the list of indices permissions
	// +optional
	Indices []ElasticsearchRoleSpecIndicesPermissions `json:"indices,omitempty"`

	// Applications is the list of application privilege
	// +optional
	Applications []ElasticsearchRoleSpecApplicationPrivileges `json:"applications,omitempty"`

	// RunAs is the list of users that the owners of this role can impersonate
	// +optional
	RunAs []string `json:"run_as,omitempty"`

	// Global  defining global privileges
	// JSON string
	// +optional
	Global string `json:"global,omitempty"`

	// Metadata is optional meta-data
	// JSON string
	// +optional
	Metadata string `json:"metadata,omitempty"`

	// JSON string
	// +optional
	TransientMetadata string `json:"transient_metadata,omitempty"`
}

// ElasticsearchRoleSpecApplicationPrivileges is the application privileges object
type ElasticsearchRoleSpecApplicationPrivileges struct {
	Application string `json:"application"`

	// +optional
	Privileges []string `json:"privileges,omitempty"`

	// +optional
	Resources []string `json:"resources,omitempty"`
}

// ElasticsearchRoleSpecIndicesPermissions is the indices permission object
type ElasticsearchRoleSpecIndicesPermissions struct {
	Names      []string `json:"names"`
	Privileges []string `json:"privileges"`

	// JSON string
	// +optional
	FieldSecurity string `json:"field_security,omitempty"`

	// +optional
	Query string `json:"query,omitempty"`
}

// ElasticsearchRoleStatus defines the observed state of ElasticsearchRole
type ElasticsearchRoleStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Conditions []metav1.Condition `json:"conditions"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ElasticsearchRole is the Schema for the elasticsearchroles API
type ElasticsearchRole struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ElasticsearchRoleSpec   `json:"spec,omitempty"`
	Status ElasticsearchRoleStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ElasticsearchRoleList contains a list of ElasticsearchRole
type ElasticsearchRoleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ElasticsearchRole `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ElasticsearchRole{}, &ElasticsearchRoleList{})
}

// GetObjectMeta permit to get the current ObjectMeta
func (h *ElasticsearchRole) GetObjectMeta() metav1.ObjectMeta {
	return h.ObjectMeta
}

// GetStatus permit to get the current status
func (h *ElasticsearchRole) GetStatus() any {
	return h.Status
}

// ToComponentTemplate permit to convert current spec to component template spec
func (h *ElasticsearchRole) ToRole() (*elasticsearchhandler.XPackSecurityRole, error) {
	role := &elasticsearchhandler.XPackSecurityRole{
		Cluster: h.Spec.Cluster,
		RunAs:   h.Spec.RunAs,
	}

	if h.Spec.Global != "" {
		global := make(map[string]any)
		if err := json.Unmarshal([]byte(h.Spec.Global), &global); err != nil {
			return nil, err
		}
		role.Global = global
	}

	if h.Spec.Metadata != "" {
		meta := make(map[string]any)
		if err := json.Unmarshal([]byte(h.Spec.Metadata), &meta); err != nil {
			return nil, err
		}
		role.Metadata = meta
	}

	if h.Spec.TransientMetadata != "" {
		tm := make(map[string]any)
		if err := json.Unmarshal([]byte(h.Spec.TransientMetadata), &tm); err != nil {
			return nil, err
		}
		role.TransientMetadata = tm
	}

	if h.Spec.Applications != nil {
		role.Applications = make([]elasticsearchhandler.XPackSecurityApplicationPrivileges, 0, len(h.Spec.Applications))
		for _, application := range h.Spec.Applications {
			role.Applications = append(role.Applications, elasticsearchhandler.XPackSecurityApplicationPrivileges{
				Application: application.Application,
				Privileges:  application.Privileges,
				Resources:   application.Resources,
			})
		}
	}

	if h.Spec.Indices != nil {
		role.Indices = make([]elasticsearchhandler.XPackSecurityIndicesPermissions, 0, len(h.Spec.Indices))
		for _, indice := range h.Spec.Indices {
			i := elasticsearchhandler.XPackSecurityIndicesPermissions{
				Names:      indice.Names,
				Privileges: indice.Privileges,
				Query:      indice.Query,
			}
			if indice.FieldSecurity != "" {
				fs := make(map[string]any)
				if err := json.Unmarshal([]byte(indice.FieldSecurity), &fs); err != nil {
					return nil, err
				}
				i.FieldSecurity = fs
			}
			role.Indices = append(role.Indices, i)
		}
	}

	return role, nil
}
