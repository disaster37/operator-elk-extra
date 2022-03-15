package elasticsearchhandler

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"testing"

	elastic "github.com/elastic/go-elasticsearch/v8"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
)

const baseURL = "http://localhost:9200"

type ElasticsearchHandlerTestSuite struct {
	suite.Suite
	esHandler ElasticsearchHandler
}

type MockTransport struct {
	Response    *http.Response
	RoundTripFn func(req *http.Request) (*http.Response, error)
}

func GetMockElasticsearchHandler(mock *MockTransport) ElasticsearchHandler {

	cfg := elastic.Config{
		Transport: mock,
	}

	esHandler, err := NewElasticsearchHandler(cfg, logrus.NewEntry(logrus.New()))
	if err != nil {
		panic(err)
	}

	return esHandler
}

func TestElasticsearchHandlerSuite(t *testing.T) {
	suite.Run(t, new(ElasticsearchHandlerTestSuite))
}

func (t *ElasticsearchHandlerTestSuite) SetupTest() {
}

func (t *ElasticsearchHandlerTestSuite) BeforeTest(suiteName, testName string) {
}

func (m *MockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.RoundTripFn(req)
}

func (m *MockTransport) SetResponse(status int, data interface{}, isError bool) {

	str, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}

	h := http.Header{}
	h.Set("X-Elastic-Product", "Elasticsearch")

	m.Response = &http.Response{
		StatusCode: status,
		Body:       ioutil.NopCloser(bytes.NewReader(str)),
		Header:     h,
	}
	if !isError {
		m.RoundTripFn = func(req *http.Request) (*http.Response, error) { return m.Response, nil }
	} else {
		m.RoundTripFn = func(req *http.Request) (*http.Response, error) { return m.Response, errors.New("Fake error") }
	}

}
