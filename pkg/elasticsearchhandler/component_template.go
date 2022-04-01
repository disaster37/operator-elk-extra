package elasticsearchhandler

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"

	olivere "github.com/olivere/elastic/v7"
	"github.com/pkg/errors"
)

// ComponentTemplateUpdate permit to update component template
func (h *ElasticsearchHandlerImpl) ComponentTemplateUpdate(name string, component *olivere.IndicesGetComponentTemplateData) (err error) {

	data, err := json.Marshal(component)
	if err != nil {
		return err
	}

	res, err := h.client.API.Cluster.PutComponentTemplate(
		name,
		bytes.NewReader(data),
		h.client.API.Cluster.PutComponentTemplate.WithContext(context.Background()),
		h.client.API.Cluster.PutComponentTemplate.WithPretty(),
	)

	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.IsError() {
		return errors.Errorf("Error when add index component template %s: %s", name, res.String())
	}

	return nil
}

// ComponentTemplateDelete permit to delete component template
func (h *ElasticsearchHandlerImpl) ComponentTemplateDelete(name string) (err error) {

	res, err := h.client.API.Cluster.DeleteComponentTemplate(
		name,
		h.client.API.Cluster.DeleteComponentTemplate.WithContext(context.Background()),
		h.client.API.Cluster.DeleteComponentTemplate.WithPretty(),
	)

	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.IsError() {
		if res.StatusCode == 404 {
			return nil
		}
		return errors.Errorf("Error when delete index component template %s: %s", name, res.String())

	}

	return nil

}

// ComponentTemplateGet permit to get component template
func (h *ElasticsearchHandlerImpl) ComponentTemplateGet(name string) (component *olivere.IndicesGetComponentTemplateData, err error) {

	res, err := h.client.API.Cluster.GetComponentTemplate(
		h.client.API.Cluster.GetComponentTemplate.WithName(name),
		h.client.API.Cluster.GetComponentTemplate.WithContext(context.Background()),
		h.client.API.Cluster.GetComponentTemplate.WithPretty(),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.IsError() {
		if res.StatusCode == 404 {
			return nil, nil
		}
		return nil, errors.Errorf("Error when get index component template %s: %s", name, res.String())

	}
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	indexComponentTemplateResp := &olivere.IndicesGetComponentTemplateResponse{}
	if err := json.Unmarshal(b, indexComponentTemplateResp); err != nil {
		return nil, err
	}

	if len(indexComponentTemplateResp.ComponentTemplates) == 0 {
		return nil, nil
	}

	return indexComponentTemplateResp.ComponentTemplates[0].ComponentTemplate.Template, nil
}

// ComponentTemplateDiff permit to check if 2 component template are the same
func (h *ElasticsearchHandlerImpl) ComponentTemplateDiff(actual, expected *olivere.IndicesGetComponentTemplateData) (diff string, err error) {
	return standartDiff(actual, expected, h.log, nil)
}
