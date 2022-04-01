package helpers

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

func StructuredToUntructured(s any) (*unstructured.Unstructured, error) {
	data, err := runtime.DefaultUnstructuredConverter.ToUnstructured(s)
	if err != nil {
		return nil, err
	}

	us := &unstructured.Unstructured{
		Object: data,
	}

	return us, nil
}

func UnstructuredToStructured(us *unstructured.Unstructured, s any) error {
	return runtime.DefaultUnstructuredConverter.FromUnstructured(us.Object, s)
}
