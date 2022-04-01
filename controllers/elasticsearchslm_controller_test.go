package controllers

import (
	"context"
	"encoding/json"
	"time"

	elkv1alpha1 "github.com/disaster37/operator-elk-extra/api/v1alpha1"
	"github.com/disaster37/operator-elk-extra/pkg/elasticsearchhandler"
	"github.com/disaster37/operator-elk-extra/pkg/helpers"
	"github.com/golang/mock/gomock"
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

func (t *ControllerTestSuite) TestElasticsearchSLMReconciler() {

	var (
		err       error
		isTimeout bool
		fetched   *elkv1alpha1.ElasticsearchSLM
		test      string
		isCreated bool
		isDeleted bool
		isUpdated bool
	)

	slmName := "t-slm-" + helpers.RandomString(10)
	key := types.NamespacedName{
		Name:      slmName,
		Namespace: "default",
	}
	t.mockElasticsearchHandler.EXPECT().SLMGet(gomock.Any()).AnyTimes().DoAndReturn(func(name string) (*elasticsearchhandler.SnapshotLifecyclePolicySpec, error) {

		switch test {
		case "no_slm":
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
					t.T().Fatal(err)
				}
				return resp, nil
			}
		case "slm_update":
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
					t.T().Fatal(err)
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
					t.T().Fatal(err)
				}
				return resp, nil
			}
		}

		return nil, nil

	})
	t.mockElasticsearchHandler.EXPECT().SLMDiff(gomock.Any(), gomock.Any()).AnyTimes().DoAndReturn(func(actual, expected *elasticsearchhandler.SnapshotLifecyclePolicySpec) (string, error) {
		switch test {
		case "no_slm":
			if !isCreated {
				return "fake change", nil
			} else {
				return "", nil
			}
		case "slm_update":
			if !isUpdated {
				return "fake change", nil
			} else {
				return "", nil
			}
		}

		return "", nil

	})
	t.mockElasticsearchHandler.EXPECT().SLMUpdate(gomock.Any(), gomock.Any()).AnyTimes().DoAndReturn(func(name string, policy *elasticsearchhandler.SnapshotLifecyclePolicySpec) error {
		switch test {
		case "no_slm":
			isCreated = true
			return nil
		case "slm_update":
			isUpdated = true
			return nil
		}

		return nil
	})
	t.mockElasticsearchHandler.EXPECT().SLMDelete(gomock.Any()).AnyTimes().DoAndReturn(func(name string) error {
		switch test {
		case "slm_delete":
			isDeleted = true
			return nil
		}

		return nil
	})

	// When add new SLM policy
	logrus.Info("==================================== When add new SLM policy")
	test = "no_slm"
	toCreate := &elkv1alpha1.ElasticsearchSLM{
		ObjectMeta: metav1.ObjectMeta{
			Name:      key.Name,
			Namespace: key.Namespace,
		},
		Spec: elkv1alpha1.ElasticsearchSLMSpec{
			ElasticsearchRefSpec: elkv1alpha1.ElasticsearchRefSpec{
				Name: "test",
			},
			Policy: `
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
			}`,
		},
	}
	if err = t.k8sClient.Create(context.Background(), toCreate); err != nil {
		t.T().Fatal(err)
	}

	isTimeout, err = RunWithTimeout(func() error {
		fetched = &elkv1alpha1.ElasticsearchSLM{}
		if err := t.k8sClient.Get(context.Background(), key, fetched); err != nil {
			t.T().Fatal(err)
		}
		if !isCreated {
			return errors.New("Not yet created")
		}
		return nil
	}, time.Second*30, time.Second*1)
	if err != nil || isTimeout {
		t.T().Fatalf("Failed to get SLM: %s", err.Error())
	}
	assert.True(t.T(), condition.IsStatusConditionPresentAndEqual(fetched.Status.Conditions, slmCondition, metav1.ConditionTrue))
	time.Sleep(10 * time.Second)

	// When update SLM policy
	logrus.Info("==================================== When update SLM policy")
	test = "slm_update"
	isUpdated = false
	fetched = &elkv1alpha1.ElasticsearchSLM{}
	if err := t.k8sClient.Get(context.Background(), key, fetched); err != nil {
		t.T().Fatal(err)
	}
	fetched.Spec.Policy = `{
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
	if err = t.k8sClient.Update(context.Background(), fetched); err != nil {
		t.T().Fatal(err)
	}

	isTimeout, err = RunWithTimeout(func() error {
		fetched = &elkv1alpha1.ElasticsearchSLM{}
		if err := t.k8sClient.Get(context.Background(), key, fetched); err != nil {
			t.T().Fatal(err)
		}
		if !isUpdated {
			return errors.New("Not yet updated")
		}
		return nil
	}, time.Second*30, time.Second*1)
	if err != nil || isTimeout {
		t.T().Fatalf("Failed to get SLM: %s", err.Error())
	}
	assert.True(t.T(), condition.IsStatusConditionPresentAndEqual(fetched.Status.Conditions, slmCondition, metav1.ConditionTrue))
	time.Sleep(10 * time.Second)

	// When remove slm policy
	logrus.Info("==================================== When remove SLM policy")
	test = "slm_delete"
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
		t.T().Fatalf("SLM stil exist: %s", err.Error())
	}
	assert.True(t.T(), isDeleted)
	time.Sleep(10 * time.Second)

}
