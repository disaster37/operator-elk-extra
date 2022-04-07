package v1alpha1

import (
	"github.com/disaster37/operator-elk-extra/pkg/helpers"
	"github.com/stretchr/testify/assert"

	"golang.org/x/net/context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func (t *V1alpha1TestSuite) TestElasticsearchWatcherCRUD() {
	var (
		key              types.NamespacedName
		created, fetched *ElasticsearchWatcher
		err              error
	)

	key = types.NamespacedName{
		Name:      "foo-" + helpers.RandomString(5),
		Namespace: "default",
	}

	// Create object
	created = &ElasticsearchWatcher{
		ObjectMeta: metav1.ObjectMeta{
			Name:      key.Name,
			Namespace: key.Namespace,
		},
		Spec: ElasticsearchWatcherSpec{
			Trigger:   "fake",
			Input:     "fake",
			Condition: "fake",
			Actions:   "fake",
		},
	}
	err = t.k8sClient.Create(context.Background(), created)
	assert.NoError(t.T(), err)

	// Get object
	fetched = &ElasticsearchWatcher{}
	err = t.k8sClient.Get(context.Background(), key, fetched)
	assert.NoError(t.T(), err)
	assert.Equal(t.T(), created, fetched)

	// Delete object
	err = t.k8sClient.Delete(context.Background(), created)
	assert.NoError(t.T(), err)
	err = t.k8sClient.Get(context.Background(), key, created)
	assert.Error(t.T(), err)
}

func (t *V1alpha1TestSuite) TestElasticsearchWatcherGetObjectMeta() {
	meta := metav1.ObjectMeta{
		Name:      "test",
		Namespace: "test",
	}
	test := &ElasticsearchWatcher{
		ObjectMeta: meta,
	}

	assert.Equal(t.T(), meta, test.GetObjectMeta())
}

func (t *V1alpha1TestSuite) TestElasticsearchWatcherGetStatus() {
	status := ElasticsearchWatcherStatus{
		Conditions: []metav1.Condition{
			{
				Type: "test",
			},
		},
	}
	test := &ElasticsearchWatcher{
		Status: status,
	}

	assert.Equal(t.T(), status, test.GetStatus())
}
