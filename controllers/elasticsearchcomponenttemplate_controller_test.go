package controllers

import (
	"context"
	"testing"
	"time"

	elkv1alpha1 "github.com/disaster37/operator-elk-extra/api/v1alpha1"
	"github.com/disaster37/operator-elk-extra/pkg/helpers"
	"github.com/disaster37/operator-elk-extra/pkg/mocks"
	"github.com/disaster37/operator-sdk-extra/pkg/test"
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

	key := types.NamespacedName{
		Name:      "t-component-" + helpers.RandomString(10),
		Namespace: "default",
	}
	ct := &elkv1alpha1.ElasticsearchComponentTemplate{}
	data := map[string]any{}

	testCase := test.NewTestCase(t.T(), t.k8sClient, key, ct, 5*time.Second, data)
	testCase.Steps = []test.TestStep{
		doCreateComponentTemplateStep(),
		doUpdateComponentTemplateStep(),
		doDeleteComponentTemplateStep(),
	}
	testCase.PreTest = doMockComponentTemplate(t.mockElasticsearchHandler)

	testCase.Run()
}

func doMockComponentTemplate(mockES *mocks.MockElasticsearchHandler) func(stepName *string, data map[string]any) error {
	return func(stepName *string, data map[string]any) (err error) {
		isCreated := false
		isUpdated := false

		mockES.EXPECT().ComponentTemplateGet(gomock.Any()).AnyTimes().DoAndReturn(func(name string) (*olivere.IndicesGetComponentTemplateData, error) {

			switch *stepName {
			case "create":
				if !isCreated {
					return nil, nil
				} else {
					resp := &olivere.IndicesGetComponentTemplateData{
						Settings: map[string]interface{}{"fake": "foo"},
					}
					return resp, nil
				}
			case "update":
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

		mockES.EXPECT().ComponentTemplateDiff(gomock.Any(), gomock.Any()).AnyTimes().DoAndReturn(func(actual, expected *olivere.IndicesGetComponentTemplateData) (string, error) {
			switch *stepName {
			case "create":
				if !isCreated {
					return "fake change", nil
				} else {
					return "", nil
				}
			case "update":
				if !isUpdated {
					return "fake change", nil
				} else {
					return "", nil
				}
			}

			return "", nil

		})

		mockES.EXPECT().ComponentTemplateUpdate(gomock.Any(), gomock.Any()).AnyTimes().DoAndReturn(func(name string, component *olivere.IndicesGetComponentTemplateData) error {
			switch *stepName {
			case "create":
				data["isCreated"] = true
				isCreated = true
				return nil
			case "update":
				data["isUpdated"] = true
				isUpdated = true
				return nil
			}

			return nil
		})

		mockES.EXPECT().ComponentTemplateDelete(gomock.Any()).AnyTimes().DoAndReturn(func(name string) error {
			data["isDeleted"] = true
			return nil
		})

		return nil
	}
}

func doCreateComponentTemplateStep() test.TestStep {
	return test.TestStep{
		Name: "create",
		Do: func(c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			logrus.Infof("=== Add new component template %s/%s ===", key.Namespace, key.Name)

			ct := &elkv1alpha1.ElasticsearchComponentTemplate{
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
			if err = c.Create(context.Background(), ct); err != nil {
				return err
			}

			return nil
		},
		Check: func(t *testing.T, c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			ct := &elkv1alpha1.ElasticsearchComponentTemplate{}
			isCreated := false

			isTimeout, err := RunWithTimeout(func() error {
				if err := c.Get(context.Background(), key, ct); err != nil {
					t.Fatal(err)
				}
				if b, ok := data["isCreated"]; ok {
					isCreated = b.(bool)
				}
				if !isCreated {
					return errors.New("Not yet created")
				}
				return nil
			}, time.Second*30, time.Second*1)

			if err != nil || isTimeout {
				t.Fatalf("Failed to get component template: %s", err.Error())
			}
			assert.True(t, condition.IsStatusConditionPresentAndEqual(ct.Status.Conditions, componentCondition, metav1.ConditionTrue))

			return nil
		},
	}
}

func doUpdateComponentTemplateStep() test.TestStep {
	return test.TestStep{
		Name: "update",
		Do: func(c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			logrus.Infof("=== Update component template %s/%s ===", key.Namespace, key.Name)

			if o == nil {
				return errors.New("Component template is null")
			}
			ct := o.(*elkv1alpha1.ElasticsearchComponentTemplate)

			ct.Spec.Settings = `{
				"fake": "foo2"
			}`
			if err = c.Update(context.Background(), ct); err != nil {
				return err
			}

			return nil
		},
		Check: func(t *testing.T, c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			ct := &elkv1alpha1.ElasticsearchComponentTemplate{}
			isUpdated := false

			isTimeout, err := RunWithTimeout(func() error {
				if err := c.Get(context.Background(), key, ct); err != nil {
					t.Fatal(err)
				}
				if b, ok := data["isUpdated"]; ok {
					isUpdated = b.(bool)
				}
				if !isUpdated {
					return errors.New("Not yet updated")
				}
				return nil
			}, time.Second*30, time.Second*1)
			if err != nil || isTimeout {
				t.Fatalf("Failed to get component template: %s", err.Error())
			}
			assert.True(t, condition.IsStatusConditionPresentAndEqual(ct.Status.Conditions, componentCondition, metav1.ConditionTrue))

			return nil
		},
	}
}

func doDeleteComponentTemplateStep() test.TestStep {
	return test.TestStep{
		Name: "delete",
		Do: func(c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			logrus.Infof("=== Delete component template %s/%s ===", key.Namespace, key.Name)

			if o == nil {
				return errors.New("Component template is null")
			}
			ct := o.(*elkv1alpha1.ElasticsearchComponentTemplate)

			wait := int64(0)
			if err = c.Delete(context.Background(), ct, &client.DeleteOptions{GracePeriodSeconds: &wait}); err != nil {
				return err
			}

			return nil
		},
		Check: func(t *testing.T, c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			ct := &elkv1alpha1.ElasticsearchComponentTemplate{}
			isDeleted := false

			isTimeout, err := RunWithTimeout(func() error {
				if err = c.Get(context.Background(), key, ct); err != nil {
					if k8serrors.IsNotFound(err) {
						isDeleted = true
						return nil
					}
					t.Fatal(err)
				}

				return errors.New("Not yet deleted")
			}, time.Second*30, time.Second*1)
			if err != nil || isTimeout {
				t.Fatalf("Component template stil exist: %s", err.Error())
			}
			assert.True(t, isDeleted)
			time.Sleep(10 * time.Second)

			return nil
		},
	}
}
