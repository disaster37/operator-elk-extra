package elasticsearchhandler

import (
	"net/http"
	"testing"

	elastic "github.com/elastic/go-elasticsearch/v8"
	"github.com/jarcoal/httpmock"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
)

const baseURL = "http://localhost:9200"

type ElasticsearchHandlerTestSuite struct {
	suite.Suite
	esHandler ElasticsearchHandler
}

func TestElasticsearchHandlerSuite(t *testing.T) {
	suite.Run(t, new(ElasticsearchHandlerTestSuite))
}

func (t *ElasticsearchHandlerTestSuite) SetupTest() {

	cfg := elastic.Config{
		Addresses: []string{baseURL},
		Transport: httpmock.DefaultTransport,
	}
	client, err := elastic.NewClient(cfg)
	if err != nil {
		panic(err)
	}

	t.esHandler = &ElasticsearchHandlerImpl{
		client: client,
		log:    logrus.NewEntry(logrus.New()),
	}

	httpmock.Activate()
}

func (t *ElasticsearchHandlerTestSuite) BeforeTest(suiteName, testName string) {
	httpmock.Reset()
}

func SetHeaders(resp *http.Response) {
	resp.Header.Add("X-Elastic-Product", "Elasticsearch")
}
