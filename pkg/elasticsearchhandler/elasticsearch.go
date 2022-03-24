package elasticsearchhandler

import (
	elastic "github.com/elastic/go-elasticsearch/v8"
	olivere "github.com/olivere/elastic/v7"
	"github.com/sirupsen/logrus"
)

type ElasticsearchHandler interface {
	// License scope
	LicenseUpdate(license string) (err error)
	LicenseDelete() (err error)
	LicenseGet() (license *olivere.XPackInfoLicense, err error)
	LicenseDiff(actual, expected *olivere.XPackInfoLicense) (diff bool)
	LicenseEnableBasic() (err error)

	// ILM scope
	ILMUpdate(name, rawPolicy string) (err error)
	ILMDelete(name string) (err error)
	ILMGet(name string) (policy map[string]any, err error)
	ILMDiff(actual, expected map[string]any) (diff bool, explain string)

	SetLogger(log *logrus.Entry)
}

type ElasticsearchHandlerImpl struct {
	client *elastic.Client
	log    *logrus.Entry
}

func NewElasticsearchHandler(cfg elastic.Config, log *logrus.Entry) (ElasticsearchHandler, error) {

	client, err := elastic.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	return &ElasticsearchHandlerImpl{
		client: client,
		log:    log,
	}, nil
}

func (h *ElasticsearchHandlerImpl) SetLogger(log *logrus.Entry) {
	h.log = log
}
