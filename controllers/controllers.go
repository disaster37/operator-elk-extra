package controllers

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	elkv1alpha1 "github.com/disaster37/operator-elk-extra/api/v1alpha1"
	es "github.com/disaster37/operator-elk-extra/pkg/elasticsearch"
	"github.com/disaster37/operator-elk-extra/pkg/elasticsearchhandler"
	"github.com/disaster37/operator-elk-extra/pkg/helpers"
	"github.com/disaster37/operator-sdk-extra/pkg/controller"
	elastic "github.com/elastic/go-elasticsearch/v8"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	core "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	waitDurationWhenError = 1 * time.Minute
	elasticBaseSecret     = "es-elastic-user"
	elasticBaseService    = "es-http"
	name                  = "elk.k8s.webcenter.fr"
)

type ElasticsearchReferer interface {
	GetElasticsearchRef() elkv1alpha1.ElasticsearchRefSpec
	IsManagedByECK() bool
}

type Reconciler struct {
	recorder      record.EventRecorder
	log           *logrus.Entry
	reconciler    controller.Reconciler
	dinamicClient dynamic.Interface
}

func (r *Reconciler) SetLogger(log *logrus.Entry) {
	r.log = log
}

func (r *Reconciler) SetRecorder(recorder record.EventRecorder) {
	r.recorder = recorder
}

func (r *Reconciler) SetReconsiler(reconciler controller.Reconciler) {
	r.reconciler = reconciler
}

func (r *Reconciler) SetDinamicClient(dc dynamic.Interface) {
	r.dinamicClient = dc
}

func GetElasticsearchHandler(ctx context.Context, resource ElasticsearchReferer, client client.Client, dinamicClient dynamic.Interface, req ctrl.Request, log *logrus.Entry) (esHandler elasticsearchhandler.ElasticsearchHandler, err error) {

	// Retrieve secret or elasticsearch resource that store the connexion credentials
	secretName := ""
	hosts := []string{}
	selfSignedCertificate := false
	if resource.IsManagedByECK() {
		// From Elasticsearch resource
		elasticsearch := &es.Elasticsearch{}
		u, err := dinamicClient.Resource(es.GVR).Namespace(req.NamespacedName.Namespace).Get(context.Background(), resource.GetElasticsearchRef().Name, meta.GetOptions{})
		if err != nil {
			if k8serrors.IsNotFound(err) {
				log.Warnf("Elasticsearch %s not yet exist, try later", resource.GetElasticsearchRef().Name)
				return nil, errors.Errorf("Elasticsearch %s not yet exist", resource.GetElasticsearchRef().Name)
			}
			log.Errorf("Error when get resource: %s", err.Error())
			return nil, err
		}
		if err = helpers.UnstructuredToStructured(u, elasticsearch); err != nil {
			return nil, err
		}

		// Get secret that store credential
		secretName = fmt.Sprintf("%s-%s", elasticsearch.Name, elasticBaseSecret)

		if elasticsearch.Spec.HTTP.TLS.SelfSignedCertificate.Disabled {
			hosts = append(hosts, fmt.Sprintf("http://%s-%s.%s:9200", elasticsearch.Name, elasticBaseService, elasticsearch.Namespace))
		} else {
			hosts = append(hosts, fmt.Sprintf("https://%s-%s.%s:9200", elasticsearch.Name, elasticBaseService, elasticsearch.Namespace))
			selfSignedCertificate = true
		}

	} else if len(resource.GetElasticsearchRef().Addresses) > 0 && resource.GetElasticsearchRef().SecretName != "" {
		secretName = resource.GetElasticsearchRef().SecretName
		hosts = resource.GetElasticsearchRef().Addresses
	} else {
		log.Error("You must set the way to connect on Elasticsearch")
		return nil, errors.New("You must set the way to connect on Elasticsearch")
	}

	// Read settings to access on Elasticsearch api
	secret := &core.Secret{}
	secretNS := types.NamespacedName{
		Namespace: req.NamespacedName.Namespace,
		Name:      secretName,
	}
	if err = client.Get(ctx, secretNS, secret); err != nil {
		if k8serrors.IsNotFound(err) {
			log.Warnf("Secret %s not yet exist, try later", secretName)
			return nil, errors.Errorf("Secret %s not yet exist", secretName)
		}
		log.Errorf("Error when get resource: %s", err.Error())
		return nil, err
	}

	transport := &http.Transport{
		Proxy:           http.ProxyFromEnvironment,
		TLSClientConfig: &tls.Config{},
	}
	cfg := elastic.Config{
		Transport: transport,
		Addresses: hosts,
	}
	for user, password := range secret.Data {
		cfg.Username = user
		cfg.Password = string(password)
		break
	}
	if selfSignedCertificate {
		transport.TLSClientConfig.InsecureSkipVerify = true
	}

	// Create Elasticsearch handler/client
	esHandler, err = elasticsearchhandler.NewElasticsearchHandler(cfg, log)
	if err != nil {
		return nil, err
	}

	return esHandler, nil
}
