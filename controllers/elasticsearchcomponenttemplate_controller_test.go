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
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	condition "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (t *ControllerTestSuite) TestElasticsearchComponentTemplateReconciler() {

	var (
		err       error
		isTimeout bool
		fetched   *elkv1alpha1.ElasticsearchComponentTemplate
		test      string
		isCreated bool
		isDeleted bool
		isUpdated bool
	)

	componentName := "t-component-" + helpers.RandomString(10)
	key := types.NamespacedName{
		Name:      componentName,
		Namespace: "default",
	}
	t.mockElasticsearchHandler.EXPECT().ComponentTemplateGet(gomock.Any()).AnyTimes().DoAndReturn(func(name string) (*olivere.IndicesGetComponentTemplateData, error) {

		switch test {
		case "no_component":
			if !isCreated {
				return nil, nil
			} else {
				resp := &olivere.IndicesGetComponentTemplateData{
					Settings: map[string]interface{}{"fake": "foo"},
				}
				return resp, nil
			}
		case "component_update":
			if !isUpdated {

				resp := &olivere.IndicesGetComponentTemplateData{
					Settings: map[string]interface{}{"fake": "foo"},
				}
				return resp, nil
			} else {
				resp := &olivere.IndicesGetComponentTemplateData{
					Settings: map[string]interface{}{"fake": "foo2"},
				}
				return resp, nil
			}
		}

		return nil, nil

	})
	t.mockElasticsearchHandler.EXPECT().ComponentTemplateDiff(gomock.Any(), gomock.Any()).AnyTimes().DoAndReturn(func(actual, expected *olivere.IndicesGetComponentTemplateData) (string, error) {
		switch test {
		case "no_component":
			if !isCreated {
				return "fake change", nil
			} else {
				return "", nil
			}
		case "component_update":
			if !isUpdated {
				return "fake change", nil
			} else {
				return "", nil
			}
		}

		return "", nil

	})
	t.mockElasticsearchHandler.EXPECT().ComponentTemplateUpdate(gomock.Any(), gomock.Any()).AnyTimes().DoAndReturn(func(name string, component *olivere.IndicesGetComponentTemplateData) error {
		switch test {
		case "no_component":
			isCreated = true
			return nil
		case "component_update":
			isUpdated = true
			return nil
		}

		return nil
	})
	t.mockElasticsearchHandler.EXPECT().ComponentTemplateDelete(gomock.Any()).AnyTimes().DoAndReturn(func(name string) error {
		switch test {
		case "component_delete":
			isDeleted = true
			return nil
		}

		return nil
	})

	// When add new component
	logrus.Info("==================================== When add new component template")
	test = "no_component"
	toCreate := &elkv1alpha1.ElasticsearchComponentTemplate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      key.Name,
			Namespace: key.Namespace,
		},
		Spec: elkv1alpha1.ElasticsearchComponentTemplateSpec{
			ElasticsearchRefSpec: elkv1alpha1.ElasticsearchRefSpec{
				Name: "test",
			},
			Settings: `
			{
				"fake": "foo"
			}
			`,
		},
	}
	if err = t.k8sClient.Create(context.Background(), toCreate); err != nil {
		t.T().Fatal(err)
	}

	isTimeout, err = RunWithTimeout(func() error {
		fetched = &elkv1alpha1.ElasticsearchComponentTemplate{}
		if err := t.k8sClient.Get(context.Background(), key, fetched); err != nil {
			t.T().Fatal(err)
		}
		if !isCreated {
			return errors.New("Not yet created")
		}
		return nil
	}, time.Second*30, time.Second*1)
	if err != nil || isTimeout {
		t.T().Fatalf("Failed to get component template: %s", err.Error())
	}
	assert.True(t.T(), condition.IsStatusConditionPresentAndEqual(fetched.Status.Conditions, componentCondition, metav1.ConditionTrue))
	time.Sleep(10 * time.Second)

	// When update component
	logrus.Info("==================================== When update component template")
	test = "component_update"
	isUpdated = false
	fetched = &elkv1alpha1.ElasticsearchComponentTemplate{}
	if err := t.k8sClient.Get(context.Background(), key, fetched); err != nil {
		t.T().Fatal(err)
	}
	fetched.Spec.Settings = `{
		"fake": "foo2"
	}`
	if err = t.k8sClient.Update(context.Background(), fetched); err != nil {
		t.T().Fatal(err)
	}

	isTimeout, err = RunWithTimeout(func() error {
		fetched = &elkv1alpha1.ElasticsearchComponentTemplate{}
		if err := t.k8sClient.Get(context.Background(), key, fetched); err != nil {
			t.T().Fatal(err)
		}
		if !isUpdated {
			return errors.New("Not yet updated")
		}
		return nil
	}, time.Second*30, time.Second*1)
	if err != nil || isTimeout {
		t.T().Fatalf("Failed to get component template: %s", err.Error())
	}
	assert.True(t.T(), condition.IsStatusConditionPresentAndEqual(fetched.Status.Conditions, componentCondition, metav1.ConditionTrue))
	time.Sleep(10 * time.Second)

	// When remove component
	logrus.Info("==================================== When remove component template")
	test = "component_delete"
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
		t.T().Fatalf("Component template stil exist: %s", err.Error())
	}
	assert.True(t.T(), isDeleted)
	time.Sleep(10 * time.Second)

}
