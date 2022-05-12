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

func (t *ControllerTestSuite) TestRoleMappingReconciler() {
	key := types.NamespacedName{
		Name:      "t-role-mapping-" + helpers.RandomString(10),
		Namespace: "default",
	}
	rm := &elkv1alpha1.RoleMapping{}
	data := map[string]any{}

	testCase := test.NewTestCase(t.T(), t.k8sClient, key, rm, 5*time.Second, data)
	testCase.Steps = []test.TestStep{
		doCreateRoleMappingStep(),
		doUpdateRoleMappingStep(),
		doDeleteRoleMappingStep(),
	}
	testCase.PreTest = doMockRoleMapping(t.mockElasticsearchHandler)

	testCase.Run()
}

func doMockRoleMapping(mockES *mocks.MockElasticsearchHandler) func(stepName *string, data map[string]any) error {
	return func(stepName *string, data map[string]any) (err error) {
		isCreated := false
		isUpdated := false

		mockES.EXPECT().RoleMappingGet(gomock.Any()).AnyTimes().DoAndReturn(func(name string) (*olivere.XPackSecurityRoleMapping, error) {

			switch *stepName {
			case "create":
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
			case "update":
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

		mockES.EXPECT().RoleMappingDiff(gomock.Any(), gomock.Any()).AnyTimes().DoAndReturn(func(actual, expected *olivere.XPackSecurityRoleMapping) (string, error) {
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

		mockES.EXPECT().RoleMappingUpdate(gomock.Any(), gomock.Any()).AnyTimes().DoAndReturn(func(name string, policy *olivere.XPackSecurityRoleMapping) error {
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

		mockES.EXPECT().RoleMappingDelete(gomock.Any()).AnyTimes().DoAndReturn(func(name string) error {
			data["isDeleted"] = true
			return nil
		})

		return nil
	}
}

func doCreateRoleMappingStep() test.TestStep {
	return test.TestStep{
		Name: "create",
		Do: func(c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			logrus.Infof("=== Add new role mapping %s/%s ===", key.Namespace, key.Name)

			rm := &elkv1alpha1.RoleMapping{
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
						"foo": "bar"
					}`,
				},
			}
			if err = c.Create(context.Background(), rm); err != nil {
				return err
			}

			return nil
		},
		Check: func(t *testing.T, c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			rm := &elkv1alpha1.RoleMapping{}
			isCreated := false

			isTimeout, err := RunWithTimeout(func() error {
				if err := c.Get(context.Background(), key, rm); err != nil {
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
				t.Fatalf("Failed to get role mapping: %s", err.Error())
			}
			assert.True(t, condition.IsStatusConditionPresentAndEqual(rm.Status.Conditions, roleMappingCondition, metav1.ConditionTrue))
			return nil
		},
	}
}

func doUpdateRoleMappingStep() test.TestStep {
	return test.TestStep{
		Name: "update",
		Do: func(c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			logrus.Infof("=== Update role mapping %s/%s ===", key.Namespace, key.Name)

			if o == nil {
				return errors.New("Role mapping is null")
			}
			rm := o.(*elkv1alpha1.RoleMapping)

			rm.Spec.Enabled = false
			if err = c.Update(context.Background(), rm); err != nil {
				return err
			}

			return nil
		},
		Check: func(t *testing.T, c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			rm := &elkv1alpha1.RoleMapping{}
			isUpdated := false

			isTimeout, err := RunWithTimeout(func() error {
				if err := c.Get(context.Background(), key, rm); err != nil {
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
				t.Fatalf("Failed to get role mapping: %s", err.Error())
			}
			assert.True(t, condition.IsStatusConditionPresentAndEqual(rm.Status.Conditions, roleMappingCondition, metav1.ConditionTrue))

			return nil
		},
	}
}

func doDeleteRoleMappingStep() test.TestStep {
	return test.TestStep{
		Name: "delete",
		Do: func(c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			logrus.Infof("=== Delete role mapping %s/%s ===", key.Namespace, key.Name)

			if o == nil {
				return errors.New("Role mapping is null")
			}
			rm := o.(*elkv1alpha1.RoleMapping)

			wait := int64(0)
			if err = c.Delete(context.Background(), rm, &client.DeleteOptions{GracePeriodSeconds: &wait}); err != nil {
				return err
			}

			return nil
		},
		Check: func(t *testing.T, c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			rm := &elkv1alpha1.RoleMapping{}
			isDeleted := false

			isTimeout, err := RunWithTimeout(func() error {
				if err = c.Get(context.Background(), key, rm); err != nil {
					if k8serrors.IsNotFound(err) {
						isDeleted = true
						return nil
					}
					t.Fatal(err)
				}

				return errors.New("Not yet deleted")
			}, time.Second*30, time.Second*1)
			if err != nil || isTimeout {
				t.Fatalf("Role mapping stil exist: %s", err.Error())
			}
			assert.True(t, isDeleted)

			return nil
		},
	}
}
