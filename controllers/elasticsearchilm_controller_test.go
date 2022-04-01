package controllers

import (
	"context"
	"encoding/json"
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

func (t *ControllerTestSuite) TestElasticsearchILMReconciler() {

	var (
		err       error
		isTimeout bool
		fetched   *elkv1alpha1.ElasticsearchILM
		test      string
		isCreated bool
		isDeleted bool
		isUpdated bool
	)

	ilmName := "t-ilm-" + helpers.RandomString(10)
	key := types.NamespacedName{
		Name:      ilmName,
		Namespace: "default",
	}
	t.mockElasticsearchHandler.EXPECT().ILMGet(gomock.Any()).AnyTimes().DoAndReturn(func(name string) (*olivere.XPackIlmGetLifecycleResponse, error) {

		switch test {
		case "no_ilm":
			if !isCreated {
				return nil, nil
			} else {
				rawPolicy := `
				{
					"policy": {
						"phases": {
							"warm": {
								"min_age": "10d",
								"actions": {
									"forcemerge": {
										"max_num_segments": 1
									}
								}
							},
							"delete": {
								"min_age": "31d",
								"actions": {
									"delete": {
										"delete_searchable_snapshot": true
									}
								}
							}
						}
					}
				}`
				resp := &olivere.XPackIlmGetLifecycleResponse{}
				if err := json.Unmarshal([]byte(rawPolicy), resp); err != nil {
					t.T().Fatal(err)
				}
				return resp, nil
			}
		case "ilm_update":
			if !isUpdated {
				rawPolicy := `
				{
					"policy": {
						"phases": {
							"warm": {
								"min_age": "10d",
								"actions": {
									"forcemerge": {
										"max_num_segments": 1
									}
								}
							},
							"delete": {
								"min_age": "31d",
								"actions": {
									"delete": {
										"delete_searchable_snapshot": true
									}
								}
							}
						}
					}
				}`
				resp := &olivere.XPackIlmGetLifecycleResponse{}
				if err := json.Unmarshal([]byte(rawPolicy), resp); err != nil {
					t.T().Fatal(err)
				}
				return resp, nil
			} else {
				rawPolicy := `
				{
					"policy": {
						"phases": {
							"warm": {
								"min_age": "30d",
								"actions": {
									"forcemerge": {
										"max_num_segments": 1
									}
								}
							},
							"delete": {
								"min_age": "31d",
								"actions": {
									"delete": {
										"delete_searchable_snapshot": true
									}
								}
							}
						}
					}
				}`
				resp := &olivere.XPackIlmGetLifecycleResponse{}
				if err := json.Unmarshal([]byte(rawPolicy), resp); err != nil {
					t.T().Fatal(err)
				}
				return resp, nil
			}
		}

		return nil, nil

	})
	t.mockElasticsearchHandler.EXPECT().ILMDiff(gomock.Any(), gomock.Any()).AnyTimes().DoAndReturn(func(actual, expected *olivere.XPackIlmGetLifecycleResponse) (string, error) {
		switch test {
		case "no_ilm":
			if !isCreated {
				return "fake change", nil
			} else {
				return "", nil
			}
		case "ilm_update":
			if !isUpdated {
				return "fake change", nil
			} else {
				return "", nil
			}
		}

		return "", nil

	})
	t.mockElasticsearchHandler.EXPECT().ILMUpdate(gomock.Any(), gomock.Any()).AnyTimes().DoAndReturn(func(name string, policy *olivere.XPackIlmGetLifecycleResponse) error {
		switch test {
		case "no_ilm":
			isCreated = true
			return nil
		case "ilm_update":
			isUpdated = true
			return nil
		}

		return nil
	})
	t.mockElasticsearchHandler.EXPECT().ILMDelete(gomock.Any()).AnyTimes().DoAndReturn(func(name string) error {
		switch test {
		case "ilm_delete":
			isDeleted = true
			return nil
		}

		return nil
	})

	// When add new ILM policy
	logrus.Info("==================================== When add new ILM policy")
	test = "no_ilm"
	toCreate := &elkv1alpha1.ElasticsearchILM{
		ObjectMeta: metav1.ObjectMeta{
			Name:      key.Name,
			Namespace: key.Namespace,
		},
		Spec: elkv1alpha1.ElasticsearchILMSpec{
			ElasticsearchRefSpec: elkv1alpha1.ElasticsearchRefSpec{
				Name: "test",
			},
			Policy: `
			{
				"policy": {
					"phases": {
						"warm": {
							"min_age": "10d",
							"actions": {
								"forcemerge": {
									"max_num_segments": 1
								}
							}
						},
						"delete": {
							"min_age": "31d",
							"actions": {
								"delete": {
									"delete_searchable_snapshot": true
								}
							}
						}
					}
				}
			}`,
		},
	}
	if err = t.k8sClient.Create(context.Background(), toCreate); err != nil {
		t.T().Fatal(err)
	}

	isTimeout, err = RunWithTimeout(func() error {
		fetched = &elkv1alpha1.ElasticsearchILM{}
		if err := t.k8sClient.Get(context.Background(), key, fetched); err != nil {
			t.T().Fatal(err)
		}
		if !isCreated {
			return errors.New("Not yet created")
		}
		return nil
	}, time.Second*30, time.Second*1)
	if err != nil || isTimeout {
		t.T().Fatalf("Failed to get ILM: %s", err.Error())
	}
	assert.True(t.T(), condition.IsStatusConditionPresentAndEqual(fetched.Status.Conditions, ilmCondition, metav1.ConditionTrue))
	time.Sleep(10 * time.Second)

	// When update ILM policy
	logrus.Info("==================================== When update ILM policy")
	test = "ilm_update"
	isUpdated = false
	fetched = &elkv1alpha1.ElasticsearchILM{}
	if err := t.k8sClient.Get(context.Background(), key, fetched); err != nil {
		t.T().Fatal(err)
	}
	fetched.Spec.Policy = `{
		"policy": {
			"phases": {
				"warm": {
					"min_age": "30d",
					"actions": {
						"forcemerge": {
							"max_num_segments": 1
						}
					}
				},
				"delete": {
					"min_age": "31d",
					"actions": {
						"delete": {
							"delete_searchable_snapshot": true
						}
					}
				}
			}
		}
	}`
	if err = t.k8sClient.Update(context.Background(), fetched); err != nil {
		t.T().Fatal(err)
	}

	isTimeout, err = RunWithTimeout(func() error {
		fetched = &elkv1alpha1.ElasticsearchILM{}
		if err := t.k8sClient.Get(context.Background(), key, fetched); err != nil {
			t.T().Fatal(err)
		}
		if !isUpdated {
			return errors.New("Not yet updated")
		}
		return nil
	}, time.Second*30, time.Second*1)
	if err != nil || isTimeout {
		t.T().Fatalf("Failed to get ILM: %s", err.Error())
	}
	assert.True(t.T(), condition.IsStatusConditionPresentAndEqual(fetched.Status.Conditions, ilmCondition, metav1.ConditionTrue))
	time.Sleep(10 * time.Second)

	// When remove ilm policy
	logrus.Info("==================================== When remove ILM policy")
	test = "ilm_delete"
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
		t.T().Fatalf("ILM stil exist: %s", err.Error())
	}
	assert.True(t.T(), isDeleted)
	time.Sleep(10 * time.Second)

}
