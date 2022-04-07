package controllers

import (
	"context"
	"time"

	elkv1alpha1 "github.com/disaster37/operator-elk-extra/api/v1alpha1"
	"github.com/disaster37/operator-elk-extra/pkg/helpers"
	"github.com/golang/mock/gomock"
	olivere "github.com/olivere/elastic/v7"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/controller-runtime/pkg/client"

	//core "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	condition "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func (t *ControllerTestSuite) TestElasticsearchIndexTemplateReconciler() {

	var (
		err       error
		isTimeout bool
		fetched   *elkv1alpha1.ElasticsearchIndexTemplate
		test      string
		isCreated bool
		isDeleted bool
		isUpdated bool
	)

	templateName := "t-itemplate-" + helpers.RandomString(10)
	key := types.NamespacedName{
		Name:      templateName,
		Namespace: "default",
	}
	t.mockElasticsearchHandler.EXPECT().IndexTemplateGet(gomock.Any()).AnyTimes().DoAndReturn(func(name string) (*olivere.IndicesGetIndexTemplate, error) {

		switch test {
		case "no_template":
			if !isCreated {
				return nil, nil
			} else {

				resp := &olivere.IndicesGetIndexTemplate{
					IndexPatterns: []string{"test"},
				}

				return resp, nil
			}
		case "template_update":
			if !isUpdated {
				resp := &olivere.IndicesGetIndexTemplate{
					IndexPatterns: []string{"test"},
				}
				return resp, nil
			} else {
				resp := &olivere.IndicesGetIndexTemplate{
					IndexPatterns: []string{"test2"},
				}
				return resp, nil
			}
		}

		return nil, nil

	})
	t.mockElasticsearchHandler.EXPECT().IndexTemplateDiff(gomock.Any(), gomock.Any()).AnyTimes().DoAndReturn(func(actual, expected *olivere.IndicesGetIndexTemplate) (string, error) {
		switch test {
		case "no_template":
			if !isCreated {
				return "fake change", nil
			} else {
				return "", nil
			}
		case "template_update":
			if !isUpdated {
				return "fake change", nil
			} else {
				return "", nil
			}
		}

		return "", nil

	})
	t.mockElasticsearchHandler.EXPECT().IndexTemplateUpdate(gomock.Any(), gomock.Any()).AnyTimes().DoAndReturn(func(name string, policy *olivere.IndicesGetIndexTemplate) error {
		switch test {
		case "no_template":
			isCreated = true
			return nil
		case "template_update":
			isUpdated = true
			return nil
		}

		return nil
	})
	t.mockElasticsearchHandler.EXPECT().IndexTemplateDelete(gomock.Any()).AnyTimes().DoAndReturn(func(name string) error {
		switch test {
		case "template_delete":
			isDeleted = true
			return nil
		}

		return nil
	})

	// When add new index template
	logrus.Info("==================================== When add new index template")
	test = "no_template"
	toCreate := &elkv1alpha1.ElasticsearchIndexTemplate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      key.Name,
			Namespace: key.Namespace,
		},
		Spec: elkv1alpha1.ElasticsearchIndexTemplateSpec{
			ElasticsearchRefSpec: elkv1alpha1.ElasticsearchRefSpec{
				Name: "test",
			},
			IndexPatterns: []string{"test"},
		},
	}
	if err = t.k8sClient.Create(context.Background(), toCreate); err != nil {
		t.T().Fatal(err)
	}

	isTimeout, err = RunWithTimeout(func() error {
		fetched = &elkv1alpha1.ElasticsearchIndexTemplate{}
		if err := t.k8sClient.Get(context.Background(), key, fetched); err != nil {
			t.T().Fatal(err)
		}
		if !isCreated {
			return errors.New("Not yet created")
		}
		return nil
	}, time.Second*30, time.Second*1)
	if err != nil || isTimeout {
		t.T().Fatalf("Failed to get index template: %s", err.Error())
	}
	assert.True(t.T(), condition.IsStatusConditionPresentAndEqual(fetched.Status.Conditions, indexTemplateCondition, metav1.ConditionTrue))
	time.Sleep(10 * time.Second)

	// When update index template
	logrus.Info("==================================== When update index template")
	test = "template_update"
	isUpdated = false
	fetched = &elkv1alpha1.ElasticsearchIndexTemplate{}
	if err := t.k8sClient.Get(context.Background(), key, fetched); err != nil {
		t.T().Fatal(err)
	}
	fetched.Spec.IndexPatterns = []string{"test2"}
	if err = t.k8sClient.Update(context.Background(), fetched); err != nil {
		t.T().Fatal(err)
	}

	isTimeout, err = RunWithTimeout(func() error {
		fetched = &elkv1alpha1.ElasticsearchIndexTemplate{}
		if err := t.k8sClient.Get(context.Background(), key, fetched); err != nil {
			t.T().Fatal(err)
		}
		if !isUpdated {
			return errors.New("Not yet updated")
		}
		return nil
	}, time.Second*30, time.Second*1)
	if err != nil || isTimeout {
		t.T().Fatalf("Failed to get index template: %s", err.Error())
	}
	assert.True(t.T(), condition.IsStatusConditionPresentAndEqual(fetched.Status.Conditions, indexTemplateCondition, metav1.ConditionTrue))
	time.Sleep(10 * time.Second)

	// When remove index template
	logrus.Info("==================================== When remove index template")
	test = "template_delete"
	wait := int64(0)
	isDeleted = false
	if err = t.k8sClient.Delete(context.Background(), fetched, &client.DeleteOptions{GracePeriodSeconds: &wait}); err != nil {
		t.T().Fatal(err)
	}
	isTimeout, err = RunWithTimeout(func() error {
		if err = t.k8sClient.Get(context.Background(), key, fetched); err != nil {
			if k8serrors.IsNotFound(err) {
				return nil
			}
			t.T().Fatal(err)
		}

		return errors.New("Not yet deleted")
	}, time.Second*30, time.Second*1)
	if err != nil || isTimeout {
		t.T().Fatalf("Index template stil exist: %s", err.Error())
	}
	assert.True(t.T(), isDeleted)
	time.Sleep(10 * time.Second)

}
