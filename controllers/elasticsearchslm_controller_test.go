package controllers

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	elkv1alpha1 "github.com/disaster37/operator-elk-extra/api/v1alpha1"
	"github.com/disaster37/operator-elk-extra/pkg/elasticsearchhandler"
	"github.com/disaster37/operator-elk-extra/pkg/helpers"
	"github.com/disaster37/operator-elk-extra/pkg/mocks"
	"github.com/disaster37/operator-sdk-extra/pkg/test"
	"github.com/golang/mock/gomock"
	olivere "github.com/olivere/elastic/v7"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/controller-runtime/pkg/client"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	condition "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func (t *ControllerTestSuite) TestElasticsearchSLMReconciler() {
	key := types.NamespacedName{
		Name:      "t-slm-" + helpers.RandomString(10),
		Namespace: "default",
	}
	slm := &elkv1alpha1.ElasticsearchSLM{}
	data := map[string]any{}

	testCase := test.NewTestCase(t.T(), t.k8sClient, key, slm, 5*time.Second, data)
	testCase.Steps = []test.TestStep{
		doCreateSLMStep(),
		doUpdateSLMStep(),
		doDeleteSLMStep(),
	}
	testCase.PreTest = doMockSLM(t.mockElasticsearchHandler)

	testCase.Run()
}

