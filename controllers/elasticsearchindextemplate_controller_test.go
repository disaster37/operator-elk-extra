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
	"sigs.k8s.io/controller-runtime/pkg/client"

	//core "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	condition "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func (t *ControllerTestSuite) TestElasticsearchIndexTemplateReconciler() {
	key := types.NamespacedName{
		Name:      "t-itemplate-" + helpers.RandomString(10),
		Namespace: "default",
	}
	template := &elkv1alpha1.ElasticsearchIndexTemplate{}
	data := map[string]any{}

	testCase := test.NewTestCase(t.T(), t.k8sClient, key, template, 5*time.Second, data)
	testCase.Steps = []test.TestStep{
		doCreateIndexTemplateStep(),
		doUpdateIndexTemplateStep(),
		doDeleteIndexTemplateStep(),
	}
	testCase.PreTest = doMockIndexTemplate(t.mockElasticsearchHandler)

	testCase.Run()
}

func doMockIndexTemplate(mockES *mocks.MockElasticsearchHandler) func(stepName *string, data map[string]any) error {
	return func(stepName *string, data map[string]any) (err error) {
		isCreated := false
		isUpdated := false

		mockES.EXPECT().IndexTemplateGet(gomock.Any()).AnyTimes().DoAndReturn(func(name string) (*olivere.IndicesGetIndexTemplate, error) {

			switch *stepName {
			case "create":
				if !isCreated {
					return nil, nil
				} else {

					resp := &olivere.IndicesGetIndexTemplate{
						IndexPatterns: []string{"test"},
					}
					return resp, nil
				}
			case "update":
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

		mockES.EXPECT().IndexTemplateDiff(gomock.Any(), gomock.Any()).AnyTimes().DoAndReturn(func(actual, expected *olivere.IndicesGetIndexTemplate) (string, error) {
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

		mockES.EXPECT().IndexTemplateUpdate(gomock.Any(), gomock.Any()).AnyTimes().DoAndReturn(func(name string, policy *olivere.IndicesGetIndexTemplate) error {
			switch *stepName {
			case "create":
				isCreated = true
				data["isCreated"] = true
				return nil
			case "update":
				isUpdated = true
				data["isUpdated"] = true
				return nil
			}

			return nil
		})

		mockES.EXPECT().IndexTemplateDelete(gomock.Any()).AnyTimes().DoAndReturn(func(name string) error {
			data["isDeleted"] = true
			return nil
		})

		return nil
	}
}

func doCreateIndexTemplateStep() test.TestStep {
	return test.TestStep{
		Name: "create",
		Do: func(c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			logrus.Infof("=== Add new index template %s/%s ===", key.Namespace, key.Name)

			template := &elkv1alpha1.ElasticsearchIndexTemplate{
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
			if err = c.Create(context.Background(), template); err != nil {
				return err
			}

			return nil
		},
		Check: func(t *testing.T, c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			template := &elkv1alpha1.ElasticsearchIndexTemplate{}
			isCreated := false

			isTimeout, err := RunWithTimeout(func() error {
				if err := c.Get(context.Background(), key, template); err != nil {
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
				t.Fatalf("Failed to get index template: %s", err.Error())
			}
			assert.True(t, condition.IsStatusConditionPresentAndEqual(template.Status.Conditions, indexTemplateCondition, metav1.ConditionTrue))
			time.Sleep(10 * time.Second)

			return nil
		},
	}
}

func doUpdateIndexTemplateStep() test.TestStep {
	return test.TestStep{
		Name: "update",
		Do: func(c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			logrus.Infof("=== Update index template %s/%s ===", key.Namespace, key.Name)

			if o == nil {
				return errors.New("Index template is null")
			}
			template := o.(*elkv1alpha1.ElasticsearchIndexTemplate)

			template.Spec.IndexPatterns = []string{"test2"}
			if err = c.Update(context.Background(), template); err != nil {
				return err
			}

			return nil
		},
		Check: func(t *testing.T, c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			template := &elkv1alpha1.ElasticsearchIndexTemplate{}
			isUpdated := false

			isTimeout, err := RunWithTimeout(func() error {
				if err := c.Get(context.Background(), key, template); err != nil {
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
				t.Fatalf("Failed to get index template: %s", err.Error())
			}
			assert.True(t, condition.IsStatusConditionPresentAndEqual(template.Status.Conditions, indexTemplateCondition, metav1.ConditionTrue))

			return nil
		},
	}
}

func doDeleteIndexTemplateStep() test.TestStep {
	return test.TestStep{
		Name: "delete",
		Do: func(c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			logrus.Infof("=== Delete index template %s/%s ===", key.Namespace, key.Name)

			if o == nil {
				return errors.New("Index template is null")
			}
			template := o.(*elkv1alpha1.ElasticsearchIndexTemplate)

			wait := int64(0)
			if err = c.Delete(context.Background(), template, &client.DeleteOptions{GracePeriodSeconds: &wait}); err != nil {
				return err
			}

			return nil
		},
		Check: func(t *testing.T, c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			template := &elkv1alpha1.ElasticsearchIndexTemplate{}
			isDeleted := false

			isTimeout, err := RunWithTimeout(func() error {
				if err = c.Get(context.Background(), key, template); err != nil {
					if k8serrors.IsNotFound(err) {
						isDeleted = true
						return nil
					}
					t.Fatal(err)
				}

				return errors.New("Not yet deleted")
			}, time.Second*30, time.Second*1)
			if err != nil || isTimeout {
				t.Fatalf("Index template stil exist: %s", err.Error())
			}
			assert.True(t, isDeleted)
			return nil
		},
	}
}
