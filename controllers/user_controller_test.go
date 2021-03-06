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

func (t *ControllerTestSuite) TestUserReconciler() {
	key := types.NamespacedName{
		Name:      "t-user-" + helpers.RandomString(10),
		Namespace: "default",
	}
	user := &elkv1alpha1.User{}
	data := map[string]any{}

	testCase := test.NewTestCase(t.T(), t.k8sClient, key, user, 5*time.Second, data)
	testCase.Steps = []test.TestStep{
		doCreateUserStep(),
		doUpdateUserStep(),
		doUpdateUserPasswordHashStep(),
		doDeleteUserStep(),
	}
	testCase.PreTest = doMockUser(t.mockElasticsearchHandler)

	testCase.Run()
}

func doMockUser(mockES *mocks.MockElasticsearchHandler) func(stepName *string, data map[string]any) error {
	return func(stepName *string, data map[string]any) (err error) {
		isCreated := false
		isUpdated := false
		isUpdatedPasswordHash := false

		mockES.EXPECT().UserGet(gomock.Any()).AnyTimes().DoAndReturn(func(name string) (*olivere.XPackSecurityUser, error) {
			switch *stepName {
			case "create":
				if !isCreated {
					return nil, nil
				} else {
					resp := &olivere.XPackSecurityUser{
						Enabled: true,
						Roles:   []string{"superuser"},
					}
					return resp, nil
				}
			case "update":
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
			case "update_password_hash":
				resp := &olivere.XPackSecurityUser{
					Enabled: false,
					Roles:   []string{"superuser"},
				}
				return resp, nil

			}

			return nil, nil
		})

		mockES.EXPECT().UserDiff(gomock.Any(), gomock.Any()).AnyTimes().DoAndReturn(func(actual, expected *olivere.XPackSecurityPutUserRequest) (string, error) {
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
			case "update_password_hash":
				if !isUpdatedPasswordHash {
					return "fake change", nil
				} else {
					return "", nil
				}
			}

			return "", nil
		})

		mockES.EXPECT().UserUpdate(gomock.Any(), gomock.Any()).AnyTimes().DoAndReturn(func(name string, policy *olivere.XPackSecurityPutUserRequest) error {
			switch *stepName {
			case "update":
				isUpdated = true
				data["isUpdated"] = true
				return nil
			case "update_password_hash":
				isUpdatedPasswordHash = true
				data["isUpdatedPasswordHash"] = true
				return nil
			}

			return nil
		})

		mockES.EXPECT().UserCreate(gomock.Any(), gomock.Any()).AnyTimes().DoAndReturn(func(name string, policy *olivere.XPackSecurityPutUserRequest) error {
			isCreated = true
			data["isCreated"] = true

			return nil
		})

		mockES.EXPECT().UserDelete(gomock.Any()).AnyTimes().DoAndReturn(func(name string) error {
			data["isDeleted"] = true
			return nil
		})

		return nil
	}
}

func doCreateUserStep() test.TestStep {
	return test.TestStep{
		Name: "create",
		Do: func(c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			logrus.Infof("=== Add new user %s/%s ===", key.Namespace, key.Name)

			user := &elkv1alpha1.User{
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
			if err = c.Create(context.Background(), user); err != nil {
				return err
			}

			return nil
		},
		Check: func(t *testing.T, c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			user := &elkv1alpha1.User{}
			isCreated := false

			isTimeout, err := RunWithTimeout(func() error {
				if err := c.Get(context.Background(), key, user); err != nil {
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
				t.Fatalf("Failed to get user: %s", err.Error())
			}
			assert.True(t, condition.IsStatusConditionPresentAndEqual(user.Status.Conditions, userCondition, metav1.ConditionTrue))

			return nil
		},
	}
}

func doUpdateUserStep() test.TestStep {
	return test.TestStep{
		Name: "update",
		Do: func(c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			logrus.Infof("=== Update user %s/%s ===", key.Namespace, key.Name)

			if o == nil {
				return errors.New("User is null")
			}
			user := o.(*elkv1alpha1.User)

			user.Spec.Enabled = false
			if err = c.Update(context.Background(), user); err != nil {
				return err
			}

			return nil
		},
		Check: func(t *testing.T, c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			user := &elkv1alpha1.User{}
			isUpdated := false

			isTimeout, err := RunWithTimeout(func() error {
				if err := c.Get(context.Background(), key, user); err != nil {
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
				t.Fatalf("Failed to get User: %s", err.Error())
			}
			assert.True(t, condition.IsStatusConditionPresentAndEqual(user.Status.Conditions, userCondition, metav1.ConditionTrue))

			return nil
		},
	}
}

func doUpdateUserPasswordHashStep() test.TestStep {
	return test.TestStep{
		Name: "update_password_hash",
		Do: func(c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			logrus.Infof("=== Update user (password hash) %s/%s ===", key.Namespace, key.Name)

			if o == nil {
				return errors.New("User is null")
			}
			user := o.(*elkv1alpha1.User)

			user.Spec.PasswordHash = "test2"
			if err = c.Update(context.Background(), user); err != nil {
				return err
			}

			return nil
		},
		Check: func(t *testing.T, c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			user := &elkv1alpha1.User{}
			isUpdated := false

			isTimeout, err := RunWithTimeout(func() error {
				if err := c.Get(context.Background(), key, user); err != nil {
					t.Fatal(err)
				}
				if b, ok := data["isUpdatedPasswordHash"]; ok {
					isUpdated = b.(bool)
				}
				if !isUpdated {
					return errors.New("Not yet updated")
				}
				return nil
			}, time.Second*30, time.Second*1)
			if err != nil || isTimeout {
				t.Fatalf("Failed to get User: %s", err.Error())
			}
			assert.True(t, condition.IsStatusConditionPresentAndEqual(user.Status.Conditions, userCondition, metav1.ConditionTrue))

			return nil
		},
	}
}

func doDeleteUserStep() test.TestStep {
	return test.TestStep{
		Name: "delete",
		Do: func(c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			logrus.Infof("=== Delete user %s/%s ===", key.Namespace, key.Name)

			if o == nil {
				return errors.New("User is null")
			}
			user := o.(*elkv1alpha1.User)

			wait := int64(0)
			if err = c.Delete(context.Background(), user, &client.DeleteOptions{GracePeriodSeconds: &wait}); err != nil {
				return err
			}

			return nil
		},
		Check: func(t *testing.T, c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			user := &elkv1alpha1.User{}
			isDeleted := false

			isTimeout, err := RunWithTimeout(func() error {
				if err = c.Get(context.Background(), key, user); err != nil {
					if k8serrors.IsNotFound(err) {
						isDeleted = true
						return nil
					}
					t.Fatal(err)
				}

				return errors.New("Not yet deleted")
			}, time.Second*30, time.Second*1)
			if err != nil || isTimeout {
				t.Fatalf("user stil exist: %s", err.Error())
			}
			assert.True(t, isDeleted)

			return nil
		},
	}
}
