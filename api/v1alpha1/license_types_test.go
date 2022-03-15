package v1alpha1

import (
	"time"

	"github.com/disaster37/operator-elk-extra/pkg/helpers"
	"github.com/stretchr/testify/assert"

	"golang.org/x/net/context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func (t *V1alpha1TestSuite) TestLicenseCRUD() {
	var (
		key              types.NamespacedName
		created, fetched *License
		err              error
	)

	key = types.NamespacedName{
		Name:      "foo-" + helpers.RandomString(5),
		Namespace: "default",
	}

	// Create object
	created = &License{
		ObjectMeta: metav1.ObjectMeta{
			Name:      key.Name,
			Namespace: key.Namespace,
		},
		Spec: LicenseSpec{
			SecretName: "test",
			ElasticsearchRef: &ElasticsearchRefSpec{
				Name: "test",
			},
		},
	}
	err = t.k8sClient.Create(context.Background(), created)
	assert.NoError(t.T(), err)

	// Get object
	fetched = &License{}
	err = t.k8sClient.Get(context.Background(), key, fetched)
	assert.NoError(t.T(), err)
	assert.Equal(t.T(), created, fetched)

	// Delete object
	err = t.k8sClient.Delete(context.Background(), created)
	assert.NoError(t.T(), err)
	err = t.k8sClient.Get(context.Background(), key, created)
	assert.Error(t.T(), err)
}

func (t *V1alpha1TestSuite) TestLicenseIsSubmitted() {
	license := &License{}
	assert.False(t.T(), license.IsSubmitted())

	license.Status.LicenseType = "basic"
	assert.True(t.T(), license.IsSubmitted())
}

func (t *V1alpha1TestSuite) TestLicenseIsBeingDeleted() {
	license := &License{
		ObjectMeta: metav1.ObjectMeta{
			DeletionTimestamp: &metav1.Time{
				Time: time.Now(),
			},
		},
	}
	assert.True(t.T(), license.IsBeingDeleted())
}

func (t *V1alpha1TestSuite) TestLicenseFinalizer() {
	license := &License{}

	license.AddFinalizer()
	assert.Equal(t.T(), 1, len(license.GetFinalizers()))
	assert.True(t.T(), license.HasFinalizer())

	license.RemoveFinalizer()
	assert.Equal(t.T(), 0, len(license.GetFinalizers()))
	assert.False(t.T(), license.HasFinalizer())
}
