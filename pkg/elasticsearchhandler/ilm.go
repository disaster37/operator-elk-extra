package elasticsearchhandler

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"

	olivere "github.com/olivere/elastic/v7"
	"github.com/pkg/errors"
)

// ILMUpdate permit to update or create policy
func (h *ElasticsearchHandlerImpl) ILMUpdate(name string, policy map[string]any) (err error) {

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
func (h *ElasticsearchHandlerImpl) ILMGet(name string) (policy map[string]any, err error) {

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

	policyResp := &olivere.XPackIlmGetLifecycleResponse{}
	err = json.Unmarshal(b, policyResp)
	if err != nil {
		return nil, err
	}

	return policyResp.Policy, nil

}

// ILMDiff permit to check if 2 policy are the same
func (h *ElasticsearchHandlerImpl) ILMDiff(actual, expected map[string]any) (diffStr string, err error) {
	return standartDiff(&actual, &expected, h.log)
}
