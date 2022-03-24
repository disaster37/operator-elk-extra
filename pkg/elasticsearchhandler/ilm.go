package elasticsearchhandler

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"strings"

	"github.com/elastic/go-ucfg"
	"github.com/elastic/go-ucfg/diff"
	ucfgjson "github.com/elastic/go-ucfg/json"
	olivere "github.com/olivere/elastic/v7"
	"github.com/pkg/errors"
)

// ILMUpdate permit to update or create policy
func (h *ElasticsearchHandlerImpl) ILMUpdate(name, rawPolicy string) (err error) {

	h.log.Debugf("Name: %s", name)
	h.log.Debugf("Policy: %s", rawPolicy)

	res, err := h.client.API.ILM.PutLifecycle(
		name,
		h.client.API.ILM.PutLifecycle.WithContext(context.Background()),
		h.client.API.ILM.PutLifecycle.WithPretty(),
		h.client.API.ILM.PutLifecycle.WithBody(strings.NewReader(rawPolicy)),
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

	acualByte, err := json.Marshal(actual)
	if err != nil {
		return diffStr, err
	}
	expectedByte, err := json.Marshal(expected)
	if err != nil {
		return diffStr, err
	}

	actualConf, err := ucfgjson.NewConfig(acualByte, ucfg.PathSep("."))
	if err != nil {
		h.log.Errorf("Error when converting current Json: %s\ndata: %s", err.Error(), string(acualByte))
		return diffStr, err
	}
	expectedConf, err := ucfgjson.NewConfig(expectedByte, ucfg.PathSep("."))
	if err != nil {
		h.log.Errorf("Error when converting new Json: %s\ndata: %s", err.Error(), string(expectedByte))
		return diffStr, err
	}

	currentDiff := diff.CompareConfigs(actualConf, expectedConf)

	return currentDiff.GoStringer(), nil
}
