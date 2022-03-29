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

	core "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	condition "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func (t *ControllerTestSuite) TestLicenseReconciler() {

	var (
		err       error
		isTimeout bool
		fetched   *elkv1alpha1.License
		test      string
		isCreated bool
		isDeleted bool
		//isUpdated bool
	)

	licenseName := "t-license-" + helpers.RandomString(10)
	key := types.NamespacedName{
		Name:      licenseName,
		Namespace: "default",
	}
	t.mockElasticsearchHandler.EXPECT().LicenseGet().AnyTimes().DoAndReturn(func() (*olivere.XPackInfoLicense, error) {

		switch test {
		case "basic_license_without_license":
			if !isCreated {
				return nil, nil
			} else {
				return &olivere.XPackInfoLicense{
					UID:  "test",
					Type: "basic",
				}, nil
			}
		case "enterprise_license_from_basic_license":
			if !isCreated {
				return &olivere.XPackInfoLicense{
					UID:  "test",
					Type: "basic",
				}, nil
			} else {
				return &olivere.XPackInfoLicense{
					UID:  "test",
					Type: "gold",
				}, nil
			}
		}

		return nil, nil

	})
	t.mockElasticsearchHandler.EXPECT().LicenseDiff(gomock.Any(), gomock.Any()).AnyTimes().DoAndReturn(func(actual, expected *olivere.XPackInfoLicense) bool {
		switch test {
		case "basic_license_without_license":
			if !isCreated {
				return true
			} else {
				return false
			}
		case "enterprise_license_from_basic_license":
			if !isCreated {
				return true
			} else {
				return false
			}
		}

		return false

	})
	t.mockElasticsearchHandler.EXPECT().LicenseEnableBasic().AnyTimes().DoAndReturn(func() error {
		switch test {
		case "basic_license_without_license":
			if !isCreated {
				isCreated = true
				return nil
			} else {
				return nil
			}
		case "enterprise_license_from_basic_license":
			isDeleted = true
			return nil
		}

		return nil
	})
	t.mockElasticsearchHandler.EXPECT().LicenseUpdate(gomock.Any()).AnyTimes().DoAndReturn(func(license string) error {
		switch test {
		case "basic_license_without_license":
			return nil
		case "enterprise_license_from_basic_license":
			isCreated = true
			return nil
		}

		return nil
	})

	// When add basic license when no license already exist
	logrus.Info("==================================== When add basic license when no license already exist")
	test = "basic_license_without_license"
	toCreate := &elkv1alpha1.License{
		ObjectMeta: metav1.ObjectMeta{
			Name:      key.Name,
			Namespace: key.Namespace,
		},
		Spec: elkv1alpha1.LicenseSpec{
			ElasticsearchRefSpec: elkv1alpha1.ElasticsearchRefSpec{
				Name: "test",
			},
			SecretName: key.Name,
			Basic:      true,
		},
	}
	if err = t.k8sClient.Create(context.Background(), toCreate); err != nil {
		t.T().Fatal(err)
	}

	isTimeout, err = RunWithTimeout(func() error {
		fetched = &elkv1alpha1.License{}
		if err := t.k8sClient.Get(context.Background(), key, fetched); err != nil {
			t.T().Fatal(err)
		}
		if !condition.IsStatusConditionPresentAndEqual(fetched.Status.Conditions, licenseCondition, metav1.ConditionTrue) {
			return errors.New("Not yet created")
		}
		return nil
	}, time.Second*30, time.Second*1)
	if err != nil || isTimeout {
		t.T().Fatalf("Failed to get License: %s", err.Error())
	}
	assert.Empty(t.T(), fetched.Status.ExpireAt)
	assert.Empty(t.T(), fetched.Status.LicenseHash)
	assert.Equal(t.T(), "basic", fetched.Status.LicenseType)
	assert.True(t.T(), condition.IsStatusConditionPresentAndEqual(fetched.Status.Conditions, licenseCondition, metav1.ConditionTrue))
	time.Sleep(10 * time.Second)

	// When remove basic license
	logrus.Info("==================================== When remove basic license")
	wait := int64(0)
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
		t.T().Fatalf("License stil exist: %s", err.Error())
	}
	time.Sleep(10 * time.Second)

	// When add enterprise license when basic license already exist
	logrus.Info("==================================== When add enterprise license when basic license already exist")

	test = "enterprise_license_from_basic_license"
	isCreated = false
	licenseJson := `
	{
		"license": {
			"uid": "test",
			"type": "gold",
			"issue_date_in_millis": 1629849600000,
			"expiry_date_in_millis": 1661990399999,
			"max_nodes": 15,
			"issued_to": "test",
			"issuer": "API",
			"signature": "test",
			"start_date_in_millis": 1629849600000
		}
	}
	`
	secret := &core.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      key.Name,
			Namespace: key.Namespace,
		},
		Data: map[string][]byte{
			"license": []byte(licenseJson),
		},
	}
	toCreate = &elkv1alpha1.License{
		ObjectMeta: metav1.ObjectMeta{
			Name:      key.Name,
			Namespace: key.Namespace,
		},
		Spec: elkv1alpha1.LicenseSpec{
			ElasticsearchRefSpec: elkv1alpha1.ElasticsearchRefSpec{
				Name: "test",
			},
			SecretName: key.Name,
			Basic:      false,
		},
	}
	if err = t.k8sClient.Create(context.Background(), secret); err != nil {
		t.T().Fatal(err)
	}
	if err = t.k8sClient.Create(context.Background(), toCreate); err != nil {
		t.T().Fatal(err)
	}

	isTimeout, err = RunWithTimeout(func() error {
		fetched = &elkv1alpha1.License{}
		if err := t.k8sClient.Get(context.Background(), key, fetched); err != nil {
			t.T().Fatal(err)
		}
		if !condition.IsStatusConditionPresentAndEqual(fetched.Status.Conditions, licenseCondition, metav1.ConditionTrue) {
			return errors.New("Not yet created")
		}
		return nil
	}, time.Second*30, time.Second*1)
	if err != nil || isTimeout {
		t.T().Fatalf("Failed to get License: %s", err.Error())
	}
	assert.NotEmpty(t.T(), fetched.Status.ExpireAt)
	assert.NotEmpty(t.T(), fetched.Status.LicenseHash)
	assert.Equal(t.T(), "gold", fetched.Status.LicenseType)
	assert.True(t.T(), condition.IsStatusConditionPresentAndEqual(fetched.Status.Conditions, licenseCondition, metav1.ConditionTrue))

	// Check we add annotation on secret
	if err := t.k8sClient.Get(context.Background(), key, secret); err != nil {
		t.T().Fatal(err)
	}
	assert.Equal(t.T(), key.Name, secret.Annotations[licenseAnnotation])
	time.Sleep(10 * time.Second)

	// When remove enterprise license
	logrus.Info("==================================== When remove enterprise license")
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
		t.T().Fatalf("License stil exist: %s", err.Error())
	}
	assert.True(t.T(), isDeleted)

}

