package elasticsearchhandler

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"

	olivere "github.com/olivere/elastic/v7"
	"github.com/pkg/errors"
)

var ignoreILMPolicyDiff = map[string]any{
	"phases.delete.actions.delete.delete_searchable_snapshot": true,
}

// ILMUpdate permit to update or create policy
func (h *ElasticsearchHandlerImpl) ILMUpdate(name string, policy *olivere.XPackIlmGetLifecycleResponse) (err error) {

	b, err := json.Marshal(policy)
	if err != nil {
		return err
	}

	res, err := h.client.API.ILM.PutLifecycle(
		name,
		h.client.API.ILM.PutLifecycle.WithContext(context.Background()),
		h.client.API.ILM.PutLifecycle.WithPretty(),
		h.client.API.ILM.PutLifecycle.WithBody(bytes.NewReader(b)),
	)

	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.IsError() {
		return errors.Errorf("Error when add lifecycle policy %s: %s", name, res.String())
	}

	return nil
}

// ILMDelete permit to delete policy
func (h *ElasticsearchHandlerImpl) ILMDelete(name string) (err error) {

	h.log.Debugf("Name: %s", name)

	res, err := h.client.API.ILM.DeleteLifecycle(
		name,
		h.client.API.ILM.DeleteLifecycle.WithContext(context.Background()),
		h.client.API.ILM.DeleteLifecycle.WithPretty(),
	)

	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.IsError() {
		if res.StatusCode == 404 {
			return nil
		}
		return errors.Errorf("Error when delete lifecycle policy %s: %s", name, res.String())
	}

	return nil
}

// ILMGet permit to get policy
func (h *ElasticsearchHandlerImpl) ILMGet(name string) (policy *olivere.XPackIlmGetLifecycleResponse, err error) {

	h.log.Debugf("Name: %s", name)

	res, err := h.client.API.ILM.GetLifecycle(
		h.client.API.ILM.GetLifecycle.WithContext(context.Background()),
		h.client.API.ILM.GetLifecycle.WithPretty(),
		h.client.API.ILM.GetLifecycle.WithPolicy(name),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.IsError() {
		if res.StatusCode == 404 {
			return nil, nil
		}
		return nil, errors.Errorf("Error when get lifecycle policy %s: %s", name, res.String())
	}
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	h.log.Debugf("Get life cycle policy %s successfully:\n%s", name, string(b))

	policyResp := make(map[string]*olivere.XPackIlmGetLifecycleResponse)
	err = json.Unmarshal(b, &policyResp)
	if err != nil {
		return nil, err
	}

	return policyResp[name], nil

}

// ILMDiff permit to check if 2 policy are the same
func (h *ElasticsearchHandlerImpl) ILMDiff(actual, expected *olivere.XPackIlmGetLifecycleResponse) (diffStr string, err error) {
	var actualPolicy, expectedPolicy map[string]any

	if actual != nil {
		actualPolicy = actual.Policy
	}
	if expected != nil {
		expectedPolicy = expected.Policy
	}
	return standartDiff(actualPolicy, expectedPolicy, h.log, ignoreILMPolicyDiff)
}
