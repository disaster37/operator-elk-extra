package elasticsearchhandler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/jarcoal/httpmock"
	olivere "github.com/olivere/elastic/v7"
	"github.com/stretchr/testify/assert"
)

var urlIndexTemplate = fmt.Sprintf("%s/_index_template/test", baseURL)

func (t *ElasticsearchHandlerTestSuite) TestIndexTemplateGet() {

	result := &olivere.IndicesGetIndexTemplateResponse{}
	template := &olivere.IndicesGetIndexTemplate{
		IndexPatterns: []string{"test-index-template"},
		Priority:      2,
		Template: &olivere.IndicesGetIndexTemplateData{
			Settings: map[string]any{
				"index.refresh_interval": "5s",
			},
		},
	}
	result.IndexTemplates = olivere.IndicesGetIndexTemplatesSlice{olivere.IndicesGetIndexTemplates{IndexTemplate: template}}

	httpmock.RegisterResponder("GET", urlIndexTemplate, func(req *http.Request) (*http.Response, error) {
		resp, err := httpmock.NewJsonResponse(200, result)
		if err != nil {
			panic(err)
		}
		SetHeaders(resp)
		return resp, nil
	})

	resp, err := t.esHandler.IndexTemplateGet("test")
	if err != nil {
		t.Fail(err.Error())
	}
	assert.Equal(t.T(), template, resp)

	// When error
	httpmock.RegisterResponder("GET", urlIndexTemplate, httpmock.NewErrorResponder(errors.New("fack error")))
	resp, err = t.esHandler.IndexTemplateGet("test")
	assert.Error(t.T(), err)
}

func (t *ElasticsearchHandlerTestSuite) TestIndexTemplateDelete() {

	httpmock.RegisterResponder("DELETE", urlIndexTemplate, func(req *http.Request) (*http.Response, error) {
		resp := httpmock.NewStringResponse(200, "")
		SetHeaders(resp)
		return resp, nil
	})

	err := t.esHandler.IndexTemplateDelete("test")
	if err != nil {
		t.Fail(err.Error())
	}

	// When error
	httpmock.RegisterResponder("DELETE", urlIndexTemplate, httpmock.NewErrorResponder(errors.New("fack error")))
	err = t.esHandler.IndexTemplateDelete("test")
	assert.Error(t.T(), err)
}

func (t *ElasticsearchHandlerTestSuite) TestIndexTemplateUpdate() {
	template := &olivere.IndicesGetIndexTemplate{
		IndexPatterns: []string{"test-index-template"},
		Priority:      2,
		Template: &olivere.IndicesGetIndexTemplateData{
			Settings: map[string]any{
				"index.refresh_interval": "5s",
			},
		},
	}

	httpmock.RegisterResponder("PUT", urlIndexTemplate, func(req *http.Request) (*http.Response, error) {
		resp := httpmock.NewStringResponse(200, "")
		SetHeaders(resp)
		return resp, nil
	})

	err := t.esHandler.IndexTemplateUpdate("test", template)
	if err != nil {
		t.Fail(err.Error())
	}

	// When error
	httpmock.RegisterResponder("PUT", urlIndexTemplate, httpmock.NewErrorResponder(errors.New("fack error")))
	err = t.esHandler.IndexTemplateUpdate("test", template)
	assert.Error(t.T(), err)
}

func (t *ElasticsearchHandlerTestSuite) TestIndexTemplateDiff() {
	var actual, expected *olivere.IndicesGetIndexTemplate

	expected = &olivere.IndicesGetIndexTemplate{
		IndexPatterns: []string{"test-index-template"},
		Priority:      2,
		Template: &olivere.IndicesGetIndexTemplateData{
			Settings: map[string]any{
				"index.refresh_interval": "5s",
			},
		},
	}

	// When template not exist yet
	actual = nil
	diff, err := t.esHandler.IndexTemplateDiff(actual, expected)
	if err != nil {
		t.Fail(err.Error())
	}
	assert.NotEmpty(t.T(), diff)

	// When template is the same
	actual = &olivere.IndicesGetIndexTemplate{
		IndexPatterns: []string{"test-index-template"},
		Priority:      2,
		Template: &olivere.IndicesGetIndexTemplateData{
			Settings: map[string]any{
				"index.refresh_interval": "5s",
			},
		},
	}
	diff, err = t.esHandler.IndexTemplateDiff(actual, expected)
	if err != nil {
		t.Fail(err.Error())
	}
	assert.Empty(t.T(), diff)

	// When template is not the same
	expected.Template = &olivere.IndicesGetIndexTemplateData{
		Mappings: map[string]any{
			"_source.enabled":           false,
			"properties.host_name.type": "keyword",
		},
	}
	diff, err = t.esHandler.IndexTemplateDiff(actual, expected)
	if err != nil {
		t.Fail(err.Error())
	}
	assert.NotEmpty(t.T(), diff)

}