func doMockSLM(mockES *mocks.MockElasticsearchHandler) func(stepName *string, data map[string]any) error {
	return func(stepName *string, data map[string]any) (err error) {
		isCreated := false
		isUpdated := false

		mockES.EXPECT().SnapshotRepositoryGet(gomock.Any()).AnyTimes().Return(&olivere.SnapshotRepositoryMetaData{
			Type: "url",
		}, nil)

		mockES.EXPECT().SLMGet(gomock.Any()).AnyTimes().DoAndReturn(func(name string) (*elasticsearchhandler.SnapshotLifecyclePolicySpec, error) {

			switch *stepName {
			case "create":
				if !isCreated {
					return nil, nil
				} else {
					rawPolicy := `
					{
						"schedule": "0 30 1 * * ?",
						"name": "<daily-snap-{now/d}>",
						"repository": "my_repository",
						"config": {
						  "indices": ["data-*", "important"],
						  "ignore_unavailable": false,
						  "include_global_state": false
						},
						"retention": {
						  "expire_after": "30d",
						  "min_count": 5,
						  "max_count": 50
						}
					}`
					resp := &elasticsearchhandler.SnapshotLifecyclePolicySpec{}
					if err := json.Unmarshal([]byte(rawPolicy), resp); err != nil {
						panic(err)
					}
					return resp, nil
				}
			case "update":
				if !isUpdated {
					rawPolicy := `
					{
						"schedule": "0 30 1 * * ?",
						"name": "<daily-snap-{now/d}>",
						"repository": "my_repository",
						"config": {
						  "indices": ["data-*", "important"],
						  "ignore_unavailable": false,
						  "include_global_state": false
						},
						"retention": {
						  "expire_after": "30d",
						  "min_count": 5,
						  "max_count": 50
						}
					}`
					resp := &elasticsearchhandler.SnapshotLifecyclePolicySpec{}
					if err := json.Unmarshal([]byte(rawPolicy), resp); err != nil {
						panic(err)
					}
					return resp, nil
				} else {
					rawPolicy := `
					{
						"schedule": "0 30 1 * * ?",
						"name": "<daily-snap-{now/d}>",
						"repository": "my_repository",
						"config": {
						  "indices": ["data-*", "important"],
						  "ignore_unavailable": false,
						  "include_global_state": false
						},
						"retention": {
						  "expire_after": "30d",
						  "min_count": 6,
						  "max_count": 50
						}
					}`
					resp := &elasticsearchhandler.SnapshotLifecyclePolicySpec{}
					if err := json.Unmarshal([]byte(rawPolicy), resp); err != nil {
						panic(err)
					}
					return resp, nil
				}
			}

			return nil, nil
		})

		mockES.EXPECT().SLMDiff(gomock.Any(), gomock.Any()).AnyTimes().DoAndReturn(func(actual, expected *elasticsearchhandler.SnapshotLifecyclePolicySpec) (string, error) {
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

		mockES.EXPECT().SLMUpdate(gomock.Any(), gomock.Any()).AnyTimes().DoAndReturn(func(name string, policy *elasticsearchhandler.SnapshotLifecyclePolicySpec) error {
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

		mockES.EXPECT().SLMDelete(gomock.Any()).AnyTimes().DoAndReturn(func(name string) error {
			data["isDeleted"] = true
			return nil
		})

		return nil
	}
}

func doCreateSLMStep() test.TestStep {
	return test.TestStep{
		Name: "create",
		Do: func(c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			logrus.Infof("=== Add new SLM policy %s/%s ===", key.Namespace, key.Name)

			slm := &elkv1alpha1.ElasticsearchSLM{
				ObjectMeta: metav1.ObjectMeta{
					Name:      key.Name,
					Namespace: key.Namespace,
				},
				Spec: elkv1alpha1.ElasticsearchSLMSpec{
					ElasticsearchRefSpec: elkv1alpha1.ElasticsearchRefSpec{
						Name: "test",
					},
					Schedule:   "0 30 1 * * ?",
					Name:       "<daily-snap-{now/d}>",
					Repository: "my_repository",
					Config: elkv1alpha1.ElasticsearchSLMConfig{
						Indices:            []string{"data-*", "important"},
						IgnoreUnavailable:  false,
						IncludeGlobalState: false,
					},
					Retention: &elkv1alpha1.ElasticsearchSLMRetention{
						ExpireAfter: "30d",
						MinCount:    5,
						MaxCount:    50,
					},
				},
			}
			if err = c.Create(context.Background(), slm); err != nil {
				return err
			}

			return nil
		},
		Check: func(t *testing.T, c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			slm := &elkv1alpha1.ElasticsearchSLM{}
			isCreated := true

			isTimeout, err := RunWithTimeout(func() error {
				if err := c.Get(context.Background(), key, slm); err != nil {
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
				t.Fatalf("Failed to get SLM: %s", err.Error())
			}
			assert.True(t, condition.IsStatusConditionPresentAndEqual(slm.Status.Conditions, slmCondition, metav1.ConditionTrue))

			return nil
		},
	}
}

func doUpdateSLMStep() test.TestStep {
	return test.TestStep{
		Name: "update",
		Do: func(c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			logrus.Infof("=== Update SLM policy %s/%s ===", key.Namespace, key.Name)

			if o == nil {
				return errors.New("SLM is null")
			}
			slm := o.(*elkv1alpha1.ElasticsearchSLM)

			slm.Spec.Retention.MinCount = 6
			if err = c.Update(context.Background(), slm); err != nil {
				return err
			}

			return nil
		},
		Check: func(t *testing.T, c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			slm := &elkv1alpha1.ElasticsearchSLM{}
			isUpdated := true

			isTimeout, err := RunWithTimeout(func() error {
				if err := c.Get(context.Background(), key, slm); err != nil {
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
				t.Fatalf("Failed to get SLM: %s", err.Error())
			}
			assert.True(t, condition.IsStatusConditionPresentAndEqual(slm.Status.Conditions, slmCondition, metav1.ConditionTrue))

			return nil
		},
	}
}

func doDeleteSLMStep() test.TestStep {
	return test.TestStep{
		Name: "delete",
		Do: func(c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			logrus.Infof("=== Delete SLM policy %s/%s ===", key.Namespace, key.Name)

			if o == nil {
				return errors.New("SLM is null")
			}
			slm := o.(*elkv1alpha1.ElasticsearchSLM)

			wait := int64(0)
			if err = c.Delete(context.Background(), slm, &client.DeleteOptions{GracePeriodSeconds: &wait}); err != nil {
				return err
			}
			return nil
		},
		Check: func(t *testing.T, c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			slm := &elkv1alpha1.ElasticsearchSLM{}
			isDeleted := true

			isTimeout, err := RunWithTimeout(func() error {
				if err = c.Get(context.Background(), key, slm); err != nil {
					if k8serrors.IsNotFound(err) {
						isDeleted = true
						return nil
					}
					t.Fatal(err)
				}

				return errors.New("Not yet deleted")
			}, time.Second*30, time.Second*1)
			if err != nil || isTimeout {
				t.Fatalf("SLM stil exist: %s", err.Error())
			}
			assert.True(t, isDeleted)

			return nil
		},
	}
}
