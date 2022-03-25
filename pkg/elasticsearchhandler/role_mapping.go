package elasticsearchhandler

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"

	olivere "github.com/olivere/elastic/v7"
	"github.com/pkg/errors"
)

// RoleMappingUpdate permit to create or update role mapping
func (h *ElasticsearchHandlerImpl) RoleMappingUpdate(name string, roleMapping *olivere.XPackSecurityRoleMapping) (err error) {

	data, err := json.Marshal(roleMapping)
	if err != nil {
		return err
	}

	res, err := h.client.API.Security.PutRoleMapping(
		name,
		bytes.NewReader(data),
		h.client.API.Security.PutRoleMapping.WithContext(context.Background()),
		h.client.API.Security.PutRoleMapping.WithPretty(),
	)

	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.IsError() {
		return errors.Errorf("Error when add role mapping %s: %s", name, res.String())
	}

	return nil
}

// RoleMappingDelete permit to delete role mapping
func (h *ElasticsearchHandlerImpl) RoleMappingDelete(name string) (err error) {

	res, err := h.client.API.Security.DeleteRoleMapping(
		name,
		h.client.API.Security.DeleteRoleMapping.WithContext(context.Background()),
		h.client.API.Security.DeleteRoleMapping.WithPretty(),
	)

	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.IsError() {
		if res.StatusCode == 404 {
			return nil
		}
		return errors.Errorf("Error when delete role mapping %s: %s", name, res.String())

	}

	h.log.Infof("Deleted role mapping %s successfully", name)

	return nil
}

// RoleMappingGet permit to get role mapping
func (h *ElasticsearchHandlerImpl) RoleMappingGet(name string) (roleMapping *olivere.XPackSecurityRoleMapping, err error) {

	res, err := h.client.API.Security.GetRoleMapping(
		h.client.API.Security.GetRoleMapping.WithContext(context.Background()),
		h.client.API.Security.GetRoleMapping.WithPretty(),
		h.client.API.Security.GetRoleMapping.WithName(name),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.IsError() {
		if res.StatusCode == 404 {
			return nil, nil
		}
		return nil, errors.Errorf("Error when get role mapping %s: %s", name, res.String())

	}
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	h.log.Debugf("Get role mapping %s successfully:\n%s", name, string(b))
	roleMappingResp := make(olivere.XPackSecurityGetRoleMappingResponse)
	err = json.Unmarshal(b, &roleMappingResp)
	if err != nil {
		return nil, err
	}

	h.log.Infof("Read role mapping %s successfully", name)

	tmp := roleMappingResp[name]
	return &tmp, nil
}

// RoleMappingDiff permit to check if 2 role mapping are the same
func (h *ElasticsearchHandlerImpl) RoleMappingDiff(actual, expected *olivere.XPackSecurityRoleMapping) (diff string, err error) {
	return standartDiff(&actual, &expected, h.log)
}
