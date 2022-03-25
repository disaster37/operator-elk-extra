package elasticsearchhandler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

var urlSLM = fmt.Sprintf("%s/_slm/policy/test", baseURL)

func (t *ElasticsearchHandlerTestSuite) TestSLMGet() {

	rawPolicy := `
{
	"policy": {
		"name": "<daily-snap-{now/d}>",
		"schedule": "0 30 1 * * ?",
		"repository": "repo",
		"config": {
			"indices": ["test-*"],
			"ignore_unavailable": false,
			"include_global_state": false
		},
		"retention": {
			"expire_after": "7d",
			"min_count": 5,
			"max_count": 10
		} 
	}
}
	`

	policyTest := &SnapshotLifecyclePolicyGet{}
	if err := json.Unmarshal([]byte(rawPolicy), policyTest); err != nil {
		panic(err)
	}

	// Normale use case
	result := map[string]any{
		"test": policyTest,
	}

	httpmock.RegisterResponder("GET", urlSLM, func(req *http.Request) (*http.Response, error) {
		resp, err := httpmock.NewJsonResponse(200, result)
		if err != nil {
			panic(err)
		}
		SetHeaders(resp)
		return resp, nil
	})

	policy, err := t.esHandler.SLMGet("test")
	if err != nil {
		t.Fail(err.Error())
	}
	assert.Equal(t.T(), policyTest.Policy, policy)

	// When error
	httpmock.RegisterResponder("GET", urlSLM, httpmock.NewErrorResponder(errors.New("fack error")))
	policy, err = t.esHandler.SLMGet("test")
	assert.Error(t.T(), err)
}

func (t *ElasticsearchHandlerTestSuite) TestSLMDelete() {

	httpmock.RegisterResponder("DELETE", urlSLM, func(req *http.Request) (*http.Response, error) {
		resp := httpmock.NewStringResponse(200, "")
		SetHeaders(resp)
		return resp, nil
	})

	err := t.esHandler.SLMDelete("test")
	if err != nil {
		t.Fail(err.Error())
	}

	// When error
	httpmock.RegisterResponder("DELETE", urlSLM, httpmock.NewErrorResponder(errors.New("fack error")))
	err = t.esHandler.SLMDelete("test")
	assert.Error(t.T(), err)
}

func (t *ElasticsearchHandlerTestSuite) TestSLMUpdate() {

	policy := &SnapshotLifecyclePolicySpec{
		Schedule:   "0 30 1 * * ?",
		Name:       "<daily-snap-{now/d}>",
		Repository: "repo",
		Configs: `{
			"indices": ["test-*"],
			"ignore_unavailable": false,
			"include_global_state": false
		}`,
		Retention: `{
			"expire_after": "7d",
			"min_count": 5,
			"max_count": 10
		}`,
	}

	httpmock.RegisterResponder("PUT", urlSLM, func(req *http.Request) (*http.Response, error) {
		resp := httpmock.NewStringResponse(200, "")
		SetHeaders(resp)
		return resp, nil
	})

	err := t.esHandler.SLMUpdate("test", policy)
	if err != nil {
		t.Fail(err.Error())
	}

	// When error
	httpmock.RegisterResponder("PUT", urlSLM, httpmock.NewErrorResponder(errors.New("fack error")))
	err = t.esHandler.SLMUpdate("test", policy)
	assert.Error(t.T(), err)
}

func (t *ElasticsearchHandlerTestSuite) TestSLMDiff() {
	var actual, expected *SnapshotLifecyclePolicySpec

	expected = &SnapshotLifecyclePolicySpec{
		Schedule:   "0 30 1 * * ?",
		Name:       "<daily-snap-{now/d}>",
		Repository: "repo",
		Configs: `{
			"indices": ["test-*"],
			"ignore_unavailable": false,
			"include_global_state": false
		}`,
		Retention: `{
			"expire_after": "7d",
			"min_count": 5,
			"max_count": 10
		}`,
	}

	// When SLM not exist yet
	actual = nil
	diff, err := t.esHandler.SLMDiff(actual, expected)
	if err != nil {
		t.Fail(err.Error())
	}
	assert.NotEmpty(t.T(), diff)

	// When policy is the same
	actual = &SnapshotLifecyclePolicySpec{
		Schedule:   "0 30 1 * * ?",
		Name:       "<daily-snap-{now/d}>",
		Repository: "repo",
		Configs: `{
			"indices": ["test-*"],
			"ignore_unavailable": false,
			"include_global_state": false
		}`,
		Retention: `{
			"expire_after": "7d",
			"min_count": 5,
			"max_count": 10
		}`,
	}
	diff, err = t.esHandler.SLMDiff(actual, expected)
	if err != nil {
		t.Fail(err.Error())
	}
	assert.Empty(t.T(), diff)

	// When policy is not the same
	expected.Repository = "repo2"
	diff, err = t.esHandler.SLMDiff(actual, expected)
	if err != nil {
		t.Fail(err.Error())
	}
	assert.NotEmpty(t.T(), diff)

}
