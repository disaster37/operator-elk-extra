package v1alpha1

type ElasticsearchRefSpec struct {
	// Name is the Elasticsearch name object
	Name string `json:"name,omitempty"`
}

type ElasticsearchExternalRefSpec struct {

	// Addresses is the list of Elasticsearch addresses
	Addresses []string `json:"addresses,omitempty"`

	// SecretName is the secret that contain the setting to connect on Elasticsearch that is not managed by ECK.
	// It need to contain only one entry. The user is the key, and the password is the data
	SecretName string `json:"secretName,omitempty"`
}

type ElasticsearchRefererStd struct {
	// ElasticsearchRef is the Elasticsearch reference to connect on it (managed by ECK)
	// +optional
	ElasticsearchRef *ElasticsearchRefSpec `json:"ref,omitempty"`

	//ElasticsearchExternalRef is the elasticsearch external reference to connect on it (not managed by ECK)
	// +optional
	ElasticsearchExternalRef *ElasticsearchExternalRefSpec `json:"externalRef,omitempty"`
}

func (h *ElasticsearchRefererStd) GetElasticsearchRef() *ElasticsearchRefSpec {
	return h.ElasticsearchRef
}

func (h *ElasticsearchRefererStd) GetElasticsearchExternalRef() *ElasticsearchExternalRefSpec {
	return h.ElasticsearchExternalRef
}
