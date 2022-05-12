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

func (t *ControllerTestSuite) TestElasticsearchSnapshotRepositoryReconciler() {
	key := types.NamespacedName{
		Name:      "t-snapshot-repo-" + helpers.RandomString(10),
		Namespace: "default",
	}
	repo := &elkv1alpha1.ElasticsearchSnapshotRepository{}
	data := map[string]any{}

	testCase := test.NewTestCase(t.T(), t.k8sClient, key, repo, 5*time.Second, data)
	testCase.Steps = []test.TestStep{
		doCreateSnapshotRepoStep(),
		doUpdateSnapshotRepoStep(),
		doDeleteSnapshotRepoStep(),
	}
	testCase.PreTest = doMockSnapshotRepo(t.mockElasticsearchHandler)

	testCase.Run()
}

func doMockSnapshotRepo(mockES *mocks.MockElasticsearchHandler) func(stepName *string, data map[string]any) error {
	return func(stepName *string, data map[string]any) (err error) {
		isCreated := false
		isUpdated := false

		mockES.EXPECT().SnapshotRepositoryGet(gomock.Any()).AnyTimes().DoAndReturn(func(name string) (*olivere.SnapshotRepositoryMetaData, error) {
			switch *stepName {
			case "create":
				if !isCreated {
					return nil, nil
				} else {
					resp := &olivere.SnapshotRepositoryMetaData{
						Type: "url",
						Settings: map[string]any{
							"url": "http://fake",
						},
					}
					return resp, nil
				}
			case "update":
				if !isUpdated {
					resp := &olivere.SnapshotRepositoryMetaData{
						Type: "url",
						Settings: map[string]any{
							"url": "http://fake",
						},
					}
					return resp, nil
				} else {
					resp := &olivere.SnapshotRepositoryMetaData{
						Type: "url",
						Settings: map[string]any{
							"url": "http://fake2",
						},
					}
					return resp, nil
				}
			}

			return nil, nil
		})

		mockES.EXPECT().SnapshotRepositoryDiff(gomock.Any(), gomock.Any()).AnyTimes().DoAndReturn(func(actual, expected *olivere.SnapshotRepositoryMetaData) (string, error) {
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

		mockES.EXPECT().SnapshotRepositoryUpdate(gomock.Any(), gomock.Any()).AnyTimes().DoAndReturn(func(name string, policy *olivere.SnapshotRepositoryMetaData) error {
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

		mockES.EXPECT().SnapshotRepositoryDelete(gomock.Any()).AnyTimes().DoAndReturn(func(name string) error {
			data["isDeleted"] = true
			return nil
		})

		return nil
	}
}

func doCreateSnapshotRepoStep() test.TestStep {
	return test.TestStep{
		Name: "create",
		Do: func(c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			logrus.Infof("=== Add new snapshot repository %s/%s ===", key.Namespace, key.Name)

			repo := &elkv1alpha1.ElasticsearchSnapshotRepository{
				ObjectMeta: metav1.ObjectMeta{
					Name:      key.Name,
					Namespace: key.Namespace,
				},
				Spec: elkv1alpha1.ElasticsearchSnapshotRepositorySpec{
					ElasticsearchRefSpec: elkv1alpha1.ElasticsearchRefSpec{
						Name: "test",
					},
					Type: "url",
					Settings: `
					{
						"url" : "http://fake"
					}
					`,
				},
			}
			if err = c.Create(context.Background(), repo); err != nil {
				return err
			}

			return nil
		},
		Check: func(t *testing.T, c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			repo := &elkv1alpha1.ElasticsearchSnapshotRepository{}
			isCreated := true

			isTimeout, err := RunWithTimeout(func() error {
				if err := c.Get(context.Background(), key, repo); err != nil {
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
				t.Fatalf("Failed to get Snapshot repository: %s", err.Error())
			}
			assert.True(t, condition.IsStatusConditionPresentAndEqual(repo.Status.Conditions, repositoryCondition, metav1.ConditionTrue))

			return nil
		},
	}
}

func doUpdateSnapshotRepoStep() test.TestStep {
	return test.TestStep{
		Name: "update",
		Do: func(c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			logrus.Infof("=== Update snapshot repository %s/%s ===", key.Namespace, key.Name)

			if o == nil {
				return errors.New("Snapshot repo is null")
			}
			repo := o.(*elkv1alpha1.ElasticsearchSnapshotRepository)

			repo.Spec.Settings = `
				{
					"url" : "http://fake2"
				}
			`
			if err = c.Update(context.Background(), repo); err != nil {
				return err
			}

			return nil
		},
		Check: func(t *testing.T, c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			repo := &elkv1alpha1.ElasticsearchSnapshotRepository{}
			isUpdated := true

			isTimeout, err := RunWithTimeout(func() error {
				if err := c.Get(context.Background(), key, repo); err != nil {
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
				t.Fatalf("Failed to get Snapshot repository: %s", err.Error())
			}
			assert.True(t, condition.IsStatusConditionPresentAndEqual(repo.Status.Conditions, repositoryCondition, metav1.ConditionTrue))

			return nil
		},
	}
}

func doDeleteSnapshotRepoStep() test.TestStep {
	return test.TestStep{
		Name: "delete",
		Do: func(c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			logrus.Infof("=== Delete snapshot repository %s/%s ===", key.Namespace, key.Name)

			if o == nil {
				return errors.New("Snapshot repo is null")
			}
			repo := o.(*elkv1alpha1.ElasticsearchSnapshotRepository)

			wait := int64(0)
			if err = c.Delete(context.Background(), repo, &client.DeleteOptions{GracePeriodSeconds: &wait}); err != nil {
				return err
			}

			return nil
		},
		Check: func(t *testing.T, c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			repo := &elkv1alpha1.ElasticsearchSnapshotRepository{}
			isDeleted := true

			isTimeout, err := RunWithTimeout(func() error {
				if err = c.Get(context.Background(), key, repo); err != nil {
					if k8serrors.IsNotFound(err) {
						isDeleted = true
						return nil
					}
					t.Fatal(err)
				}

				return errors.New("Not yet deleted")
			}, time.Second*30, time.Second*1)
			if err != nil || isTimeout {
				t.Fatalf("Snapshot repository stil exist: %s", err.Error())
			}
			assert.True(t, isDeleted)
			return nil
		},
	}
}
