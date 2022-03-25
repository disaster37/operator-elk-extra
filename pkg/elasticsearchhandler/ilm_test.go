package elasticsearchhandler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/jarcoal/httpmock"
	olivere "github.com/olivere/elastic/v7"
	"github.com/stretchr/testify/assert"
)

var urlILM = fmt.Sprintf("%s/_ilm/policy/test", baseURL)

func (t *ElasticsearchHandlerTestSuite) TestILMGet() {

	rawPolicy := `
{
	"policy": {
		"phases": {
			"warm": {
				"min_age": "10d",
				"actions": {
					"forcemerge": {
						"max_num_segments": 1
					}
				}
			},
			"delete": {
				"min_age": "31d",
				"actions": {
					"delete": {
						"delete_searchable_snapshot": true
					}
				}
			}
		}
	}
}
	`

	policyTest := map[string]interface{}{}
	if err := json.Unmarshal([]byte(rawPolicy), &policyTest); err != nil {
		panic(err)
	}

	// Normale use case
	result := &olivere.XPackIlmGetLifecycleResponse{
		Policy: policyTest,
	}

	httpmock.RegisterResponder("GET", urlILM, func(req *http.Request) (*http.Response, error) {
		resp, err := httpmock.NewJsonResponse(200, result)
		if err != nil {
			panic(err)
		}
		SetHeaders(resp)
		return resp, nil
	})

	policy, err := t.esHandler.ILMGet("test")
	if err != nil {
		t.Fail(err.Error())
	}
	assert.Equal(t.T(), policyTest, policy)

	// When error
	httpmock.RegisterResponder("GET", urlILM, httpmock.NewErrorResponder(errors.New("fack error")))
	policy, err = t.esHandler.ILMGet("test")
	assert.Error(t.T(), err)
}

func (t *ElasticsearchHandlerTestSuite) TestILMDelete() {

	httpmock.RegisterResponder("DELETE", urlILM, func(req *http.Request) (*http.Response, error) {
		resp := httpmock.NewStringResponse(200, "")
		SetHeaders(resp)
		return resp, nil
	})

	err := t.esHandler.ILMDelete("test")
	if err != nil {
		t.Fail(err.Error())
	}

	// When error
	httpmock.RegisterResponder("DELETE", urlILM, httpmock.NewErrorResponder(errors.New("fack error")))
	err = t.esHandler.ILMDelete("test")
	assert.Error(t.T(), err)
}

func (t *ElasticsearchHandlerTestSuite) TestILMUpdate() {

	rawPolicy := `
{
	"phases": {
		"warm": {
			"min_age": "10d",
			"actions": {
				"forcemerge": {
					"max_num_segments": 1
				}
			}
		},
		"delete": {
			"min_age": "31d",
			"actions": {
				"delete": {
					"delete_searchable_snapshot": true
				}
			}
		}
	}
}
	`

	httpmock.RegisterResponder("PUT", urlILM, func(req *http.Request) (*http.Response, error) {
		resp := httpmock.NewStringResponse(200, "")
		SetHeaders(resp)
		return resp, nil
	})

	err := t.esHandler.ILMUpdate("test", rawPolicy)
	if err != nil {
		t.Fail(err.Error())
	}

	// When error
	httpmock.RegisterResponder("PUT", urlILM, httpmock.NewErrorResponder(errors.New("fack error")))
	err = t.esHandler.ILMUpdate("test", rawPolicy)
	assert.Error(t.T(), err)
}

func (t *ElasticsearchHandlerTestSuite) TestILMDiff() {
	var actual, expected map[string]interface{}

	rawPolicy := `
{
	"policy": {
		"phases": {
			"warm": {
				"min_age": "10d",
				"actions": {
					"forcemerge": {
						"max_num_segments": 1
					}
				}
			},
			"delete": {
				"min_age": "31d",
				"actions": {
					"delete": {
						"delete_searchable_snapshot": true
					}
				}
			}
		}
	}
}
	`

	expected = map[string]interface{}{}
	if err := json.Unmarshal([]byte(rawPolicy), &expected); err != nil {
		panic(err)
	}

	// When ILM not exist yet
	actual = nil
	diff, err := t.esHandler.ILMDiff(actual, expected)
	if err != nil {
		t.Fail(err.Error())
	}
	assert.NotEmpty(t.T(), diff)

	// When policy is the same
	actual = map[string]interface{}{}
	if err := json.Unmarshal([]byte(rawPolicy), &actual); err != nil {
		panic(err)
	}
	diff, err = t.esHandler.ILMDiff(actual, expected)
	if err != nil {
		t.Fail(err.Error())
	}
	assert.Empty(t.T(), diff)

	// When policy is not the same
	rawPolicy = `
{
	"policy": {
		"phases": {
			"warm": {
				"min_age": "20d",
				"actions": {
					"forcemerge": {
						"max_num_segments": 1
					}
				}
			},
			"delete": {
				"min_age": "20d",
				"actions": {
					"delete": {
						"delete_searchable_snapshot": true
					}
				}
			}
		}
	}
}
	`
	expected = map[string]interface{}{}
	if err := json.Unmarshal([]byte(rawPolicy), &expected); err != nil {
		panic(err)
	}
	diff, err = t.esHandler.ILMDiff(actual, expected)
	if err != nil {
		t.Fail(err.Error())
	}
	assert.NotEmpty(t.T(), diff)

}
