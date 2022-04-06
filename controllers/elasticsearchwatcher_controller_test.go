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

func (t *ControllerTestSuite) TestElasticsearchWatcherReconciler() {

	var (
		err       error
		isTimeout bool
		fetched   *elkv1alpha1.ElasticsearchWatcher
		test      string
		isCreated bool
		isDeleted bool
		isUpdated bool
	)

	watchName := "t-watch-" + helpers.RandomString(10)
	key := types.NamespacedName{
		Name:      watchName,
		Namespace: "default",
	}
	t.mockElasticsearchHandler.EXPECT().WatchGet(gomock.Any()).AnyTimes().DoAndReturn(func(name string) (*olivere.XPackWatch, error) {

		switch test {
		case "no_watch":
			if !isCreated {
				return nil, nil
			} else {
				resp := &olivere.XPackWatch{
					Trigger: map[string]map[string]any{
						"schedule": map[string]any{
							"cron": "0 0/1 * * * ?",
						},
					},
					Input: map[string]map[string]any{
						"search": map[string]any{
							"request": "fake",
						},
					},
					Condition: map[string]map[string]any{
						"compare": map[string]any{
							"ctx.payload.hits.total": "fake",
						},
					},
					Actions: map[string]map[string]any{
						"email": map[string]any{
							"email": "fake",
						},
					},
				}
				return resp, nil
			}
		case "watch_update":
			if !isUpdated {
				resp := &olivere.XPackWatch{
					Trigger: map[string]map[string]any{
						"schedule": map[string]any{
							"cron": "0 0/1 * * * ?",
						},
					},
					Input: map[string]map[string]any{
						"search": map[string]any{
							"request": "fake",
						},
					},
					Condition: map[string]map[string]any{
						"compare": map[string]any{
							"ctx.payload.hits.total": "fake",
						},
					},
					Actions: map[string]map[string]any{
						"email": map[string]any{
							"email": "fake",
						},
					},
				}
				return resp, nil
			} else {
				resp := &olivere.XPackWatch{
					Trigger: map[string]map[string]any{
						"schedule": map[string]any{
							"cron": "0 0/1 * * * ?",
						},
					},
					Input: map[string]map[string]any{
						"search": map[string]any{
							"request": "fake",
						},
					},
					Condition: map[string]map[string]any{
						"compare": map[string]any{
							"ctx.payload.hits.total": "fake",
						},
					},
					Actions: map[string]map[string]any{
						"email": map[string]any{
							"email": "fake2",
						},
					},
				}
				return resp, nil
			}
		}

		return nil, nil

	})
	t.mockElasticsearchHandler.EXPECT().WatchDiff(gomock.Any(), gomock.Any()).AnyTimes().DoAndReturn(func(actual, expected *olivere.XPackWatch) (string, error) {
		switch test {
		case "no_watch":
			if !isCreated {
				return "fake change", nil
			} else {
				return "", nil
			}
		case "watch_update":
			if !isUpdated {
				return "fake change", nil
			} else {
				return "", nil
			}
		}

		return "", nil

	})
	t.mockElasticsearchHandler.EXPECT().WatchUpdate(gomock.Any(), gomock.Any()).AnyTimes().DoAndReturn(func(name string, policy *olivere.XPackWatch) error {
		switch test {
		case "no_watch":
			isCreated = true
			return nil
		case "watch_update":
			isUpdated = true
			return nil
		}

		return nil
	})
	t.mockElasticsearchHandler.EXPECT().WatchDelete(gomock.Any()).AnyTimes().DoAndReturn(func(name string) error {
		switch test {
		case "watch_delete":
			isDeleted = true
			return nil
		}

		return nil
	})

	// When add new watch
	logrus.Info("==================================== When add new watch")
	test = "no_watch"
	toCreate := &elkv1alpha1.ElasticsearchWatcher{
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

	if err = t.k8sClient.Create(context.Background(), toCreate); err != nil {
		t.T().Fatal(err)
	}

	isTimeout, err = RunWithTimeout(func() error {
		fetched = &elkv1alpha1.ElasticsearchWatcher{}
		if err := t.k8sClient.Get(context.Background(), key, fetched); err != nil {
			t.T().Fatal(err)
		}
		if !isCreated {
			return errors.New("Not yet created")
		}
		return nil
	}, time.Second*30, time.Second*1)
	if err != nil || isTimeout {
		t.T().Fatalf("Failed to get Watch: %s", err.Error())
	}
	assert.True(t.T(), condition.IsStatusConditionPresentAndEqual(fetched.Status.Conditions, watchCondition, metav1.ConditionTrue))
	time.Sleep(10 * time.Second)

	// When update watch
	logrus.Info("==================================== When update watch")
	test = "watch_update"
	isUpdated = false
	fetched = &elkv1alpha1.ElasticsearchWatcher{}
	if err := t.k8sClient.Get(context.Background(), key, fetched); err != nil {
		t.T().Fatal(err)
	}
	fetched.Spec.Actions = `
	{
		"email_admin" : {
		  "email" : "fake2"
		}
	}
	`
	if err = t.k8sClient.Update(context.Background(), fetched); err != nil {
		t.T().Fatal(err)
	}

	isTimeout, err = RunWithTimeout(func() error {
		fetched = &elkv1alpha1.ElasticsearchWatcher{}
		if err := t.k8sClient.Get(context.Background(), key, fetched); err != nil {
			t.T().Fatal(err)
		}
		if !isUpdated {
			return errors.New("Not yet updated")
		}
		return nil
	}, time.Second*30, time.Second*1)
	if err != nil || isTimeout {
		t.T().Fatalf("Failed to get Watch: %s", err.Error())
	}
	assert.True(t.T(), condition.IsStatusConditionPresentAndEqual(fetched.Status.Conditions, watchCondition, metav1.ConditionTrue))
	time.Sleep(10 * time.Second)

	// When remove watch
	logrus.Info("==================================== When remove watch")
	test = "watch_delete"
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
		t.T().Fatalf("Watch stil exist: %s", err.Error())
	}
	assert.True(t.T(), isDeleted)
	time.Sleep(10 * time.Second)

}
