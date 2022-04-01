package v1alpha1

import (
	"github.com/disaster37/operator-elk-extra/pkg/helpers"
	"github.com/stretchr/testify/assert"

	"golang.org/x/net/context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func (t *V1alpha1TestSuite) TestElasticsearchSLMCRUD() {
	var (
		key              types.NamespacedName
		created, fetched *ElasticsearchSLM
		err              error
	)

	key = types.NamespacedName{
		Name:      "foo-" + helpers.RandomString(5),
		Namespace: "default",
	}

	// Create object
	created = &ElasticsearchSLM{
		ObjectMeta: metav1.ObjectMeta{
			Name:      key.Name,
			Namespace: key.Namespace,
		},
		Spec: ElasticsearchSLMSpec{
			Policy: "fake",
		},
	}
	err = t.k8sClient.Create(context.Background(), created)
	assert.NoError(t.T(), err)

	// Get object
	fetched = &ElasticsearchSLM{}
	err = t.k8sClient.Get(context.Background(), key, fetched)
	assert.NoError(t.T(), err)
	assert.Equal(t.T(), created, fetched)

	// Delete object
	err = t.k8sClient.Delete(context.Background(), created)
	assert.NoError(t.T(), err)
	err = t.k8sClient.Get(context.Background(), key, created)
	assert.Error(t.T(), err)
}

func (t *V1alpha1TestSuite) TestElasticsearchSLMGetObjectMeta() {
	meta := metav1.ObjectMeta{
		Name:      "test",
		Namespace: "test",
	}
	test := &ElasticsearchSLM{
		ObjectMeta: meta,
		Spec: ElasticsearchSLMSpec{
			Policy: "fake",
		},
	}

	assert.Equal(t.T(), meta, test.GetObjectMeta())
}

func (t *V1alpha1TestSuite) TestElasticsearchSLMGetStatus() {
	status := ElasticsearchSLMStatus{
		Conditions: []metav1.Condition{
			{
				Type: "test",
			},
		},
	}
	test := &ElasticsearchSLM{
		Spec: ElasticsearchSLMSpec{
			Policy: "fake",
		},
		Status: status,
	}

	assert.Equal(t.T(), status, test.GetStatus())
}
