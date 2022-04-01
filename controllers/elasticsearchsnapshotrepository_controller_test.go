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

func (t *ControllerTestSuite) TestElasticsearchSnapshotRepositoryReconciler() {

	var (
		err       error
		isTimeout bool
		fetched   *elkv1alpha1.ElasticsearchSnapshotRepository
		test      string
		isCreated bool
		isDeleted bool
		isUpdated bool
	)

	repoName := "t-ilm-" + helpers.RandomString(10)
	key := types.NamespacedName{
		Name:      repoName,
		Namespace: "default",
	}
	t.mockElasticsearchHandler.EXPECT().SnapshotRepositoryGet(gomock.Any()).AnyTimes().DoAndReturn(func(name string) (*olivere.SnapshotRepositoryMetaData, error) {

		switch test {
		case "no_repo":
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
		case "repo_update":
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
	t.mockElasticsearchHandler.EXPECT().SnapshotRepositoryDiff(gomock.Any(), gomock.Any()).AnyTimes().DoAndReturn(func(actual, expected *olivere.SnapshotRepositoryMetaData) (string, error) {
		switch test {
		case "no_repo":
			if !isCreated {
				return "fake change", nil
			} else {
				return "", nil
			}
		case "repo_update":
			if !isUpdated {
				return "fake change", nil
			} else {
				return "", nil
			}
		}

		return "", nil

	})
	t.mockElasticsearchHandler.EXPECT().SnapshotRepositoryUpdate(gomock.Any(), gomock.Any()).AnyTimes().DoAndReturn(func(name string, policy *olivere.SnapshotRepositoryMetaData) error {
		switch test {
		case "no_repo":
			isCreated = true
			return nil
		case "repo_update":
			isUpdated = true
			return nil
		}

		return nil
	})
	t.mockElasticsearchHandler.EXPECT().SnapshotRepositoryDelete(gomock.Any()).AnyTimes().DoAndReturn(func(name string) error {
		switch test {
		case "repo_delete":
			isDeleted = true
			return nil
		}

		return nil
	})

	// When add new Snapshot repository
	logrus.Info("==================================== When add new snapshot repository")
	test = "no_repo"
	toCreate := &elkv1alpha1.ElasticsearchSnapshotRepository{
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
	if err = t.k8sClient.Create(context.Background(), toCreate); err != nil {
		t.T().Fatal(err)
	}

	isTimeout, err = RunWithTimeout(func() error {
		fetched = &elkv1alpha1.ElasticsearchSnapshotRepository{}
		if err := t.k8sClient.Get(context.Background(), key, fetched); err != nil {
			t.T().Fatal(err)
		}
		if !isCreated {
			return errors.New("Not yet created")
		}
		return nil
	}, time.Second*30, time.Second*1)
	if err != nil || isTimeout {
		t.T().Fatalf("Failed to get Snapshot repository: %s", err.Error())
	}
	assert.True(t.T(), condition.IsStatusConditionPresentAndEqual(fetched.Status.Conditions, repositoryCondition, metav1.ConditionTrue))
	time.Sleep(10 * time.Second)

	// When update snapshot repository
	logrus.Info("==================================== When update snapshot repository")
	test = "repo_update"
	isUpdated = false
	fetched = &elkv1alpha1.ElasticsearchSnapshotRepository{}
	if err := t.k8sClient.Get(context.Background(), key, fetched); err != nil {
		t.T().Fatal(err)
	}
	fetched.Spec.Settings = `
	{
		"url" : "http://fake2"
	}
	`
	if err = t.k8sClient.Update(context.Background(), fetched); err != nil {
		t.T().Fatal(err)
	}

	isTimeout, err = RunWithTimeout(func() error {
		fetched = &elkv1alpha1.ElasticsearchSnapshotRepository{}
		if err := t.k8sClient.Get(context.Background(), key, fetched); err != nil {
			t.T().Fatal(err)
		}
		if !isUpdated {
			return errors.New("Not yet updated")
		}
		return nil
	}, time.Second*30, time.Second*1)
	if err != nil || isTimeout {
		t.T().Fatalf("Failed to get Snapshot repository: %s", err.Error())
	}
	assert.True(t.T(), condition.IsStatusConditionPresentAndEqual(fetched.Status.Conditions, repositoryCondition, metav1.ConditionTrue))
	time.Sleep(10 * time.Second)

	// When remove snapshot repository
	logrus.Info("==================================== When remove snapshot repository")
	test = "repo_delete"
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
		t.T().Fatalf("Snapshot repository stil exist: %s", err.Error())
	}
	assert.True(t.T(), isDeleted)
	time.Sleep(10 * time.Second)

}
