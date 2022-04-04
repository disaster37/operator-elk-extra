//go:build !ignore_autogenerated
// +build !ignore_autogenerated

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

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ElasticsearchComponentTemplate) DeepCopyInto(out *ElasticsearchComponentTemplate) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ElasticsearchComponentTemplate.
func (in *ElasticsearchComponentTemplate) DeepCopy() *ElasticsearchComponentTemplate {
	if in == nil {
		return nil
	}
	out := new(ElasticsearchComponentTemplate)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ElasticsearchComponentTemplate) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ElasticsearchComponentTemplateList) DeepCopyInto(out *ElasticsearchComponentTemplateList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]ElasticsearchComponentTemplate, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ElasticsearchComponentTemplateList.
func (in *ElasticsearchComponentTemplateList) DeepCopy() *ElasticsearchComponentTemplateList {
	if in == nil {
		return nil
	}
	out := new(ElasticsearchComponentTemplateList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ElasticsearchComponentTemplateList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ElasticsearchComponentTemplateSpec) DeepCopyInto(out *ElasticsearchComponentTemplateSpec) {
	*out = *in
	in.ElasticsearchRefSpec.DeepCopyInto(&out.ElasticsearchRefSpec)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ElasticsearchComponentTemplateSpec.
func (in *ElasticsearchComponentTemplateSpec) DeepCopy() *ElasticsearchComponentTemplateSpec {
	if in == nil {
		return nil
	}
	out := new(ElasticsearchComponentTemplateSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ElasticsearchComponentTemplateStatus) DeepCopyInto(out *ElasticsearchComponentTemplateStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]v1.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ElasticsearchComponentTemplateStatus.
func (in *ElasticsearchComponentTemplateStatus) DeepCopy() *ElasticsearchComponentTemplateStatus {
	if in == nil {
		return nil
	}
	out := new(ElasticsearchComponentTemplateStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ElasticsearchILM) DeepCopyInto(out *ElasticsearchILM) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ElasticsearchILM.
func (in *ElasticsearchILM) DeepCopy() *ElasticsearchILM {
	if in == nil {
		return nil
	}
	out := new(ElasticsearchILM)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ElasticsearchILM) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ElasticsearchILMList) DeepCopyInto(out *ElasticsearchILMList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]ElasticsearchILM, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ElasticsearchILMList.
func (in *ElasticsearchILMList) DeepCopy() *ElasticsearchILMList {
	if in == nil {
		return nil
	}
	out := new(ElasticsearchILMList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ElasticsearchILMList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ElasticsearchILMSpec) DeepCopyInto(out *ElasticsearchILMSpec) {
	*out = *in
	in.ElasticsearchRefSpec.DeepCopyInto(&out.ElasticsearchRefSpec)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ElasticsearchILMSpec.
func (in *ElasticsearchILMSpec) DeepCopy() *ElasticsearchILMSpec {
	if in == nil {
		return nil
	}
	out := new(ElasticsearchILMSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ElasticsearchILMStatus) DeepCopyInto(out *ElasticsearchILMStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]v1.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ElasticsearchILMStatus.
func (in *ElasticsearchILMStatus) DeepCopy() *ElasticsearchILMStatus {
	if in == nil {
		return nil
	}
	out := new(ElasticsearchILMStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ElasticsearchIndexTemplate) DeepCopyInto(out *ElasticsearchIndexTemplate) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ElasticsearchIndexTemplate.
func (in *ElasticsearchIndexTemplate) DeepCopy() *ElasticsearchIndexTemplate {
	if in == nil {
		return nil
	}
	out := new(ElasticsearchIndexTemplate)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ElasticsearchIndexTemplate) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ElasticsearchIndexTemplateList) DeepCopyInto(out *ElasticsearchIndexTemplateList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]ElasticsearchIndexTemplate, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ElasticsearchIndexTemplateList.
func (in *ElasticsearchIndexTemplateList) DeepCopy() *ElasticsearchIndexTemplateList {
	if in == nil {
		return nil
	}
	out := new(ElasticsearchIndexTemplateList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ElasticsearchIndexTemplateList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ElasticsearchIndexTemplateSpec) DeepCopyInto(out *ElasticsearchIndexTemplateSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ElasticsearchIndexTemplateSpec.
func (in *ElasticsearchIndexTemplateSpec) DeepCopy() *ElasticsearchIndexTemplateSpec {
	if in == nil {
		return nil
	}
	out := new(ElasticsearchIndexTemplateSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ElasticsearchIndexTemplateStatus) DeepCopyInto(out *ElasticsearchIndexTemplateStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ElasticsearchIndexTemplateStatus.
func (in *ElasticsearchIndexTemplateStatus) DeepCopy() *ElasticsearchIndexTemplateStatus {
	if in == nil {
		return nil
	}
	out := new(ElasticsearchIndexTemplateStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ElasticsearchRefSpec) DeepCopyInto(out *ElasticsearchRefSpec) {
	*out = *in
	if in.Addresses != nil {
		in, out := &in.Addresses, &out.Addresses
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ElasticsearchRefSpec.
func (in *ElasticsearchRefSpec) DeepCopy() *ElasticsearchRefSpec {
	if in == nil {
		return nil
	}
	out := new(ElasticsearchRefSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ElasticsearchSLM) DeepCopyInto(out *ElasticsearchSLM) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ElasticsearchSLM.
func (in *ElasticsearchSLM) DeepCopy() *ElasticsearchSLM {
	if in == nil {
		return nil
	}
	out := new(ElasticsearchSLM)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ElasticsearchSLM) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ElasticsearchSLMConfig) DeepCopyInto(out *ElasticsearchSLMConfig) {
	*out = *in
	if in.Indices != nil {
		in, out := &in.Indices, &out.Indices
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.FeatureStates != nil {
		in, out := &in.FeatureStates, &out.FeatureStates
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Metadata != nil {
		in, out := &in.Metadata, &out.Metadata
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ElasticsearchSLMConfig.
func (in *ElasticsearchSLMConfig) DeepCopy() *ElasticsearchSLMConfig {
	if in == nil {
		return nil
	}
	out := new(ElasticsearchSLMConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ElasticsearchSLMList) DeepCopyInto(out *ElasticsearchSLMList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]ElasticsearchSLM, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ElasticsearchSLMList.
func (in *ElasticsearchSLMList) DeepCopy() *ElasticsearchSLMList {
	if in == nil {
		return nil
	}
	out := new(ElasticsearchSLMList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ElasticsearchSLMList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ElasticsearchSLMRetention) DeepCopyInto(out *ElasticsearchSLMRetention) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ElasticsearchSLMRetention.
func (in *ElasticsearchSLMRetention) DeepCopy() *ElasticsearchSLMRetention {
	if in == nil {
		return nil
	}
	out := new(ElasticsearchSLMRetention)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ElasticsearchSLMSpec) DeepCopyInto(out *ElasticsearchSLMSpec) {
	*out = *in
	in.ElasticsearchRefSpec.DeepCopyInto(&out.ElasticsearchRefSpec)
	in.Config.DeepCopyInto(&out.Config)
	if in.Retention != nil {
		in, out := &in.Retention, &out.Retention
		*out = new(ElasticsearchSLMRetention)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ElasticsearchSLMSpec.
func (in *ElasticsearchSLMSpec) DeepCopy() *ElasticsearchSLMSpec {
	if in == nil {
		return nil
	}
	out := new(ElasticsearchSLMSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ElasticsearchSLMStatus) DeepCopyInto(out *ElasticsearchSLMStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]v1.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ElasticsearchSLMStatus.
func (in *ElasticsearchSLMStatus) DeepCopy() *ElasticsearchSLMStatus {
	if in == nil {
		return nil
	}
	out := new(ElasticsearchSLMStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ElasticsearchSnapshotRepository) DeepCopyInto(out *ElasticsearchSnapshotRepository) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ElasticsearchSnapshotRepository.
func (in *ElasticsearchSnapshotRepository) DeepCopy() *ElasticsearchSnapshotRepository {
	if in == nil {
		return nil
	}
	out := new(ElasticsearchSnapshotRepository)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ElasticsearchSnapshotRepository) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ElasticsearchSnapshotRepositoryList) DeepCopyInto(out *ElasticsearchSnapshotRepositoryList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]ElasticsearchSnapshotRepository, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ElasticsearchSnapshotRepositoryList.
func (in *ElasticsearchSnapshotRepositoryList) DeepCopy() *ElasticsearchSnapshotRepositoryList {
	if in == nil {
		return nil
	}
	out := new(ElasticsearchSnapshotRepositoryList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ElasticsearchSnapshotRepositoryList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ElasticsearchSnapshotRepositorySpec) DeepCopyInto(out *ElasticsearchSnapshotRepositorySpec) {
	*out = *in
	in.ElasticsearchRefSpec.DeepCopyInto(&out.ElasticsearchRefSpec)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ElasticsearchSnapshotRepositorySpec.
func (in *ElasticsearchSnapshotRepositorySpec) DeepCopy() *ElasticsearchSnapshotRepositorySpec {
	if in == nil {
		return nil
	}
	out := new(ElasticsearchSnapshotRepositorySpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ElasticsearchSnapshotRepositoryStatus) DeepCopyInto(out *ElasticsearchSnapshotRepositoryStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]v1.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ElasticsearchSnapshotRepositoryStatus.
func (in *ElasticsearchSnapshotRepositoryStatus) DeepCopy() *ElasticsearchSnapshotRepositoryStatus {
	if in == nil {
		return nil
	}
	out := new(ElasticsearchSnapshotRepositoryStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ElasticsearchWatcher) DeepCopyInto(out *ElasticsearchWatcher) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ElasticsearchWatcher.
func (in *ElasticsearchWatcher) DeepCopy() *ElasticsearchWatcher {
	if in == nil {
		return nil
	}
	out := new(ElasticsearchWatcher)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ElasticsearchWatcher) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ElasticsearchWatcherList) DeepCopyInto(out *ElasticsearchWatcherList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]ElasticsearchWatcher, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ElasticsearchWatcherList.
func (in *ElasticsearchWatcherList) DeepCopy() *ElasticsearchWatcherList {
	if in == nil {
		return nil
	}
	out := new(ElasticsearchWatcherList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ElasticsearchWatcherList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ElasticsearchWatcherSpec) DeepCopyInto(out *ElasticsearchWatcherSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ElasticsearchWatcherSpec.
func (in *ElasticsearchWatcherSpec) DeepCopy() *ElasticsearchWatcherSpec {
	if in == nil {
		return nil
	}
	out := new(ElasticsearchWatcherSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ElasticsearchWatcherStatus) DeepCopyInto(out *ElasticsearchWatcherStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ElasticsearchWatcherStatus.
func (in *ElasticsearchWatcherStatus) DeepCopy() *ElasticsearchWatcherStatus {
	if in == nil {
		return nil
	}
	out := new(ElasticsearchWatcherStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *License) DeepCopyInto(out *License) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new License.
func (in *License) DeepCopy() *License {
	if in == nil {
		return nil
	}
	out := new(License)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *License) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LicenseList) DeepCopyInto(out *LicenseList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]License, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LicenseList.
func (in *LicenseList) DeepCopy() *LicenseList {
	if in == nil {
		return nil
	}
	out := new(LicenseList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *LicenseList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LicenseSpec) DeepCopyInto(out *LicenseSpec) {
	*out = *in
	in.ElasticsearchRefSpec.DeepCopyInto(&out.ElasticsearchRefSpec)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LicenseSpec.
func (in *LicenseSpec) DeepCopy() *LicenseSpec {
	if in == nil {
		return nil
	}
	out := new(LicenseSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LicenseStatus) DeepCopyInto(out *LicenseStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]v1.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LicenseStatus.
func (in *LicenseStatus) DeepCopy() *LicenseStatus {
	if in == nil {
		return nil
	}
	out := new(LicenseStatus)
	in.DeepCopyInto(out)
	return out
}
