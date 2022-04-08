package controllers

import (
	"context"
	"time"

	elkv1alpha1 "github.com/disaster37/operator-elk-extra/api/v1alpha1"
	"github.com/disaster37/operator-elk-extra/pkg/elasticsearchhandler"
	"github.com/disaster37/operator-elk-extra/pkg/helpers"
	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	condition "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (t *ControllerTestSuite) TestElasticsearchRoleReconciler() {

	var (
		err       error
		isTimeout bool
		fetched   *elkv1alpha1.ElasticsearchRole
		test      string
		isCreated bool
		isDeleted bool
		isUpdated bool
	)

	roleName := "t-es-role-" + helpers.RandomString(10)
	key := types.NamespacedName{
		Name:      roleName,
		Namespace: "default",
	}
	t.mockElasticsearchHandler.EXPECT().RoleGet(gomock.Any()).AnyTimes().DoAndReturn(func(name string) (*elasticsearchhandler.XPackSecurityRole, error) {

		switch test {
		case "no_role":
			if !isCreated {
				return nil, nil
			} else {

				resp := &elasticsearchhandler.XPackSecurityRole{
					RunAs: []string{"test"},
				}
				return resp, nil
			}
		case "role_update":
			if !isUpdated {
				resp := &elasticsearchhandler.XPackSecurityRole{
					RunAs: []string{"test"},
				}
				return resp, nil
			} else {
				resp := &elasticsearchhandler.XPackSecurityRole{
					RunAs: []string{"test2"},
				}
				return resp, nil
			}
		}

		return nil, nil

	})
	t.mockElasticsearchHandler.EXPECT().RoleDiff(gomock.Any(), gomock.Any()).AnyTimes().DoAndReturn(func(actual, expected *elasticsearchhandler.XPackSecurityRole) (string, error) {
		switch test {
		case "no_role":
			if !isCreated {
				return "fake change", nil
			} else {
				return "", nil
			}
		case "role_update":
			if !isUpdated {
				return "fake change", nil
			} else {
				return "", nil
			}
		}

		return "", nil

	})
	t.mockElasticsearchHandler.EXPECT().RoleUpdate(gomock.Any(), gomock.Any()).AnyTimes().DoAndReturn(func(name string, policy *elasticsearchhandler.XPackSecurityRole) error {
		switch test {
		case "no_role":
			isCreated = true
			return nil
		case "role_update":
			isUpdated = true
			return nil
		}

		return nil
	})
	t.mockElasticsearchHandler.EXPECT().RoleDelete(gomock.Any()).AnyTimes().DoAndReturn(func(name string) error {
		switch test {
		case "role_delete":
			isDeleted = true
			return nil
		}

		return nil
	})

	// When add new role
	logrus.Info("==================================== When add new elasticsearch role")
	test = "no_role"
	toCreate := &elkv1alpha1.ElasticsearchRole{
		ObjectMeta: metav1.ObjectMeta{
			Name:      key.Name,
			Namespace: key.Namespace,
		},
		Spec: elkv1alpha1.ElasticsearchRoleSpec{
			ElasticsearchRefSpec: elkv1alpha1.ElasticsearchRefSpec{
				Name: "test",
			},
			RunAs: []string{"test"},
		},
	}
	if err = t.k8sClient.Create(context.Background(), toCreate); err != nil {
		t.T().Fatal(err)
	}

	isTimeout, err = RunWithTimeout(func() error {
		fetched = &elkv1alpha1.ElasticsearchRole{}
		if err := t.k8sClient.Get(context.Background(), key, fetched); err != nil {
			t.T().Fatal(err)
		}
		if !isCreated {
			return errors.New("Not yet created")
		}
		return nil
	}, time.Second*30, time.Second*1)
	if err != nil || isTimeout {
		t.T().Fatalf("Failed to get elasticsearch role: %s", err.Error())
	}
	assert.True(t.T(), condition.IsStatusConditionPresentAndEqual(fetched.Status.Conditions, elasticsearchRoleCondition, metav1.ConditionTrue))
	time.Sleep(10 * time.Second)

	// When update role
	logrus.Info("==================================== When update elasticsearch role")
	test = "role_update"
	isUpdated = false
	fetched = &elkv1alpha1.ElasticsearchRole{}
	if err := t.k8sClient.Get(context.Background(), key, fetched); err != nil {
		t.T().Fatal(err)
	}
	fetched.Spec.RunAs = []string{"test2"}
	if err = t.k8sClient.Update(context.Background(), fetched); err != nil {
		t.T().Fatal(err)
	}

	isTimeout, err = RunWithTimeout(func() error {
		fetched = &elkv1alpha1.ElasticsearchRole{}
		if err := t.k8sClient.Get(context.Background(), key, fetched); err != nil {
			t.T().Fatal(err)
		}
		if !isUpdated {
			return errors.New("Not yet updated")
		}
		return nil
	}, time.Second*30, time.Second*1)
	if err != nil || isTimeout {
		t.T().Fatalf("Failed to get elasticsearch role: %s", err.Error())
	}
	assert.True(t.T(), condition.IsStatusConditionPresentAndEqual(fetched.Status.Conditions, elasticsearchRoleCondition, metav1.ConditionTrue))
	time.Sleep(10 * time.Second)

	// When remove role
	logrus.Info("==================================== When remove role")
	test = "role_delete"
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
		t.T().Fatalf("Elasticsearch role stil exist: %s", err.Error())
	}
	assert.True(t.T(), isDeleted)
	time.Sleep(10 * time.Second)

}
