package elasticsearch

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	GVR = schema.GroupVersionResource{
		Group:    "elasticsearch.k8s.elastic.co",
		Version:  "v1",
		Resource: "elasticsearches",
	}
)

type Elasticsearch struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ElasticsearchSpec `json:"spec,omitempty"`
}

type ElasticsearchSpec struct {
	HTTP ElasticsearchSpecHTTP `json:"http,omitempty"`
}

type ElasticsearchSpecHTTP struct {
	TLS ElasticsearchSpecHTTPTLS `json:"tls,omitempty"`
}

type ElasticsearchSpecHTTPTLS struct {
	SelfSignedCertificate ElasticsearchSpecHTTPTLSSelfSignedCertificate `json:"selfSignedCertificate,omitempty"`
}

type ElasticsearchSpecHTTPTLSSelfSignedCertificate struct {
	Disabled bool `json:"disabled,omitempty"`
}