/*
func (t *ControllerTestSuite) TestLicenseReconcilerWithEnterpriseLicense() {

	var (
		err       error
		isTimeout bool
		fetched   *elkv1alpha1.License
	)

	licenseName := "t-license-" + helpers.RandomString(10)
	key := types.NamespacedName{
		Name:      licenseName,
		Namespace: "default",
	}

	gomock.InOrder(
		t.mockElasticsearchHandler.EXPECT().LicenseGet().Return(&olivere.XPackInfoLicense{
			UID:  "test",
			Type: "basic",
		}, nil),
		t.mockElasticsearchHandler.EXPECT().LicenseGet().AnyTimes().Return(&olivere.XPackInfoLicense{
			UID:  "test",
			Type: "gold",
		}, nil))
	gomock.InOrder(
		t.mockElasticsearchHandler.EXPECT().LicenseDiff(gomock.Any(), gomock.Any()).Return(true),
		t.mockElasticsearchHandler.EXPECT().LicenseDiff(gomock.Any(), gomock.Any()).AnyTimes().Return(true),
	)
	t.mockElasticsearchHandler.EXPECT().LicenseUpdate(gomock.Any()).Return(nil)



	toCreate := &elkv1alpha1.License{
		ObjectMeta: metav1.ObjectMeta{
			Name:      key.Name,
			Namespace: key.Namespace,
		},
		Spec: elkv1alpha1.LicenseSpec{
			ElasticsearchRefSpec: elkv1alpha1.ElasticsearchRefSpec{
				Name: "test",
			},
			SecretName: key.Name,
			Basic:      false,
		},
	}
	if err = t.k8sClient.Create(context.Background(), secret); err != nil {
		t.T().Fatal(err)
	}
	if err = t.k8sClient.Create(context.Background(), toCreate); err != nil {
		t.T().Fatal(err)
	}

	isTimeout, err = RunWithTimeout(func() error {
		fetched = &elkv1alpha1.License{}
		if err := t.k8sClient.Get(context.Background(), key, fetched); err != nil {
			t.T().Fatal(err)
		}
		if !condition.IsStatusConditionPresentAndEqual(fetched.Status.Conditions, licenseCondition, metav1.ConditionTrue) {
			return errors.New("Not yet created")
		}
		return nil
	}, time.Second*30, time.Second*1)
	if err != nil || isTimeout {
		t.T().Fatalf("Failed to get License: %s", err.Error())
	}
	assert.NotEmpty(t.T(), fetched.Status.ExpireAt)
	assert.NotEmpty(t.T(), fetched.Status.LicenseHash)
	assert.Equal(t.T(), "gold", fetched.Status.LicenseType)
	assert.True(t.T(), condition.IsStatusConditionPresentAndEqual(fetched.Status.Conditions, licenseCondition, metav1.ConditionTrue))
	time.Sleep(10 * time.Second)

}
*/
