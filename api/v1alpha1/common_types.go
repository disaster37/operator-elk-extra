package v1alpha1

type ElasticsearchRefSpec struct {

	// Name is the Elasticsearch name object
	Name string `json:"name,omitempty"`
}

type ElasticsearchExternalRefSpec struct {

	// SecretName is the secret that contain the setting to connect on Elasticsearch that is not managed by ECK.
	// It need to contain the following keys: hosts, username, password, slefSignedCertificate
	SecretName string `json:"secretName,omitempty"`
}
