/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/onsi/gomega/gexec"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	elkv1alpha1 "github.com/disaster37/operator-elk-extra/api/v1alpha1"
	"github.com/disaster37/operator-elk-extra/pkg/mocks"
	"github.com/disaster37/operator-sdk-extra/pkg/mock"
	//+kubebuilder:scaffold:imports
)

var testEnv *envtest.Environment

type ControllerTestSuite struct {
	suite.Suite
	k8sClient                client.Client
	cfg                      *rest.Config
	mockCtrl                 *gomock.Controller
	mockElasticsearchHandler *mocks.MockElasticsearchHandler
}

func TestControllerSuite(t *testing.T) {
	suite.Run(t, new(ControllerTestSuite))
}

func (t *ControllerTestSuite) SetupSuite() {

	t.mockCtrl = gomock.NewController(t.T())
	t.mockElasticsearchHandler = mocks.NewMockElasticsearchHandler(t.mockCtrl)

	logf.SetLogger(zap.New(zap.UseDevMode(true)))
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetFormatter(&logrus.TextFormatter{
		DisableQuote: true,
	})

	// Setup testenv
	testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{
			filepath.Join("..", "config", "crd", "bases"),
		},
		ErrorIfCRDPathMissing:    true,
		ControlPlaneStopTimeout:  120 * time.Second,
		ControlPlaneStartTimeout: 120 * time.Second,
	}
	cfg, err := testEnv.Start()
	if err != nil {
		panic(err)
	}
	t.cfg = cfg

	// Add CRD sheme
	err = scheme.AddToScheme(scheme.Scheme)
	if err != nil {
		panic(err)
	}
	err = elkv1alpha1.AddToScheme(scheme.Scheme)
	if err != nil {
		panic(err)
	}

	// Init k8smanager and k8sclient
	k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme: scheme.Scheme,
	})
	if err != nil {
		panic(err)
	}
	k8sClient := k8sManager.GetClient()
	t.k8sClient = k8sClient

	// Init controlles
	licenseReconciler := &LicenseReconciler{
		Client: k8sClient,
		Scheme: scheme.Scheme,
	}
	licenseReconciler.SetLogger(logrus.WithFields(logrus.Fields{
		"type": "licenseController",
	}))
	licenseReconciler.SetRecorder(k8sManager.GetEventRecorderFor("license-controller"))
	licenseReconciler.SetReconsiler(mock.NewMockReconciler(licenseReconciler, t.mockElasticsearchHandler))
	if err = licenseReconciler.SetupWithManager(k8sManager); err != nil {
		panic(err)
	}
	secretReconciler := &SecretReconciler{
		Client: k8sClient,
		Scheme: scheme.Scheme,
	}
	secretReconciler.SetLogger(logrus.WithFields(logrus.Fields{
		"type": "secretController",
	}))
	secretReconciler.SetRecorder(k8sManager.GetEventRecorderFor("secret-controller"))
	if err = secretReconciler.SetupWithManager(k8sManager); err != nil {
		panic(err)
	}
	ilmReconciler := &ElasticsearchILMReconciler{
		Client: k8sClient,
		Scheme: scheme.Scheme,
	}
	ilmReconciler.SetLogger(logrus.WithFields(logrus.Fields{
		"type": "ilmController",
	}))
	ilmReconciler.SetRecorder(k8sManager.GetEventRecorderFor("ilm-controller"))
	ilmReconciler.SetReconsiler(mock.NewMockReconciler(ilmReconciler, t.mockElasticsearchHandler))
	if err = ilmReconciler.SetupWithManager(k8sManager); err != nil {
		panic(err)
	}
	repositoryReconciler := &ElasticsearchSnapshotRepositoryReconciler{
		Client: k8sClient,
		Scheme: scheme.Scheme,
	}
	repositoryReconciler.SetLogger(logrus.WithFields(logrus.Fields{
		"type": "repositoryController",
	}))
	repositoryReconciler.SetRecorder(k8sManager.GetEventRecorderFor("repository-controller"))
	repositoryReconciler.SetReconsiler(mock.NewMockReconciler(repositoryReconciler, t.mockElasticsearchHandler))
	if err = repositoryReconciler.SetupWithManager(k8sManager); err != nil {
		panic(err)
	}

	go func() {
		err = k8sManager.Start(ctrl.SetupSignalHandler())
		if err != nil {
			panic(err)
		}
	}()
}

func (t *ControllerTestSuite) TearDownSuite() {
	gexec.KillAndWait(5 * time.Second)

	// Teardown the test environment once controller is fnished.
	// Otherwise from Kubernetes 1.21+, teardon timeouts waiting on
	// kube-apiserver to return
	err := testEnv.Stop()
	if err != nil {
		panic(err)
	}
}

func (t *ControllerTestSuite) BeforeTest(suiteName, testName string) {
	//t.mockCentreonService.EXPECT().SetLogger(gomock.Any()).AnyTimes().Return()
	// Init mock
}

func (t *ControllerTestSuite) AfterTest(suiteName, testName string) {
	defer t.mockCtrl.Finish()
}

func RunWithTimeout(f func() error, timeout time.Duration, interval time.Duration) (isTimeout bool, err error) {
	control := make(chan bool)
	timeoutTimer := time.NewTimer(timeout)
	go func() {
		loop := true
		intervalTimer := time.NewTimer(interval)
		for loop {
			select {
			case <-control:
				return
			case <-intervalTimer.C:
				err = f()
				if err != nil {
					intervalTimer.Reset(interval)
				} else {
					loop = false
				}
			}
		}
		control <- true
		return
	}()

	select {
	case <-control:
		return false, nil
	case <-timeoutTimer.C:
		control <- true
		return true, err
	}
}
