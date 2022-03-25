package elasticsearchhandler

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"

	"github.com/pkg/errors"
)

// SnapshotLifecyclePolicy object returned by API
type SnapshotLifecyclePolicy map[string]*SnapshotLifecyclePolicyGet

// SnapshotLifecyclePolicySpec is the snapshot lifecycle policy object
type SnapshotLifecyclePolicySpec struct {
	Schedule   string      `json:"schedule"`
	Name       string      `json:"name"`
	Repository string      `json:"repository"`
	Configs    interface{} `json:"config,omitempty"`
	Retention  interface{} `json:"retention,omitempty"`
}

// SnapshotLifecyclePolicyGet is the policy
type SnapshotLifecyclePolicyGet struct {
	Policy *SnapshotLifecyclePolicySpec `json:"policy"`
}

// SLMUpdate permit to add or update SLM policy
func (h *ElasticsearchHandlerImpl) SLMUpdate(name string, policy *SnapshotLifecyclePolicySpec) (err error) {

	b, err := json.Marshal(policy)
	if err != nil {
		return err
	}

	res, err := h.client.API.SlmPutLifecycle(
		name,
		h.client.API.SlmPutLifecycle.WithBody(bytes.NewReader(b)),
		h.client.API.SlmPutLifecycle.WithContext(context.Background()),
		h.client.API.SlmPutLifecycle.WithPretty(),
	)

	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.IsError() {
		return errors.Errorf("Error when add snapshot lifecycle policy %s: %s", name, res.String())
	}

	return nil
}

// SLMDelete permit to delete SLM policy
func (h *ElasticsearchHandlerImpl) SLMDelete(name string) (err error) {

	res, err := h.client.API.SlmDeleteLifecycle(
		name,
		h.client.API.SlmDeleteLifecycle.WithContext(context.Background()),
		h.client.API.SlmDeleteLifecycle.WithPretty(),
	)

	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.IsError() {
		if res.StatusCode == 404 {
			return nil
		}
		return errors.Errorf("Error when delete snapshot lifecycle policy %s: %s", name, res.String())

	}

	return nil
}

// SLMGet permit to get SLM policy
func (h *ElasticsearchHandlerImpl) SLMGet(name string) (policy *SnapshotLifecyclePolicySpec, err error) {

	res, err := h.client.API.SlmGetLifecycle(
		h.client.API.SlmGetLifecycle.WithContext(context.Background()),
		h.client.API.SlmGetLifecycle.WithPretty(),
		h.client.API.SlmGetLifecycle.WithPolicyID(name),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.IsError() {
		if res.StatusCode == 404 {
			return nil, nil
		}
		return nil, errors.Errorf("Error when get snapshot lifecycle policy %s: %s", name, res.String())

	}
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	h.log.Debugf("Get snapshot lifecycle policy successfully:\n%s", string(b))

	slm := make(SnapshotLifecyclePolicy)
	err = json.Unmarshal(b, &slm)
	if err != nil {
		return nil, err
	}

	// Manage bug https://github.com/elastic/elasticsearch/issues/47664
	if len(slm) == 0 {
		return nil, nil
	}

	return slm[name].Policy, nil
}

// SLMDiff permit to check if 2 policy are the same
func (h *ElasticsearchHandlerImpl) SLMDiff(actual, expected *SnapshotLifecyclePolicySpec) (diffStr string, err error) {
	return standartDiff(&actual, &expected, h.log)
}
