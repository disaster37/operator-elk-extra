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

func (t *ControllerTestSuite) TestUserReconciler() {

	var (
		err       error
		isTimeout bool
		fetched   *elkv1alpha1.User
		test      string
		isCreated bool
		isDeleted bool
		isUpdated bool
	)

	userName := "t-user-" + helpers.RandomString(10)
	key := types.NamespacedName{
		Name:      userName,
		Namespace: "default",
	}
	t.mockElasticsearchHandler.EXPECT().UserGet(gomock.Any()).AnyTimes().DoAndReturn(func(name string) (*olivere.XPackSecurityUser, error) {

		switch test {
		case "no_user":
			if !isCreated {
				return nil, nil
			} else {
				resp := &olivere.XPackSecurityUser{
					Enabled: true,
					Roles:   []string{"superuser"},
				}
				return resp, nil
			}
		case "user_update":
			if !isUpdated {
				resp := &olivere.XPackSecurityUser{
					Enabled: true,
					Roles:   []string{"superuser"},
				}
				return resp, nil
			} else {
				resp := &olivere.XPackSecurityUser{
					Enabled: false,
					Roles:   []string{"superuser"},
				}
				return resp, nil
			}
		case "user_update_password_hash":
			resp := &olivere.XPackSecurityUser{
				Enabled: false,
				Roles:   []string{"superuser"},
			}
			return resp, nil

		}

		return nil, nil

	})
	t.mockElasticsearchHandler.EXPECT().UserDiff(gomock.Any(), gomock.Any()).AnyTimes().DoAndReturn(func(actual, expected *olivere.XPackSecurityPutUserRequest) (string, error) {
		switch test {
		case "no_user":
			if !isCreated {
				return "fake change", nil
			} else {
				return "", nil
			}
		case "user_update":
			if !isUpdated {
				return "fake change", nil
			} else {
				return "", nil
			}
		case "user_update_password_hash":
			if !isUpdated {
				return "fake change", nil
			} else {
				return "", nil
			}
		}

		return "", nil

	})
	t.mockElasticsearchHandler.EXPECT().UserUpdate(gomock.Any(), gomock.Any()).AnyTimes().DoAndReturn(func(name string, policy *olivere.XPackSecurityPutUserRequest) error {
		switch test {
		case "no_user":
			isCreated = true
			return nil
		case "user_update":
			isUpdated = true
			return nil
		case "user_update_password_hash":
			isUpdated = true
			return nil
		}

		return nil
	})
	t.mockElasticsearchHandler.EXPECT().UserCreate(gomock.Any(), gomock.Any()).AnyTimes().DoAndReturn(func(name string, policy *olivere.XPackSecurityPutUserRequest) error {
		switch test {
		case "no_user":
			isCreated = true
			return nil
		case "user_update":
			isUpdated = true
			return nil
		}

		return nil
	})
	t.mockElasticsearchHandler.EXPECT().UserDelete(gomock.Any()).AnyTimes().DoAndReturn(func(name string) error {
		switch test {
		case "user_delete":
			isDeleted = true
			return nil
		}

		return nil
	})

	// When add new user
	logrus.Info("==================================== When add new user")
	test = "no_user"
	toCreate := &elkv1alpha1.User{
		ObjectMeta: metav1.ObjectMeta{
			Name:      key.Name,
			Namespace: key.Namespace,
		},
		Spec: elkv1alpha1.UserSpec{
			ElasticsearchRefSpec: elkv1alpha1.ElasticsearchRefSpec{
				Name: "test",
			},
			Enabled:      true,
			Roles:        []string{"superuser"},
			PasswordHash: "test",
		},
	}
	if err = t.k8sClient.Create(context.Background(), toCreate); err != nil {
		t.T().Fatal(err)
	}

	isTimeout, err = RunWithTimeout(func() error {
		fetched = &elkv1alpha1.User{}
		if err := t.k8sClient.Get(context.Background(), key, fetched); err != nil {
			t.T().Fatal(err)
		}
		if !isCreated {
			return errors.New("Not yet created")
		}
		return nil
	}, time.Second*30, time.Second*1)
	if err != nil || isTimeout {
		t.T().Fatalf("Failed to get user: %s", err.Error())
	}
	assert.True(t.T(), condition.IsStatusConditionPresentAndEqual(fetched.Status.Conditions, userCondition, metav1.ConditionTrue))
	time.Sleep(10 * time.Second)

	// When update user
	logrus.Info("==================================== When update user")
	test = "user_update"
	isUpdated = false
	fetched = &elkv1alpha1.User{}
	if err := t.k8sClient.Get(context.Background(), key, fetched); err != nil {
		t.T().Fatal(err)
	}
	fetched.Spec.Enabled = false
	if err = t.k8sClient.Update(context.Background(), fetched); err != nil {
		t.T().Fatal(err)
	}

	isTimeout, err = RunWithTimeout(func() error {
		fetched = &elkv1alpha1.User{}
		if err := t.k8sClient.Get(context.Background(), key, fetched); err != nil {
			t.T().Fatal(err)
		}
		if !isUpdated {
			return errors.New("Not yet updated")
		}
		return nil
	}, time.Second*30, time.Second*1)
	if err != nil || isTimeout {
		t.T().Fatalf("Failed to get User: %s", err.Error())
	}
	assert.True(t.T(), condition.IsStatusConditionPresentAndEqual(fetched.Status.Conditions, userCondition, metav1.ConditionTrue))
	time.Sleep(10 * time.Second)

	// When change password hash
	logrus.Info("==================================== When update user")
	test = "user_update_password_hash"
	isUpdated = false
	fetched = &elkv1alpha1.User{}
	if err := t.k8sClient.Get(context.Background(), key, fetched); err != nil {
		t.T().Fatal(err)
	}
	fetched.Spec.PasswordHash = "test2"
	if err = t.k8sClient.Update(context.Background(), fetched); err != nil {
		t.T().Fatal(err)
	}

	isTimeout, err = RunWithTimeout(func() error {
		fetched = &elkv1alpha1.User{}
		if err := t.k8sClient.Get(context.Background(), key, fetched); err != nil {
			t.T().Fatal(err)
		}
		if !isUpdated {
			return errors.New("Not yet updated")
		}
		return nil
	}, time.Second*30, time.Second*1)
	if err != nil || isTimeout {
		t.T().Fatalf("Failed to get User: %s", err.Error())
	}
	assert.True(t.T(), condition.IsStatusConditionPresentAndEqual(fetched.Status.Conditions, userCondition, metav1.ConditionTrue))
	time.Sleep(10 * time.Second)

	// When remove user
	logrus.Info("==================================== When remove user")
	test = "user_delete"
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
		t.T().Fatalf("user stil exist: %s", err.Error())
	}
	assert.True(t.T(), isDeleted)
	time.Sleep(10 * time.Second)

}
