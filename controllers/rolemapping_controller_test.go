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

func (t *ControllerTestSuite) TestRoleMappingReconciler() {

	var (
		err       error
		isTimeout bool
		fetched   *elkv1alpha1.RoleMapping
		test      string
		isCreated bool
		isDeleted bool
		isUpdated bool
	)

	roleMappingName := "t-role-mapping-" + helpers.RandomString(10)
	key := types.NamespacedName{
		Name:      roleMappingName,
		Namespace: "default",
	}
	t.mockElasticsearchHandler.EXPECT().RoleMappingGet(gomock.Any()).AnyTimes().DoAndReturn(func(name string) (*olivere.XPackSecurityRoleMapping, error) {

		switch test {
		case "no_role_mapping":
			if !isCreated {
				return nil, nil
			} else {
				resp := &olivere.XPackSecurityRoleMapping{
					Enabled: true,
					Roles:   []string{"superuser"},
					Rules: map[string]any{
						"foo": "bar",
					},
				}
				return resp, nil
			}
		case "role_mapping_update":
			if !isUpdated {
				resp := &olivere.XPackSecurityRoleMapping{
					Enabled: true,
					Roles:   []string{"superuser"},
					Rules: map[string]any{
						"foo": "bar",
					},
				}
				return resp, nil
			} else {
				resp := &olivere.XPackSecurityRoleMapping{
					Enabled: false,
					Roles:   []string{"superuser"},
					Rules: map[string]any{
						"foo": "bar",
					},
				}
				return resp, nil
			}
		}

		return nil, nil

	})
	t.mockElasticsearchHandler.EXPECT().RoleMappingDiff(gomock.Any(), gomock.Any()).AnyTimes().DoAndReturn(func(actual, expected *olivere.XPackSecurityRoleMapping) (string, error) {
		switch test {
		case "no_role_mapping":
			if !isCreated {
				return "fake change", nil
			} else {
				return "", nil
			}
		case "role_mapping_update":
			if !isUpdated {
				return "fake change", nil
			} else {
				return "", nil
			}
		}

		return "", nil

	})
	t.mockElasticsearchHandler.EXPECT().RoleMappingUpdate(gomock.Any(), gomock.Any()).AnyTimes().DoAndReturn(func(name string, policy *olivere.XPackSecurityRoleMapping) error {
		switch test {
		case "no_role_mapping":
			isCreated = true
			return nil
		case "role_mapping_update":
			isUpdated = true
			return nil
		}

		return nil
	})
	t.mockElasticsearchHandler.EXPECT().RoleMappingDelete(gomock.Any()).AnyTimes().DoAndReturn(func(name string) error {
		switch test {
		case "role_mapping_delete":
			isDeleted = true
			return nil
		}

		return nil
	})

	// When add new role mapping
	logrus.Info("==================================== When add new role mapping")
	test = "no_role_mapping"
	toCreate := &elkv1alpha1.RoleMapping{
		ObjectMeta: metav1.ObjectMeta{
			Name:      key.Name,
			Namespace: key.Namespace,
		},
		Spec: elkv1alpha1.RoleMappingSpec{
			ElasticsearchRefSpec: elkv1alpha1.ElasticsearchRefSpec{
				Name: "test",
			},
			Enabled: true,
			Roles:   []string{"superuser"},
			Rules: `{
				"foo": "bar",
			}`,
		},
	}
	if err = t.k8sClient.Create(context.Background(), toCreate); err != nil {
		t.T().Fatal(err)
	}

	isTimeout, err = RunWithTimeout(func() error {
		fetched = &elkv1alpha1.RoleMapping{}
		if err := t.k8sClient.Get(context.Background(), key, fetched); err != nil {
			t.T().Fatal(err)
		}
		if !isCreated {
			return errors.New("Not yet created")
		}
		return nil
	}, time.Second*30, time.Second*1)
	if err != nil || isTimeout {
		t.T().Fatalf("Failed to get role mapping: %s", err.Error())
	}
	assert.True(t.T(), condition.IsStatusConditionPresentAndEqual(fetched.Status.Conditions, roleMappingCondition, metav1.ConditionTrue))
	time.Sleep(10 * time.Second)

	// When update role mapping
	logrus.Info("==================================== When update role mapping")
	test = "role_mapping_update"
	isUpdated = false
	fetched = &elkv1alpha1.RoleMapping{}
	if err := t.k8sClient.Get(context.Background(), key, fetched); err != nil {
		t.T().Fatal(err)
	}
	fetched.Spec.Enabled = false
	if err = t.k8sClient.Update(context.Background(), fetched); err != nil {
		t.T().Fatal(err)
	}

	isTimeout, err = RunWithTimeout(func() error {
		fetched = &elkv1alpha1.RoleMapping{}
		if err := t.k8sClient.Get(context.Background(), key, fetched); err != nil {
			t.T().Fatal(err)
		}
		if !isUpdated {
			return errors.New("Not yet updated")
		}
		return nil
	}, time.Second*30, time.Second*1)
	if err != nil || isTimeout {
		t.T().Fatalf("Failed to get role mapping: %s", err.Error())
	}
	assert.True(t.T(), condition.IsStatusConditionPresentAndEqual(fetched.Status.Conditions, roleMappingCondition, metav1.ConditionTrue))
	time.Sleep(10 * time.Second)

	// When remove role mapping
	logrus.Info("==================================== When remove role mapping")
	test = "role_mapping_delete"
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
		t.T().Fatalf("Role mapping stil exist: %s", err.Error())
	}
	assert.True(t.T(), isDeleted)
	time.Sleep(10 * time.Second)

}
