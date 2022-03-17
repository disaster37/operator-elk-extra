package controllers

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"time"

	elkv1alpha1 "github.com/disaster37/operator-elk-extra/api/v1alpha1"
	"github.com/disaster37/operator-elk-extra/pkg/elasticsearchhandler"
	es "github.com/elastic/cloud-on-k8s/pkg/apis/elasticsearch/v1"
	elastic "github.com/elastic/go-elasticsearch/v8"
	"github.com/sirupsen/logrus"
	core "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	waitDurationWhenError = 1 * time.Minute
	elasticBaseSecret     = "-es-elastic-user"
	elasticBaseService    = "-es-http"
)

type ElasticsearchReferer interface {
	GetElasticsearchRef() elkv1alpha1.ElasticsearchRefSpec
	GetElasticsearchExternalRef() elkv1alpha1.ElasticsearchExternalRefSpec
}

func GetElasticsearchHandler(ctx context.Context, resource ElasticsearchReferer, client client.Client, req ctrl.Request, log *logrus.Entry) (esHandler elasticsearchhandler.ElasticsearchHandler, res *ctrl.Result, err error) {

	// Retrieve secret or elasticsearch resource that store the connexion credentials
	secretName := ""
	hosts := []string{}
	selfSignedCertificate := false
	if resource.GetElasticsearchRef().Name != "" {
		// From Elasticsearch resource
		elasticsearch := &es.Elasticsearch{}
		elasticsearchNs := types.NamespacedName{
			Namespace: req.NamespacedName.Namespace,
			Name:      resource.GetElasticsearchRef().Name,
		}
		if err := client.Get(ctx, elasticsearchNs, elasticsearch); err != nil {
			if k8serrors.IsNotFound(err) {
				log.Warnf("Elasticsearch %s not yet exist, try later", resource.GetElasticsearchRef().Name)
				return nil, &ctrl.Result{RequeueAfter: waitDurationWhenError}, nil
			}
			log.Errorf("Error when get resource: %s", err.Error())
			return nil, nil, err
		}

		// Get secret that store credential
		secretName = fmt.Sprintf("%s-%s", elasticsearch.Name, elasticBaseSecret)
		hosts = append(hosts, fmt.Sprintf("http://%s-%s.%s:9200", elasticsearch.Name, elasticBaseService, elasticsearch.Namespace))

	} else if resource.GetElasticsearchExternalRef().SecretName != "" {
		secretName = resource.GetElasticsearchExternalRef().SecretName
		hosts = resource.GetElasticsearchExternalRef().Addresses
	} else {
		log.Error("You must set the way to connect on Elasticsearch")
		return nil, nil, errors.New("You must set the way to connect on Elasticsearch")
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
			return nil, &ctrl.Result{RequeueAfter: waitDurationWhenError}, nil
		}
		log.Errorf("Error when get resource: %s", err.Error())
		return nil, nil, err
	}

	transport := &http.Transport{
		Proxy:           http.ProxyFromEnvironment,
		TLSClientConfig: &tls.Config{},
	}
	cfg := elastic.Config{
		Transport: transport,
		Addresses: hosts,
	}
	for user, passb64 := range secret.Data {
		cfg.Username = user

		password, err := base64.RawStdEncoding.DecodeString(string(passb64))
		if err != nil {
			return nil, nil, err
		}
		cfg.Password = string(password)
		break
	}
	if selfSignedCertificate {
		transport.TLSClientConfig.InsecureSkipVerify = true
	}

	// Create Elasticsearch handler/client
	esHandler, err = elasticsearchhandler.NewElasticsearchHandler(cfg, log)
	if err != nil {
		return nil, nil, err
	}

	return esHandler, nil, nil
}
