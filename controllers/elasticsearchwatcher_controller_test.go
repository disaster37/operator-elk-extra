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

func (t *ControllerTestSuite) TestElasticsearchWatcherReconciler() {
	key := types.NamespacedName{
		Name:      "t-watch-" + helpers.RandomString(10),
		Namespace: "default",
	}
	watch := &elkv1alpha1.ElasticsearchWatcher{}
	data := map[string]any{}

	testCase := test.NewTestCase(t.T(), t.k8sClient, key, watch, 5*time.Second, data)
	testCase.Steps = []test.TestStep{
		doCreateWatcherStep(),
		doUpdateWatcherStep(),
		doDeleteWatcherStep(),
	}
	testCase.PreTest = doMockWatcher(t.mockElasticsearchHandler)

	testCase.Run()
}

func doMockWatcher(mockES *mocks.MockElasticsearchHandler) func(stepName *string, data map[string]any) error {
	return func(stepName *string, data map[string]any) (err error) {
		isCreated := false
		isUpdated := false

		mockES.EXPECT().WatchGet(gomock.Any()).AnyTimes().DoAndReturn(func(name string) (*olivere.XPackWatch, error) {

			switch *stepName {
			case "create":
				if !isCreated {
					return nil, nil
				} else {
					resp := &olivere.XPackWatch{
						Trigger: map[string]map[string]any{
							"schedule": {
								"cron": "0 0/1 * * * ?",
							},
						},
						Input: map[string]map[string]any{
							"search": {
								"request": "fake",
							},
						},
						Condition: map[string]map[string]any{
							"compare": {
								"ctx.payload.hits.total": "fake",
							},
						},
						Actions: map[string]map[string]any{
							"email": {
								"email": "fake",
							},
						},
					}
					return resp, nil
				}
			case "update":
				if !isUpdated {
					resp := &olivere.XPackWatch{
						Trigger: map[string]map[string]any{
							"schedule": {
								"cron": "0 0/1 * * * ?",
							},
						},
						Input: map[string]map[string]any{
							"search": {
								"request": "fake",
							},
						},
						Condition: map[string]map[string]any{
							"compare": {
								"ctx.payload.hits.total": "fake",
							},
						},
						Actions: map[string]map[string]any{
							"email": {
								"email": "fake",
							},
						},
					}
					return resp, nil
				} else {
					resp := &olivere.XPackWatch{
						Trigger: map[string]map[string]any{
							"schedule": {
								"cron": "0 0/1 * * * ?",
							},
						},
						Input: map[string]map[string]any{
							"search": {
								"request": "fake",
							},
						},
						Condition: map[string]map[string]any{
							"compare": {
								"ctx.payload.hits.total": "fake",
							},
						},
						Actions: map[string]map[string]any{
							"email": {
								"email": "fake2",
							},
						},
					}
					return resp, nil
				}
			}

			return nil, nil
		})

		mockES.EXPECT().WatchDiff(gomock.Any(), gomock.Any()).AnyTimes().DoAndReturn(func(actual, expected *olivere.XPackWatch) (string, error) {
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

		mockES.EXPECT().WatchUpdate(gomock.Any(), gomock.Any()).AnyTimes().DoAndReturn(func(name string, policy *olivere.XPackWatch) error {
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

		mockES.EXPECT().WatchDelete(gomock.Any()).AnyTimes().DoAndReturn(func(name string) error {
			data["isDeleted"] = true
			return nil
		})

		return nil
	}
}

func doCreateWatcherStep() test.TestStep {
	return test.TestStep{
		Name: "create",
		Do: func(c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			logrus.Infof("=== Add new watch %s/%s ===", key.Namespace, key.Name)

			watch := &elkv1alpha1.ElasticsearchWatcher{
				ObjectMeta: metav1.ObjectMeta{
					Name:      key.Name,
					Namespace: key.Namespace,
				},
				Spec: elkv1alpha1.ElasticsearchWatcherSpec{
					ElasticsearchRefSpec: elkv1alpha1.ElasticsearchRefSpec{
						Name: "test",
					},
					Trigger: `
					{
						"schedule" : { "cron" : "0 0/1 * * * ?" }
					}
					`,
					Input: `
					{
						"search" : {
						  "request" : "fake"
						}
					}
					`,
					Condition: `
					{
						"compare" : { "ctx.payload.hits.total" : "fake"}
					}
					`,
					Actions: `
					{
						"email_admin" : {
						  "email" : "fake"
						}
					}
					`,
				},
			}

			if err = c.Create(context.Background(), watch); err != nil {
				return err
			}

			return nil
		},
		Check: func(t *testing.T, c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			watch := &elkv1alpha1.ElasticsearchWatcher{}
			isCreated := true

			isTimeout, err := RunWithTimeout(func() error {
				if err := c.Get(context.Background(), key, watch); err != nil {
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
				t.Fatalf("Failed to get Watch: %s", err.Error())
			}
			assert.True(t, condition.IsStatusConditionPresentAndEqual(watch.Status.Conditions, watchCondition, metav1.ConditionTrue))
			return nil
		},
	}
}

func doUpdateWatcherStep() test.TestStep {
	return test.TestStep{
		Name: "update",
		Do: func(c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			logrus.Infof("=== Update watch %s/%s ===", key.Namespace, key.Name)

			if o == nil {
				return errors.New("Watch is null")
			}
			watch := o.(*elkv1alpha1.ElasticsearchWatcher)

			watch.Spec.Actions = `
			{
				"email_admin" : {
				"email" : "fake2"
				}
			}
			`
			if err = c.Update(context.Background(), watch); err != nil {
				return err
			}

			return nil
		},
		Check: func(t *testing.T, c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			watch := &elkv1alpha1.ElasticsearchWatcher{}
			isUpdated := true

			isTimeout, err := RunWithTimeout(func() error {
				if err := c.Get(context.Background(), key, watch); err != nil {
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
				t.Fatalf("Failed to get Watch: %s", err.Error())
			}
			assert.True(t, condition.IsStatusConditionPresentAndEqual(watch.Status.Conditions, watchCondition, metav1.ConditionTrue))
			time.Sleep(10 * time.Second)

			return nil
		},
	}
}

func doDeleteWatcherStep() test.TestStep {
	return test.TestStep{
		Name: "delete",
		Do: func(c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			logrus.Infof("=== Delete watch %s/%s ===", key.Namespace, key.Name)

			if o == nil {
				return errors.New("Watch is null")
			}
			watch := o.(*elkv1alpha1.ElasticsearchWatcher)

			wait := int64(0)

			if err = c.Delete(context.Background(), watch, &client.DeleteOptions{GracePeriodSeconds: &wait}); err != nil {
				return err
			}

			return nil
		},
		Check: func(t *testing.T, c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			watch := &elkv1alpha1.ElasticsearchWatcher{}
			isDeleted := true

			isTimeout, err := RunWithTimeout(func() error {
				if err = c.Get(context.Background(), key, watch); err != nil {
					if k8serrors.IsNotFound(err) {
						isDeleted = true
						return nil
					}
					t.Fatal(err)
				}

				return errors.New("Not yet deleted")
			}, time.Second*30, time.Second*1)
			if err != nil || isTimeout {
				t.Fatalf("Watch stil exist: %s", err.Error())
			}
			assert.True(t, isDeleted)

			return nil
		},
	}
}

/*




	// When add new watch










	// When remove watch




}
*/
