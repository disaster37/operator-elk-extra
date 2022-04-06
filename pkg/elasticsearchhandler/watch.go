package elasticsearchhandler

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"

	olivere "github.com/olivere/elastic/v7"
	"github.com/pkg/errors"
)

var ignoreWatchDiff = map[string]any{
	"status": nil,
}

// ILMUpdate permit to update or create policy
func (h *ElasticsearchHandlerImpl) WatchUpdate(name string, watch *olivere.XPackWatch) (err error) {

	b, err := json.Marshal(watch)
	if err != nil {
		return err
	}

	res, err := h.client.API.Watcher.PutWatch(
		name,
		h.client.API.Watcher.PutWatch.WithContext(context.Background()),
		h.client.API.Watcher.PutWatch.WithPretty(),
		h.client.API.Watcher.PutWatch.WithBody(bytes.NewReader(b)),
	)

	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.IsError() {
		return errors.Errorf("Error when add watch %s: %s", name, res.String())
	}

	return nil
}

// ILMDelete permit to delete policy
func (h *ElasticsearchHandlerImpl) WatchDelete(name string) (err error) {

	h.log.Debugf("Name: %s", name)

	res, err := h.client.API.Watcher.DeleteWatch(
		name,
		h.client.API.Watcher.DeleteWatch.WithContext(context.Background()),
		h.client.API.Watcher.DeleteWatch.WithPretty(),
	)

	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.IsError() {
		if res.StatusCode == 404 {
			return nil
		}
		return errors.Errorf("Error when delete watch %s: %s", name, res.String())
	}

	return nil
}

// ILMGet permit to get policy
func (h *ElasticsearchHandlerImpl) WatchGet(name string) (watch *olivere.XPackWatch, err error) {

	h.log.Debugf("Name: %s", name)

	res, err := h.client.API.Watcher.GetWatch(
		name,
		h.client.API.Watcher.GetWatch.WithContext(context.Background()),
		h.client.API.Watcher.GetWatch.WithPretty(),
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

	watchResp := &olivere.XPackWatcherGetWatchResponse{}
	err = json.Unmarshal(b, watchResp)
	if err != nil {
		return nil, err
	}

	return watchResp.Watch, nil

}

// ILMDiff permit to check if 2 policy are the same
func (h *ElasticsearchHandlerImpl) WatchDiff(actual, expected *olivere.XPackWatch) (diffStr string, err error) {
	return standartDiff(actual, expected, h.log, ignoreWatchDiff)
}
