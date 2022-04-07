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

// UserSpec defines the desired state of User
// +k8s:openapi-gen=true
type UserSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	ElasticsearchRefSpec `json:"elasticsearchRef"`

	// Enabled permit to enable user
	Enabled bool `json:"enabled"`

	// Email is the email user
	// +optional
	Email string `json:"email,omitempty"`

	// FullName is the full name
	// +optional
	FullName string `json:"full_name,omitempty"`

	// Metadata is the meta data
	// Is JSON string
	// +optional
	Metadata string `json:"metadata,omitempty"`

	// Secret permit to set password. Or you can use password hash
	// +optional
	Secret *UserSecret `json:"secret,omitempty"`

	// PasswordHash is the password as hash
	// +optional
	PasswordHash string `json:"password_hash,omitempty"`

	// Roles is the list of roles
	Roles []string `json:"roles"`
}

type UserSecret struct {

	// Name is the secret name
	Name string `json:"name"`

	// key is the key name on secret to read the effective password
	Key string `json:"key"`
}

// UserStatus defines the observed state of User
type UserStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Conditions []metav1.Condition `json:"conditions"`

	PasswordHash string `json:"passwordHash"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// User is the Schema for the users API
type User struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   UserSpec   `json:"spec,omitempty"`
	Status UserStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// UserList contains a list of User
type UserList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []User `json:"items"`
}

func init() {
	SchemeBuilder.Register(&User{}, &UserList{})
}

// GetObjectMeta permit to get the current ObjectMeta
func (h *User) GetObjectMeta() metav1.ObjectMeta {
	return h.ObjectMeta
}

// GetStatus permit to get the current status
func (h *User) GetStatus() any {
	return h.Status
}

// ToUser permit to convert to User object
func (h *User) ToUser() (*olivere.XPackSecurityPutUserRequest, error) {
	user := &olivere.XPackSecurityPutUserRequest{
		Enabled:      h.Spec.Enabled,
		Email:        h.Spec.Email,
		FullName:     h.Spec.FullName,
		Roles:        h.Spec.Roles,
		PasswordHash: h.Spec.PasswordHash,
	}

	if h.Spec.Metadata != "" {
		meta := make(map[string]any)
		if err := json.Unmarshal([]byte(h.Spec.Metadata), &meta); err != nil {
			return nil, err
		}
		user.Metadata = meta
	}

	return user, nil
}
