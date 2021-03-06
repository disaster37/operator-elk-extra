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

package main

import (
	"flag"
	"os"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	"k8s.io/client-go/dynamic"
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"github.com/sirupsen/logrus"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	elkv1alpha1 "github.com/disaster37/operator-elk-extra/api/v1alpha1"
	"github.com/disaster37/operator-elk-extra/controllers"
	"github.com/disaster37/operator-elk-extra/pkg/helpers"
	//+kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
	version  = "develop"
	commit   = ""
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(elkv1alpha1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme

	utilruntime.Must(core.AddToScheme(scheme))

}

func main() {
	var (
		metricsAddr          string
		enableLeaderElection bool
		probeAddr            string
	)
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))
	log := logrus.New()
	log.SetLevel(getLogrusLogLevel())
	log.SetFormatter(&logrus.TextFormatter{
		DisableQuote: true,
	})

	watchNamespace, err := getWatchNamespace()
	var namespace string
	var multiNamespacesCached cache.NewCacheFunc

	if err != nil {
		setupLog.Info("WATCH_NAMESPACES env variable not setted, the manager will watch and manage resources in all namespaces")
	} else {
		setupLog.Info("Manager look only resources on namespaces %s", watchNamespace)
		watchNamespaces := helpers.StringToSlice(watchNamespace, ",")
		if len(watchNamespaces) == 1 {
			namespace = watchNamespace
		} else {
			multiNamespacesCached = cache.MultiNamespacedCacheBuilder(watchNamespaces)
		}

	}

	printVersion(ctrl.Log, metricsAddr, probeAddr)
	log.Infof("monitoring-operator version: %s - %s", version, commit)

	cfg := ctrl.GetConfigOrDie()
	timeout, err := getKubeClientTimeout()
	if err != nil {
		setupLog.Error(err, "KUBE_CLIENT_TIMEOUT must be a valid duration: %s", err.Error())
		os.Exit(1)
	}
	cfg.Timeout = timeout

	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		Port:                   9443,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "9275a4fc.k8s.webcenter.fr",
		Namespace:              namespace,
		NewCache:               multiNamespacesCached,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	dinamicClient, err := dynamic.NewForConfig(cfg)
	if err != nil {
		setupLog.Error(err, "unable to init dinamic client")
		os.Exit(1)
	}

	// License controller
	licenseController := &controllers.LicenseReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}
	licenseController.SetLogger(log.WithFields(logrus.Fields{
		"type": "LicenseController",
	}))
	licenseController.SetRecorder(mgr.GetEventRecorderFor("license-controller"))
	licenseController.SetReconsiler(licenseController)
	licenseController.SetDinamicClient(dinamicClient)

	if err = licenseController.SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "License")
		os.Exit(1)
	}

	// Secret controller
	secretController := &controllers.SecretReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}
	secretController.SetLogger(log.WithFields(logrus.Fields{
		"type": "SecretController",
	}))
	secretController.SetRecorder(mgr.GetEventRecorderFor("secret-controller"))

	if err = secretController.SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Secret")
		os.Exit(1)
	}

	// ILM controller
	ilmController := &controllers.ElasticsearchILMReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}
	ilmController.SetLogger(log.WithFields(logrus.Fields{
		"type": "ILMController",
	}))
	ilmController.SetRecorder(mgr.GetEventRecorderFor("ilm-controller"))
	ilmController.SetReconsiler(ilmController)
	ilmController.SetDinamicClient(dinamicClient)

	if err = ilmController.SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "ILM")
		os.Exit(1)
	}

	// SLM controller
	slmController := &controllers.ElasticsearchSLMReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}
	slmController.SetLogger(log.WithFields(logrus.Fields{
		"type": "SLMController",
	}))
	slmController.SetRecorder(mgr.GetEventRecorderFor("slm-controller"))
	slmController.SetReconsiler(slmController)
	slmController.SetDinamicClient(dinamicClient)

	if err = slmController.SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "SLM")
		os.Exit(1)
	}

	// Snapshot repository controller
	repositoryController := &controllers.ElasticsearchSnapshotRepositoryReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}
	repositoryController.SetLogger(log.WithFields(logrus.Fields{
		"type": "RepositoryController",
	}))
	repositoryController.SetRecorder(mgr.GetEventRecorderFor("repository-controller"))
	repositoryController.SetReconsiler(repositoryController)
	repositoryController.SetDinamicClient(dinamicClient)
	if err = repositoryController.SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Repository")
		os.Exit(1)
	}

	// Component template controller
	componentTemplateController := &controllers.ElasticsearchComponentTemplateReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}
	componentTemplateController.SetLogger(log.WithFields(logrus.Fields{
		"type": "ComponentTemplateController",
	}))
	componentTemplateController.SetRecorder(mgr.GetEventRecorderFor("component-template-controller"))
	componentTemplateController.SetReconsiler(componentTemplateController)
	componentTemplateController.SetDinamicClient(dinamicClient)
	if err = componentTemplateController.SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "ComponentTemplate")
		os.Exit(1)
	}

	// Index template controller
	indexTemplateController := &controllers.ElasticsearchIndexTemplateReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}
	indexTemplateController.SetLogger(log.WithFields(logrus.Fields{
		"type": "IndexTemplateController",
	}))
	indexTemplateController.SetRecorder(mgr.GetEventRecorderFor("index-template-controller"))
	indexTemplateController.SetReconsiler(indexTemplateController)
	indexTemplateController.SetDinamicClient(dinamicClient)
	if err = indexTemplateController.SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "IndexTemplate")
		os.Exit(1)
	}

	// Elasticsearch role controller
	elasticsearchRoleController := &controllers.ElasticsearchRoleReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}
	elasticsearchRoleController.SetLogger(log.WithFields(logrus.Fields{
		"type": "ElasticsearchRoleController",
	}))
	elasticsearchRoleController.SetRecorder(mgr.GetEventRecorderFor("es-role-controller"))
	elasticsearchRoleController.SetReconsiler(elasticsearchRoleController)
	elasticsearchRoleController.SetDinamicClient(dinamicClient)
	if err = elasticsearchRoleController.SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "ElasticsearchRole")
		os.Exit(1)
	}

	// Role mapping controller
	roleMappingController := &controllers.RoleMappingReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}
	roleMappingController.SetLogger(log.WithFields(logrus.Fields{
		"type": "RoleMappingController",
	}))
	roleMappingController.SetRecorder(mgr.GetEventRecorderFor("role-mapping-controller"))
	roleMappingController.SetReconsiler(roleMappingController)
	roleMappingController.SetDinamicClient(dinamicClient)
	if err = roleMappingController.SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "RoleMapping")
		os.Exit(1)
	}

	// User controller
	userController := &controllers.UserReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}
	userController.SetLogger(log.WithFields(logrus.Fields{
		"type": "UserController",
	}))
	userController.SetRecorder(mgr.GetEventRecorderFor("user-controller"))
	userController.SetReconsiler(userController)
	userController.SetDinamicClient(dinamicClient)
	if err = userController.SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "User")
		os.Exit(1)
	}

	// Watch controller
	watchController := &controllers.ElasticsearchWatcherReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}
	watchController.SetLogger(log.WithFields(logrus.Fields{
		"type": "WatchController",
	}))
	watchController.SetRecorder(mgr.GetEventRecorderFor("watch-controller"))
	watchController.SetReconsiler(watchController)
	watchController.SetDinamicClient(dinamicClient)
	if err = watchController.SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Watch")
		os.Exit(1)
	}

	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
